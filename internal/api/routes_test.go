package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/76Parker/metrico/internal/config"
	"github.com/76Parker/metrico/internal/signature"
	"github.com/gin-gonic/gin"
)

func TestNewHTTPServer_WithHashKey(t *testing.T) {
	const key = "test-key"
	requestBody := []byte(`{"id":"Alloc"}`)
	responseBody := []byte(`{"ok":true}`)
	router := gin.New()
	router.POST("/updates", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json", responseBody)
	})
	server := newHttpServer(config.HTTP{HashKey: key}, router)
	request := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(requestBody))
	request.Header.Set(signature.HeaderName, signature.Sum(requestBody, key))
	response := httptest.NewRecorder()

	server.Handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if got := response.Header().Get(signature.HeaderName); got != signature.Sum(responseBody, key) {
		t.Fatalf("response hash = %q, want %q", got, signature.Sum(responseBody, key))
	}
}
