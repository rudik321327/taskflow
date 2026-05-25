package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/taskflow/taskflow/internal/auth"
	"github.com/taskflow/taskflow/internal/cache"
	"github.com/taskflow/taskflow/internal/config"
	"github.com/taskflow/taskflow/internal/middleware"
)

type Handlers struct {
	Auth    *AuthHandler
	Project *ProjectHandler
	Task    *TaskHandler
	Comment *CommentHandler
	Stats   *StatsHandler
}

func NewRouter(cfg *config.Config, log *zap.Logger, issuer auth.Issuer, c cache.Cache, h Handlers) *gin.Engine {
	if cfg.App.Env != "development" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()

	r.Use(middleware.RequestID())
	r.Use(middleware.Recovery(log))
	r.Use(middleware.Logging(log))
	r.Use(middleware.RateLimit(c, cfg.Limiter.RequestsPerMinute, log))

	r.GET("/healthz", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")
	{
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", h.Auth.Register)
			authGroup.POST("/login", h.Auth.Login)
		}

		authed := api.Group("")
		authed.Use(middleware.JWT(issuer))
		{
			authed.GET("/auth/me", h.Auth.Me)

			projects := authed.Group("/projects")
			{
				projects.POST("", h.Project.Create)
				projects.GET("", h.Project.List)
				projects.GET("/:id", h.Project.Get)
				projects.PUT("/:id", h.Project.Update)
				projects.DELETE("/:id", h.Project.Delete)
				projects.POST("/:id/members", h.Project.AddMember)
				projects.GET("/:id/members", h.Project.ListMembers)
			}

			tasks := authed.Group("/tasks")
			{
				tasks.POST("", h.Task.Create)
				tasks.GET("", h.Task.List)
				tasks.GET("/:id", h.Task.Get)
				tasks.PUT("/:id", h.Task.Update)
				tasks.DELETE("/:id", h.Task.Delete)
				tasks.POST("/:id/comments", h.Comment.Add)
				tasks.GET("/:id/comments", h.Comment.List)
			}

			stats := authed.Group("/stats")
			{
				stats.GET("/projects/:id", h.Stats.Project)
				stats.GET("/users/:id", h.Stats.User)
			}
		}
	}
	return r
}
