package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Logging(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		if raw != "" {
			path = path + "?" + raw
		}
		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("took", time.Since(start)),
			zap.String("client_ip", c.ClientIP()),
			zap.String("request_id", c.GetString(CtxRequestID)),
		}
		if uid, ok := c.Get(CtxUserID); ok {
			fields = append(fields, zap.Any("user_id", uid))
		}
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
			log.Error("http request", fields...)
			return
		}
		log.Info("http request", fields...)
	}
}
