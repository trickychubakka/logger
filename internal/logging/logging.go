// Package logging -- пакет конфигурирования логирования.
package logging

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"time"
)

// WithLogging обертка над gin.HandlerFunc для внедрения Zap логирования.
func WithLogging(sugar *zap.SugaredLogger) gin.HandlerFunc {
	logFn := func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		sugar.Infoln(
			"uri", c.Request.RequestURI,
			"method", c.Request.Method,
			"status", c.Writer.Status(), // получаем перехваченный код статуса ответа
			"duration", duration,
			"size", c.Writer.Size(), // получаем перехваченный размер ответа
		)
	}
	return logFn
}
