package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/taskflow/taskflow/internal/auth"
	"github.com/taskflow/taskflow/internal/dto"
)

func JWT(issuer auth.Issuer) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "missing_token"})
			return
		}
		token := strings.TrimPrefix(header, "Bearer ")
		claims, err := issuer.Parse(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "invalid_token"})
			return
		}
		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxUserEmail, claims.Email)
		c.Next()
	}
}

func UserIDFrom(c *gin.Context) int64 {
	v, ok := c.Get(CtxUserID)
	if !ok {
		return 0
	}
	id, _ := v.(int64)
	return id
}
