package service

import (
	"context"
	"errors"

	"github.com/taskflow/taskflow/internal/dto"
	"github.com/taskflow/taskflow/internal/model"
	"github.com/taskflow/taskflow/internal/repository"
	"github.com/taskflow/taskflow/internal/utils"
	"github.com/taskflow/taskflow/internal/worker"
)

type projectService struct {
	repo repository.ProjectRepository
	bus  *worker.Pool
}

func NewProjectService(repo repository.ProjectRepository, bus *worker.Pool) ProjectService {
	return &projectService{repo: repo, bus: bus}
}

func (s *projectService) Create(ctx context.Context, ownerID int64, req dto.CreateProjectRequest) (*model.Project, error) {
	p := &model.Project{
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     ownerID,
	}
	if _, err := s.repo.CreateWithOwner(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *projectService) Get(ctx context.Context, userID, projectID int64) (*model.Project, error) {
	if err := s.requireMember(ctx, projectID, userID); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, projectID)
}

func (s *projectService) List(ctx context.Context, userID int64, page, limit int) ([]model.Project, int64, error) {
	return s.repo.ListForUser(ctx, userID, page, limit)
}

func (s *projectService) Update(ctx context.Context, userID, projectID int64, req dto.UpdateProjectRequest) error {
	if err := s.requireRole(ctx, projectID, userID, model.RoleOwner, model.RoleAdmin); err != nil {
		return err
	}
	return s.repo.Update(ctx, projectID, req.Name, req.Description)
}

func (s *projectService) Delete(ctx context.Context, userID, projectID int64) error {
	if err := s.requireRole(ctx, projectID, userID, model.RoleOwner); err != nil {
		return err
	}
	return s.repo.Delete(ctx, projectID)
}

func (s *projectService) AddMember(ctx context.Context, actorID, projectID int64, req dto.AddMemberRequest) error {
	if err := s.requireRole(ctx, projectID, actorID, model.RoleOwner, model.RoleAdmin); err != nil {
		return err
	}
	role := model.RoleMember
	if req.Role != "" {
		role = model.ProjectRole(req.Role)
	}
	if err := s.repo.AddMember(ctx, projectID, req.UserID, role); err != nil {
		return err
	}
	s.bus.Publish(worker.Event{
		Type:      worker.EventMemberAdded,
		UserID:    req.UserID,
		ActorID:   actorID,
		ProjectID: projectID,
		Message:   "You were added to a project",
	})
	return nil
}

func (s *projectService) ListMembers(ctx context.Context, userID, projectID int64) ([]model.ProjectMember, error) {
	if err := s.requireMember(ctx, projectID, userID); err != nil {
		return nil, err
	}
	return s.repo.ListMembers(ctx, projectID)
}

func (s *projectService) requireMember(ctx context.Context, projectID, userID int64) error {
	if _, err := s.repo.MemberRole(ctx, projectID, userID); err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return utils.ErrForbidden
		}
		return err
	}
	return nil
}

func (s *projectService) requireRole(ctx context.Context, projectID, userID int64, allowed ...model.ProjectRole) error {
	role, err := s.repo.MemberRole(ctx, projectID, userID)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return utils.ErrForbidden
		}
		return err
	}
	for _, a := range allowed {
		if role == a {
			return nil
		}
	}
	return utils.ErrForbidden
}
