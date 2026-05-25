package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/taskflow/taskflow/internal/dto"
	"github.com/taskflow/taskflow/internal/middleware"
	"github.com/taskflow/taskflow/internal/model"
	"github.com/taskflow/taskflow/internal/service"
)

type CommentHandler struct{ svc service.CommentService }

func NewCommentHandler(svc service.CommentService) *CommentHandler {
	return &CommentHandler{svc: svc}
}

func (h *CommentHandler) Add(c *gin.Context) {
	taskID, ok := param64(c, "id")
	if !ok {
		return
	}
	var req dto.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}
	uid := middleware.UserIDFrom(c)
	cm, err := h.svc.Add(c.Request.Context(), uid, taskID, req)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, cm)
}

func (h *CommentHandler) List(c *gin.Context) {
	taskID, ok := param64(c, "id")
	if !ok {
		return
	}
	uid := middleware.UserIDFrom(c)
	page := queryInt(c, "page", 1)
	limit := queryInt(c, "limit", 20)
	items, total, err := h.svc.List(c.Request.Context(), uid, taskID, page, limit)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ListResponse[model.Comment]{
		Items:      items,
		Pagination: dto.Pagination{Page: page, Limit: limit, Total: total},
	})
}
