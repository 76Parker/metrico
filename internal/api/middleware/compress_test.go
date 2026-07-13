package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestCompressOnlySupportedContentTypes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		contentType string
		body        string
		compressed  bool
	}{
		{
			name:        "compresses JSON with charset",
			contentType: "application/json; charset=utf-8",
			body:        `{"status":"ok"}`,
			compressed:  true,
		},
		{
			name:        "compresses HTML with charset",
			contentType: "text/html; charset=utf-8",
			body:        "<h1>Hello</h1>",
			compressed:  true,
		},
		{
			name:        "does not compress plain text",
			contentType: "text/plain; charset=utf-8",
			body:        "Hello",
			compressed:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(Compress())
			router.GET("/", func(c *gin.Context) {
				c.Header("Content-Type", tt.contentType)
				c.String(http.StatusOK, tt.body)
			})

			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request.Header.Set("Accept-Encoding", "gzip")
			router.ServeHTTP(response, request)

			if !tt.compressed {
				require.Empty(t, response.Header().Get("Content-Encoding"))
				require.Equal(t, tt.body, response.Body.String())
				return
			}

			require.Equal(t, "gzip", response.Header().Get("Content-Encoding"))
			reader, err := gzip.NewReader(response.Body)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, reader.Close()) })

			body, err := io.ReadAll(reader)
			require.NoError(t, err)
			require.Equal(t, tt.body, strings.TrimSpace(string(body)))
		})
	}
}
