package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/taskflow/taskflow/internal/middleware"
	"github.com/taskflow/taskflow/internal/service"
)

type StatsHandler struct{ svc service.StatsService }

func NewStatsHandler(svc service.StatsService) *StatsHandler { return &StatsHandler{svc: svc} }

func (h *StatsHandler) Project(c *gin.Context) {
	id, ok := param64(c, "id")
	if !ok {
		return
	}
	uid := middleware.UserIDFrom(c)
	stats, err := h.svc.Project(c.Request.Context(), uid, id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *StatsHandler) User(c *gin.Context) {
	id, ok := param64(c, "id")
	if !ok {
		return
	}
	uid := middleware.UserIDFrom(c)
	stats, err := h.svc.User(c.Request.Context(), uid, id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, stats)
}
