package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/taskflow/taskflow/internal/dto"
	"github.com/taskflow/taskflow/internal/middleware"
	"github.com/taskflow/taskflow/internal/model"
	"github.com/taskflow/taskflow/internal/service"
)

type TaskHandler struct{ svc service.TaskService }

func NewTaskHandler(svc service.TaskService) *TaskHandler { return &TaskHandler{svc: svc} }

func (h *TaskHandler) Create(c *gin.Context) {
	var req dto.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}
	uid := middleware.UserIDFrom(c)
	t, err := h.svc.Create(c.Request.Context(), uid, req)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, t)
}

func (h *TaskHandler) List(c *gin.Context) {
	uid := middleware.UserIDFrom(c)

	f := dto.TaskFilter{
		Status:   c.Query("status"),
		Priority: c.Query("priority"),
		Sort:     c.Query("sort"),
		Page:     queryInt(c, "page", 1),
		Limit:    queryInt(c, "limit", 20),
	}
	if raw := c.Query("project_id"); raw != "" {
		if id, err := strconv.ParseInt(raw, 10, 64); err == nil {
			f.ProjectID = &id
		}
	}
	if raw := c.Query("assignee_id"); raw != "" {
		if id, err := strconv.ParseInt(raw, 10, 64); err == nil {
			f.AssigneeID = &id
		}
	}

	items, total, err := h.svc.List(c.Request.Context(), uid, f)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ListResponse[model.Task]{
		Items:      items,
		Pagination: dto.Pagination{Page: f.Page, Limit: f.Limit, Total: total},
	})
}

func (h *TaskHandler) Get(c *gin.Context) {
	id, ok := param64(c, "id")
	if !ok {
		return
	}
	uid := middleware.UserIDFrom(c)
	t, err := h.svc.Get(c.Request.Context(), uid, id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, t)
}

func (h *TaskHandler) Update(c *gin.Context) {
	id, ok := param64(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}
	uid := middleware.UserIDFrom(c)
	t, err := h.svc.Update(c.Request.Context(), uid, id, req)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, t)
}

func (h *TaskHandler) Delete(c *gin.Context) {
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
