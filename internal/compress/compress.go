// Package compress -- пакет с объектами, используемыми для HTTP сжатия.
package compress

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	//"logger/cmd/server/initconf"
	"logger/config"
	"net/http"
)

// const синонимы для степеней сжатия из пакета compress/gzip
const (
	BestCompression    = gzip.BestCompression
	BestSpeed          = gzip.BestSpeed
	DefaultCompression = gzip.DefaultCompression
	NoCompression      = gzip.NoCompression
)

// contentTypeToCompressMap -- map с вариантами Content-Type, для которых отрабатывает сжатие.
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

// checkSign функция проверки подписи.
func checkSign(body []byte, hash string, config *config.Config) (bool, error) {
	if config.Key == "" {
		return false, nil
	}
	var (
		data []byte // Декодированный hash с подписью
		err  error
		sign []byte // HMAC-подпись от идентификатора
	)

	data, err = hex.DecodeString(hash)
	if err != nil {
		log.Println("checkSign: hex.DecodeString error", err)
		return true, err
	}
	h := hmac.New(sha256.New, []byte(config.Key))
	h.Write(body)
	sign = h.Sum(nil)

	if hmac.Equal(sign, data) {
		fmt.Println("Подпись подлинна.")
		return true, nil
	} else {
		log.Println("Подпись неверна.")
		return true, fmt.Errorf("%s %v", "checkSign error: signature is incorrect ", err)
	}
}

// GzipRequestHandle хэндлер распаковки body входящего Request запроса.
func GzipRequestHandle(_ context.Context, config *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body []byte
		var newBody *bytes.Reader
		var gz *gzip.Reader
		var err, err1 error
		if c.Request.Header.Get(`Content-Encoding`) == `compress` {
			log.Println("c.Request.Header.Get(\"HashSHA256\") is :", c.Request.Header.Get("HashSHA256"))
			if hash := c.Request.Header.Get("HashSHA256"); hash != "" {
				body, err = io.ReadAll(c.Request.Body)
				if err != nil {
					log.Println("GzipRequestHandle: ioutil.ReadAll body error", err)
					c.Status(http.StatusInternalServerError)
					return
				}
				keyBool, err := checkSign(body, hash, config)
				if keyBool {
					if err != nil {
						log.Println("GzipRequestHandle: checkSign error", err)
						c.Set("GzipRequestHandle", "error")
						c.Status(http.StatusBadRequest)
						return
					}
				}
				// Ввиду того, что c.Request.Body был вычитан из body с помощью io.ReadAll -- делаем его копию для передачи в gz.

				newBody = bytes.NewReader(body)
				gz, err1 = gzip.NewReader(newBody)
				if err1 != nil {
					log.Println("Error in GzipRequestHandle1:", err1)
					c.Set("GzipRequestHandle", "error")
					c.Status(http.StatusInternalServerError)
					return
				}
			} else {
				gz, err = gzip.NewReader(c.Request.Body)
				if err != nil {
					log.Println("Error in GzipRequestHandle2:", err)
					c.Set("GzipRequestHandle", "error")
					c.Status(http.StatusInternalServerError)
					return
				}
			}
			log.Println("compress decompression")

			defer func(gz *gzip.Reader) {
				err1 := gz.Close()
				if err1 != nil {
					log.Println("gz.Close() error")
				}
			}(gz)
			c.Request.Body = gz
			c.Set("GzipRequestHandle", "success")
			c.Next()
		}
	}
}
