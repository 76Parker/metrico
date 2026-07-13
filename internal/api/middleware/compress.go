package middleware

import (
	"compress/gzip"
	"mime"

	"github.com/gin-gonic/gin"
)

type gzipCompressor struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func (g *gzipCompressor) Write(p []byte) (int, error) {
	if g.writer == nil {
		if !shouldCompress(g.Header().Get("Content-Type")) {
			return g.ResponseWriter.Write(p)
		}

		g.Header().Set("Content-Encoding", "gzip")
		g.Header().Del("Content-Length")

		writer, err := gzip.NewWriterLevel(g.ResponseWriter, gzip.BestSpeed)
		if err != nil {
			return 0, err
		}
		g.writer = writer
	}

	return g.writer.Write(p)
}

func (g *gzipCompressor) WriteString(s string) (int, error) {
	return g.Write([]byte(s))
}

func shouldCompress(contentType string) bool {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}

	return mediaType == "text/html" || mediaType == "application/json"
}

func Compress() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		if c.GetHeader("Accept-Encoding") != "gzip" {
			c.Next()
			return
		}

		writer := &gzipCompressor{
			ResponseWriter: c.Writer,
		}
		c.Writer = writer
		c.Next()

		if writer.writer != nil {
			_ = writer.writer.Close()
		}
	})
}
