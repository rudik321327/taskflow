package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/taskflow/taskflow/internal/dto"
	"github.com/taskflow/taskflow/internal/middleware"
	"github.com/taskflow/taskflow/internal/service"
)

type AuthHandler struct{ svc service.AuthService }

func NewAuthHandler(svc service.AuthService) *AuthHandler { return &AuthHandler{svc: svc} }

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}
	res, err := h.svc.Register(c.Request.Context(), req)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, res)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}
	res, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *AuthHandler) Me(c *gin.Context) {
	uid := middleware.UserIDFrom(c)
	u, err := h.svc.Me(c.Request.Context(), uid)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.UserView{ID: u.ID, Email: u.Email, Name: u.Name})
}
