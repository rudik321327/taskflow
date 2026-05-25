package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const headerRequestID = "X-Request-ID"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(headerRequestID)
		if rid == "" {
			rid = uuid.NewString()
		}
		c.Set(CtxRequestID, rid)
		c.Writer.Header().Set(headerRequestID, rid)
		c.Next()
	}
}
