package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/taskflow/taskflow/internal/dto"
)

func Recovery(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Error("handler panic",
					zap.Any("panic", rec),
					zap.String("request_id", c.GetString(CtxRequestID)),
					zap.ByteString("stack", debug.Stack()),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{
					Error: "internal_error",
				})
			}
		}()
		c.Next()
	}
}
