package logging

import (
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"
)

func New(w io.Writer, version string) *slog.Logger {
	handler := slog.NewJSONHandler(w, &slog.HandlerOptions{
		ReplaceAttr: replaceAttr,
	})
	logger := slog.New(handler).With("service", "movie-log-api")
	if version != "" {
		logger = logger.With("version", version)
	}
	return logger
}

func replaceAttr(groups []string, a slog.Attr) slog.Attr {
	if len(groups) > 0 {
		return a
	}
	switch a.Key {
	case slog.TimeKey:
		return slog.Time(slog.TimeKey, a.Value.Time().UTC())
	case slog.LevelKey:
		return slog.String(slog.LevelKey, strings.ToLower(a.Value.String()))
	case slog.MessageKey:
		return slog.String("event", a.Value.String())
	}
	return a
}

func Error(c *gin.Context, event string, err error, args ...any) {
	logRequest(c, slog.LevelError, event, err, args...)
}

func Warn(c *gin.Context, event string, err error, args ...any) {
	logRequest(c, slog.LevelWarn, event, err, args...)
}

func logRequest(c *gin.Context, level slog.Level, event string, err error, args ...any) {
	attrs := []any{}
	if userID, ok := c.Get("userID"); ok {
		attrs = append(attrs, "userId", fmt.Sprint(userID))
	}
	attrs = append(attrs, "path", c.Request.URL.Path)
	if err != nil {
		attrs = append(attrs, "message", err.Error())
	}
	attrs = append(attrs, args...)
	slog.Log(c.Request.Context(), level, event, attrs...)
}
