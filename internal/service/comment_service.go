package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/taskflow/taskflow/internal/dto"
	"github.com/taskflow/taskflow/internal/model"
	"github.com/taskflow/taskflow/internal/repository"
	"github.com/taskflow/taskflow/internal/utils"
	"github.com/taskflow/taskflow/internal/worker"
)

type commentService struct {
	comments repository.CommentRepository
	tasks    repository.TaskRepository
	projects repository.ProjectRepository
	bus      *worker.Pool
}

func NewCommentService(
	comments repository.CommentRepository,
	tasks repository.TaskRepository,
	projects repository.ProjectRepository,
	bus *worker.Pool,
) CommentService {
	return &commentService{comments: comments, tasks: tasks, projects: projects, bus: bus}
}

func (s *commentService) Add(ctx context.Context, authorID, taskID int64, req dto.CreateCommentRequest) (*model.Comment, error) {
	t, err := s.tasks.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if _, err := s.projects.MemberRole(ctx, t.ProjectID, authorID); err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return nil, utils.ErrForbidden
		}
		return nil, err
	}

	c := &model.Comment{TaskID: taskID, AuthorID: authorID, Content: req.Content}
	if _, err := s.comments.Create(ctx, c); err != nil {
		return nil, err
	}

	if t.AssigneeID != nil && *t.AssigneeID != authorID {
		s.bus.Publish(worker.Event{
			Type:      worker.EventCommentAdded,
			ProjectID: t.ProjectID,
			TaskID:    taskID,
			ActorID:   authorID,
			UserID:    *t.AssigneeID,
			Message:   fmt.Sprintf("New comment on: %s", t.Title),
		})
	}
	return c, nil
}

func (s *commentService) List(ctx context.Context, userID, taskID int64, page, limit int) ([]model.Comment, int64, error) {
	t, err := s.tasks.GetByID(ctx, taskID)
	if err != nil {
		return nil, 0, err
	}
	if _, err := s.projects.MemberRole(ctx, t.ProjectID, userID); err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return nil, 0, utils.ErrForbidden
		}
		return nil, 0, err
	}
	return s.comments.ListByTask(ctx, taskID, page, limit)
}
