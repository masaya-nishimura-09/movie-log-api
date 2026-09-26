package middleware

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBodyLimit(t *testing.T) {
	t.Run(
		"passes the request body to the next handler when it is within the limit",
		func(t *testing.T) {
			r := gin.New()

			var gotBody []byte
			var gotErr error

			r.Use(BodyLimit(10))
			r.POST("/", func(c *gin.Context) {
				gotBody, gotErr = io.ReadAll(c.Request.Body)
			})

			body := strings.Repeat("a", 10)
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if gotErr != nil {
				t.Errorf(
					"BodyLimit() read error = %v, want nil",
					gotErr,
				)
			}
			if string(gotBody) != body {
				t.Errorf(
					"BodyLimit() body = %v, want %v",
					string(gotBody), body,
				)
			}
		},
	)

	t.Run(
		"returns MaxBytesError when the request body exceeds the limit",
		func(t *testing.T) {
			r := gin.New()

			var gotErr error

			r.Use(BodyLimit(10))
			r.POST("/", func(c *gin.Context) {
				_, gotErr = io.ReadAll(c.Request.Body)
			})

			body := strings.Repeat("a", 11)
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			var maxBytesErr *http.MaxBytesError
			if !errors.As(gotErr, &maxBytesErr) {
				t.Errorf(
					"BodyLimit() read error = %v, want %T",
					gotErr, maxBytesErr,
				)
			}
		},
	)
}
