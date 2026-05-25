package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/taskflow/taskflow/internal/dto"
	"github.com/taskflow/taskflow/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, u *model.User) (int64, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByID(ctx context.Context, id int64) (*model.User, error)
}

type ProjectRepository interface {
	CreateWithOwner(ctx context.Context, p *model.Project) (int64, error)
	GetByID(ctx context.Context, id int64) (*model.Project, error)
	ListForUser(ctx context.Context, userID int64, page, limit int) ([]model.Project, int64, error)
	Update(ctx context.Context, id int64, name, description *string) error
	Delete(ctx context.Context, id int64) error

	AddMember(ctx context.Context, projectID, userID int64, role model.ProjectRole) error
	ListMembers(ctx context.Context, projectID int64) ([]model.ProjectMember, error)
	MemberRole(ctx context.Context, projectID, userID int64) (model.ProjectRole, error)
}

type TaskRepository interface {
	Create(ctx context.Context, t *model.Task) (int64, error)
	GetByID(ctx context.Context, id int64) (*model.Task, error)
	List(ctx context.Context, f dto.TaskFilter) ([]model.Task, int64, error)
	Update(ctx context.Context, id int64, patch TaskPatch) error
	Delete(ctx context.Context, id int64) error
}

type TaskPatch struct {
	Title       *string
	Description *string
	AssigneeID  **int64
	Status      *model.TaskStatus
	Priority    *model.TaskPriority
	DueDate     **time.Time
}

type CommentRepository interface {
	Create(ctx context.Context, c *model.Comment) (int64, error)
	ListByTask(ctx context.Context, taskID int64, page, limit int) ([]model.Comment, int64, error)
}

type NotificationRepository interface {
	Create(ctx context.Context, n *model.Notification) (int64, error)
	ListByUser(ctx context.Context, userID int64, page, limit int) ([]model.Notification, int64, error)
}

type StatsRepository interface {
	ProjectStats(ctx context.Context, projectID int64) (*model.ProjectStats, error)
	UserStats(ctx context.Context, userID int64) (*model.UserStats, error)
}

func IsNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

type Pool = pgxpool.Pool
