package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestClientKey(t *testing.T) {
	const secret = "internal-secret"
	const remoteIP = "192.0.2.1"

	tests := []struct {
		name    string
		secret  string
		headers map[string]string
		want    string
	}{
		{
			name:   "uses X-Real-IP when the internal secret matches",
			secret: secret,
			headers: map[string]string{
				"X-Internal-Secret": secret,
				"X-Real-IP":         "203.0.113.7",
			},
			want: "203.0.113.7",
		},
		{
			name:   "uses the remote address when the internal secret is wrong",
			secret: secret,
			headers: map[string]string{
				"X-Internal-Secret": "wrong",
				"X-Real-IP":         "203.0.113.7",
			},
			want: remoteIP,
		},
		{
			name:   "uses the remote address when the internal secret is missing",
			secret: secret,
			headers: map[string]string{
				"X-Real-IP": "203.0.113.7",
			},
			want: remoteIP,
		},
		{
			name:   "uses the remote address when X-Real-IP is not an IP",
			secret: secret,
			headers: map[string]string{
				"X-Internal-Secret": secret,
				"X-Real-IP":         "not-an-ip",
			},
			want: remoteIP,
		},
		{
			name:   "uses the remote address when no internal secret is configured",
			secret: "",
			headers: map[string]string{
				"X-Internal-Secret": "",
				"X-Real-IP":         "203.0.113.7",
			},
			want: remoteIP,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.RemoteIPHeaders = []string{"X-Forwarded-For"}
			if err := r.SetTrustedProxies([]string{"0.0.0.0/0"}); err != nil {
				t.Fatal(err)
			}

			var got string
			r.GET("/", func(c *gin.Context) {
				got = ClientKey(tt.secret)(c)
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = remoteIP + ":12345"
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			r.ServeHTTP(httptest.NewRecorder(), req)

			if got != tt.want {
				t.Errorf("ClientKey() = %q, want %q", got, tt.want)
			}
		})
	}
}
