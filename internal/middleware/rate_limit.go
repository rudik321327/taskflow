package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/taskflow/taskflow/internal/cache"
	"github.com/taskflow/taskflow/internal/dto"
)

func RateLimit(c cache.Cache, perMinute int, log *zap.Logger) gin.HandlerFunc {
	if perMinute <= 0 {
		return func(ctx *gin.Context) { ctx.Next() }
	}
	limit := int64(perMinute)
	return func(ctx *gin.Context) {
		key := fmt.Sprintf("rl:%s", ctx.ClientIP())
		count, err := c.Incr(ctx.Request.Context(), key, time.Minute)
		if err != nil {
			log.Warn("rate limiter unavailable", zap.Error(err))
			ctx.Next()
			return
		}
		ctx.Writer.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		if count > limit {
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, dto.ErrorResponse{
				Error:   "rate_limited",
				Message: "too many requests, slow down",
			})
			return
		}
		ctx.Next()
	}
}
