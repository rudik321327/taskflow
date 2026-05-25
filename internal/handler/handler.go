package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/taskflow/taskflow/internal/auth"
	"github.com/taskflow/taskflow/internal/cache"
	"github.com/taskflow/taskflow/internal/dto"
	"github.com/taskflow/taskflow/internal/utils"
)

func respondError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, utils.ErrValidation):
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
	case errors.Is(err, utils.ErrNotFound), errors.Is(err, cache.ErrCacheMiss):
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "not_found"})
	case errors.Is(err, utils.ErrAlreadyExists), errors.Is(err, utils.ErrConflict):
		c.JSON(http.StatusConflict, dto.ErrorResponse{Error: "conflict", Message: err.Error()})
	case errors.Is(err, utils.ErrUnauthorized), errors.Is(err, auth.ErrInvalidToken), errors.Is(err, auth.ErrExpiredToken):
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
	case errors.Is(err, utils.ErrInvalidCreds):
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "invalid_credentials"})
	case errors.Is(err, utils.ErrForbidden):
		c.JSON(http.StatusForbidden, dto.ErrorResponse{Error: "forbidden"})
	default:
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal_error"})
	}
}

func param64(c *gin.Context, name string) (int64, bool) {
	raw := c.Param(name)
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_id"})
		return 0, false
	}
	return id, true
}

func queryInt(c *gin.Context, key string, def int) int {
	raw := c.Query(key)
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return def
	}
	return n
}
