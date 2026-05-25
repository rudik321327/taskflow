package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/taskflow/taskflow/internal/dto"
	"github.com/taskflow/taskflow/internal/middleware"
	"github.com/taskflow/taskflow/internal/model"
	"github.com/taskflow/taskflow/internal/service"
)

type ProjectHandler struct{ svc service.ProjectService }

func NewProjectHandler(svc service.ProjectService) *ProjectHandler { return &ProjectHandler{svc: svc} }

func (h *ProjectHandler) Create(c *gin.Context) {
	var req dto.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}
	uid := middleware.UserIDFrom(c)
	p, err := h.svc.Create(c.Request.Context(), uid, req)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, p)
}

func (h *ProjectHandler) List(c *gin.Context) {
	uid := middleware.UserIDFrom(c)
	page := queryInt(c, "page", 1)
	limit := queryInt(c, "limit", 20)
	items, total, err := h.svc.List(c.Request.Context(), uid, page, limit)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ListResponse[model.Project]{
		Items:      items,
		Pagination: dto.Pagination{Page: page, Limit: limit, Total: total},
	})
}

func (h *ProjectHandler) Get(c *gin.Context) {
	id, ok := param64(c, "id")
	if !ok {
		return
	}
	uid := middleware.UserIDFrom(c)
	p, err := h.svc.Get(c.Request.Context(), uid, id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *ProjectHandler) Update(c *gin.Context) {
	id, ok := param64(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}
	uid := middleware.UserIDFrom(c)
	if err := h.svc.Update(c.Request.Context(), uid, id, req); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ProjectHandler) Delete(c *gin.Context) {
	id, ok := param64(c, "id")
	if !ok {
		return
	}
	uid := middleware.UserIDFrom(c)
	if err := h.svc.Delete(c.Request.Context(), uid, id); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ProjectHandler) AddMember(c *gin.Context) {
	id, ok := param64(c, "id")
	if !ok {
		return
	}
	var req dto.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}
	uid := middleware.UserIDFrom(c)
	if err := h.svc.AddMember(c.Request.Context(), uid, id, req); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusCreated)
}

func (h *ProjectHandler) ListMembers(c *gin.Context) {
	id, ok := param64(c, "id")
	if !ok {
		return
	}
	uid := middleware.UserIDFrom(c)
	members, err := h.svc.ListMembers(c.Request.Context(), uid, id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, members)
}
