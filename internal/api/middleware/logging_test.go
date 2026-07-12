package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/76Parker/metrico/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type testLogger struct {
	fields map[string]any
}

func (l *testLogger) Warn(string, ...any)  {}
func (l *testLogger) Error(string, ...any) {}
func (l *testLogger) Debug(string, ...any) {}

func (l *testLogger) Info(_ string, keysAndValues ...any) {
	l.fields = make(map[string]any, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues); i += 2 {
		l.fields[keysAndValues[i].(string)] = keysAndValues[i+1]
	}
}

func (l *testLogger) With(...any) logger.Logger      { return l }
func (l *testLogger) WithGroup(string) logger.Logger { return l }

func TestWithLoggingLogsResponseSize(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := &testLogger{}
	router := gin.New()
	router.Use(WithLogging(log))
	router.GET("/metrics", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	require.Equal(t, 2, log.fields["response_size"])
}
