package logging

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func decode(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, log = %q", err, buf.String())
	}
	return entry
}

func useLogger(t *testing.T, version string) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(New(&buf, version))
	t.Cleanup(func() { slog.SetDefault(prev) })
	return &buf
}

func newContext(target string) *gin.Context {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, target, nil)
	return c
}

func TestNew(t *testing.T) {
	t.Run(
		"writes one JSON line with the time in UTC, a lowercase level, the event, the service and the version",
		func(t *testing.T) {
			var buf bytes.Buffer
			New(&buf, "abc1234").Warn("something_happened")

			if n := bytes.Count(buf.Bytes(), []byte("\n")); n != 1 {
				t.Fatalf("New() lines = %d, want %d", n, 1)
			}

			entry := decode(t, &buf)
			got, err := time.Parse(time.RFC3339Nano, entry["time"].(string))
			if err != nil {
				t.Fatalf("New() time = %v, want RFC 3339", entry["time"])
			}
			if got.Location() != time.UTC {
				t.Errorf("New() time location = %v, want %v", got.Location(), time.UTC)
			}

			want := map[string]any{
				"level":   "warn",
				"event":   "something_happened",
				"service": "movie-log-api",
				"version": "abc1234",
			}
			for k, v := range want {
				if entry[k] != v {
					t.Errorf("New() %s = %v, want %v", k, entry[k], v)
				}
			}
			if _, ok := entry["msg"]; ok {
				t.Errorf("New() msg = %v, want no msg", entry["msg"])
			}
		},
	)

	t.Run(
		"leaves out the version when it is empty",
		func(t *testing.T) {
			var buf bytes.Buffer
			New(&buf, "").Error("something_failed")

			entry := decode(t, &buf)
			if _, ok := entry["version"]; ok {
				t.Errorf("New() version = %v, want no version", entry["version"])
			}
		},
	)
}

func TestError(t *testing.T) {
	t.Run(
		"logs the user ID, the path without the query and the error message",
		func(t *testing.T) {
			buf := useLogger(t, "")
			c := newContext("/records/12?title=secret")
			c.Set("userID", uint(3))

			Error(c, "request_failed", errors.New("get record: connection refused"))

			entry := decode(t, buf)
			want := map[string]any{
				"level":   "error",
				"event":   "request_failed",
				"userId":  "3",
				"path":    "/records/12",
				"message": "get record: connection refused",
			}
			for k, v := range want {
				if entry[k] != v {
					t.Errorf("Error() %s = %v, want %v", k, entry[k], v)
				}
			}
		},
	)

	t.Run(
		"adds the extra fields after the error message",
		func(t *testing.T) {
			buf := useLogger(t, "")
			c := newContext("/records/")

			Error(c, "panic_recovered", errors.New("boom"), "stack", "goroutine 1")

			entry := decode(t, buf)
			if entry["stack"] != "goroutine 1" {
				t.Errorf("Error() stack = %v, want %v", entry["stack"], "goroutine 1")
			}
		},
	)
}

func TestWarn(t *testing.T) {
	t.Run(
		"leaves out the user ID and the message when they are absent",
		func(t *testing.T) {
			buf := useLogger(t, "")
			c := newContext("/auth/login")

			Warn(c, "rate_limited", nil)

			entry := decode(t, buf)
			if entry["level"] != "warn" {
				t.Errorf("Warn() level = %v, want %v", entry["level"], "warn")
			}
			if entry["path"] != "/auth/login" {
				t.Errorf("Warn() path = %v, want %v", entry["path"], "/auth/login")
			}
			for _, k := range []string{"userId", "message"} {
				if _, ok := entry[k]; ok {
					t.Errorf("Warn() %s = %v, want no %s", k, entry[k], k)
				}
			}
		},
	)
}
