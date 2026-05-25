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

type taskService struct {
	tasks    repository.TaskRepository
	projects repository.ProjectRepository
	bus      *worker.Pool
}

func NewTaskService(tasks repository.TaskRepository, projects repository.ProjectRepository, bus *worker.Pool) TaskService {
	return &taskService{tasks: tasks, projects: projects, bus: bus}
}

func (s *taskService) Create(ctx context.Context, creatorID int64, req dto.CreateTaskRequest) (*model.Task, error) {
	if err := s.requireProjectMember(ctx, req.ProjectID, creatorID); err != nil {
		return nil, err
	}
	priority := model.PriorityMedium
	if req.Priority != "" {
		priority = model.TaskPriority(req.Priority)
	}
	t := &model.Task{
		ProjectID:   req.ProjectID,
		CreatorID:   creatorID,
		AssigneeID:  req.AssigneeID,
		Title:       req.Title,
		Description: req.Description,
		Status:      model.TaskStatusTodo,
		Priority:    priority,
		DueDate:     req.DueDate,
	}
	if _, err := s.tasks.Create(ctx, t); err != nil {
		return nil, err
	}

	s.bus.Publish(worker.Event{
		Type:      worker.EventTaskCreated,
		ProjectID: t.ProjectID,
		TaskID:    t.ID,
		ActorID:   creatorID,
		UserID:    deref(t.AssigneeID, creatorID),
		Message:   fmt.Sprintf("Task created: %s", t.Title),
	})
	if t.AssigneeID != nil && *t.AssigneeID != creatorID {
		s.bus.Publish(worker.Event{
			Type:      worker.EventTaskAssigned,
			ProjectID: t.ProjectID,
			TaskID:    t.ID,
			ActorID:   creatorID,
			UserID:    *t.AssigneeID,
			Message:   fmt.Sprintf("Assigned to you: %s", t.Title),
		})
	}
	return t, nil
}

func (s *taskService) Get(ctx context.Context, userID, taskID int64) (*model.Task, error) {
	t, err := s.tasks.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if err := s.requireProjectMember(ctx, t.ProjectID, userID); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *taskService) List(ctx context.Context, userID int64, f dto.TaskFilter) ([]model.Task, int64, error) {
	if f.ProjectID != nil {
		if err := s.requireProjectMember(ctx, *f.ProjectID, userID); err != nil {
			return nil, 0, err
		}
	} else {
		uid := userID
		f.AssigneeID = &uid
	}
	return s.tasks.List(ctx, f)
}

func (s *taskService) Update(ctx context.Context, userID, taskID int64, req dto.UpdateTaskRequest) (*model.Task, error) {
	existing, err := s.tasks.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if err := s.requireProjectMember(ctx, existing.ProjectID, userID); err != nil {
		return nil, err
	}

	patch := repository.TaskPatch{}
	if req.Title != nil {
		patch.Title = req.Title
	}
	if req.Description != nil {
		patch.Description = req.Description
	}
	if req.AssigneeID != nil {
		v := req.AssigneeID
		patch.AssigneeID = &v
	}
	if req.Status != nil {
		st := model.TaskStatus(*req.Status)
		patch.Status = &st
	}
	if req.Priority != nil {
		pr := model.TaskPriority(*req.Priority)
		patch.Priority = &pr
	}
	if req.DueDate != nil {
		v := req.DueDate
		patch.DueDate = &v
	}

	if err := s.tasks.Update(ctx, taskID, patch); err != nil {
		return nil, err
	}
	updated, err := s.tasks.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	if patch.AssigneeID != nil && !sameInt64Ptr(existing.AssigneeID, updated.AssigneeID) && updated.AssigneeID != nil {
		s.bus.Publish(worker.Event{
			Type:      worker.EventTaskAssigned,
			ProjectID: updated.ProjectID,
			TaskID:    updated.ID,
			ActorID:   userID,
			UserID:    *updated.AssigneeID,
			Message:   fmt.Sprintf("Assigned to you: %s", updated.Title),
		})
	}
	s.bus.Publish(worker.Event{
		Type:      worker.EventTaskUpdated,
		ProjectID: updated.ProjectID,
		TaskID:    updated.ID,
		ActorID:   userID,
		UserID:    deref(updated.AssigneeID, updated.CreatorID),
		Message:   fmt.Sprintf("Task updated: %s", updated.Title),
	})
	return updated, nil
}

func (s *taskService) Delete(ctx context.Context, userID, taskID int64) error {
	t, err := s.tasks.GetByID(ctx, taskID)
	if err != nil {
		return err
	}
	if err := s.requireProjectMember(ctx, t.ProjectID, userID); err != nil {
		return err
	}
	return s.tasks.Delete(ctx, taskID)
}

func (s *taskService) requireProjectMember(ctx context.Context, projectID, userID int64) error {
	if _, err := s.projects.MemberRole(ctx, projectID, userID); err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return utils.ErrForbidden
		}
		return err
	}
	return nil
}

func deref(p *int64, fallback int64) int64 {
	if p == nil {
		return fallback
	}
	return *p
}

func sameInt64Ptr(a, b *int64) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}
