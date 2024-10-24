package gzip

import (
	"compress/gzip"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	BestCompression    = gzip.BestCompression
	BestSpeed          = gzip.BestSpeed
	DefaultCompression = gzip.DefaultCompression
	NoCompression      = gzip.NoCompression
)

var contentTypeToCompressMap = map[string]bool{
	"text/html":                true,
	"text/html; charset=utf-8": true,
	"application/json":         true,
}

type gzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func (g gzipWriter) Write(data []byte) (int, error) {
	g.Header().Del("Content-Length")
	return g.writer.Write(data)
}

func (g *gzipWriter) WriteString(s string) (int, error) {
	g.Header().Del("Content-Length")
	return g.writer.Write([]byte(s))
}

func GzipResponseHandle(level int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// read
		t1 := time.Now()
		if !shouldCompress(c.Request) {
			// если gzip не поддерживается, передаём управление
			// дальше без изменений
			c.Next()
			//time.Sleep(500 * time.Millisecond)
			return
		}
		//// создаём gzip.Writer поверх текущего c.Writer
		gz, err := gzip.NewWriterLevel(c.Writer, level)
		if err != nil {
			io.WriteString(c.Writer, err.Error())
			return
		}
		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")
		c.Writer = &gzipWriter{c.Writer, gz}
		defer func() {
			c.Header("Content-Length", "0")
			gz.Close()
		}()
		c.Next()
		log.Println("********TIMING of GzipResponseHandle is:***************", time.Now().Sub(t1))
	}
}

func shouldCompress(req *http.Request) bool {
	if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
		log.Println("There is no Accept-Encoding.")
		return false
	}

	// Если Content-Type запроса содержится в contentTypeToCompressMap -- включается сжатие
	if contentTypeToCompressMap[req.Header.Get("content-type")] {
		log.Println("gzip compression for Content-Type", req.Header.Get("content-type"), "enabled")
		return true
		//return false
	}

	if contentTypeToCompressMap[req.Header.Get("Accept")] {
		log.Println("gzip compression for Content-Type", req.Header.Get("content-type"), "enabled")
		return true
		//return false
	}

	log.Println("Default - do not encode. Content-Type is", req.Header.Get("Content-Type"))
	return false
}

func GzipRequestHandle(c *gin.Context) {
	if c.Request.Header.Get(`Content-Encoding`) == `gzip` {
		log.Println("gzip decompression")
		gz, err := gzip.NewReader(c.Request.Body)
		//log.Println("gz is:", gz, "error is:", err)
		if err != nil {
			log.Println("Error in GzipRequestHandle:", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		defer gz.Close()
		c.Request.Body = gz
		//log.Println("c.Request.Body from GzipRequestHandle:", c.Request.Body)
		c.Next()
	}
}
