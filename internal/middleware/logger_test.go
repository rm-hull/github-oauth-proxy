package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestLoggerMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Logs standard paths", func(t *testing.T) {
		buf := &bytes.Buffer{}
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		log := zerolog.New(buf)
		zerolog.DefaultContextLogger = &log

		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Request = c.Request.WithContext(log.WithContext(c.Request.Context()))
			c.Next()
		})
		r.Use(Logger("/healthz"))
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, buf.String(), "\"method\":\"GET\"")
		assert.Contains(t, buf.String(), "\"path\":\"/test\"")
		assert.Contains(t, buf.String(), "\"latency_ms\"")
	})

	t.Run("Excludes paths", func(t *testing.T) {
		buf := &bytes.Buffer{}
		log := zerolog.New(buf)
		zerolog.DefaultContextLogger = &log

		r := gin.New()
		r.Use(Logger("/healthz"))
		r.GET("/healthz", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest("GET", "/healthz", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, buf.String(), "Log output should be empty for excluded path")
	})
}
