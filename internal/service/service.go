package service

import (
	"context"

	"github.com/taskflow/taskflow/internal/dto"
	"github.com/taskflow/taskflow/internal/model"
)

type AuthService interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error)
	Me(ctx context.Context, userID int64) (*model.User, error)
}

type ProjectService interface {
	Create(ctx context.Context, ownerID int64, req dto.CreateProjectRequest) (*model.Project, error)
	Get(ctx context.Context, userID, projectID int64) (*model.Project, error)
	List(ctx context.Context, userID int64, page, limit int) ([]model.Project, int64, error)
	Update(ctx context.Context, userID, projectID int64, req dto.UpdateProjectRequest) error
	Delete(ctx context.Context, userID, projectID int64) error
	AddMember(ctx context.Context, actorID, projectID int64, req dto.AddMemberRequest) error
	ListMembers(ctx context.Context, userID, projectID int64) ([]model.ProjectMember, error)
}

type TaskService interface {
	Create(ctx context.Context, creatorID int64, req dto.CreateTaskRequest) (*model.Task, error)
	Get(ctx context.Context, userID, taskID int64) (*model.Task, error)
	List(ctx context.Context, userID int64, f dto.TaskFilter) ([]model.Task, int64, error)
	Update(ctx context.Context, userID, taskID int64, req dto.UpdateTaskRequest) (*model.Task, error)
	Delete(ctx context.Context, userID, taskID int64) error
}

type CommentService interface {
	Add(ctx context.Context, authorID, taskID int64, req dto.CreateCommentRequest) (*model.Comment, error)
	List(ctx context.Context, userID, taskID int64, page, limit int) ([]model.Comment, int64, error)
}

type StatsService interface {
	Project(ctx context.Context, userID, projectID int64) (*model.ProjectStats, error)
	User(ctx context.Context, callerID, targetID int64) (*model.UserStats, error)
}
