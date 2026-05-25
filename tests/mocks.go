package tests

import (
	"context"
	"sync"

	"github.com/taskflow/taskflow/internal/dto"
	"github.com/taskflow/taskflow/internal/model"
	"github.com/taskflow/taskflow/internal/repository"
	"github.com/taskflow/taskflow/internal/utils"
)

type MockUserRepo struct {
	mu       sync.Mutex
	ByEmail  map[string]*model.User
	ByID     map[int64]*model.User
	NextID   int64
	CreateFn func(ctx context.Context, u *model.User) (int64, error)
}

func NewMockUserRepo() *MockUserRepo {
	return &MockUserRepo{
		ByEmail: make(map[string]*model.User),
		ByID:    make(map[int64]*model.User),
		NextID:  1,
	}
}

func (m *MockUserRepo) Create(ctx context.Context, u *model.User) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.CreateFn != nil {
		return m.CreateFn(ctx, u)
	}
	if _, exists := m.ByEmail[u.Email]; exists {
		return 0, utils.ErrAlreadyExists
	}
	u.ID = m.NextID
	m.NextID++
	m.ByEmail[u.Email] = u
	m.ByID[u.ID] = u
	return u.ID, nil
}

func (m *MockUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.ByEmail[email]
	if !ok {
		return nil, utils.ErrNotFound
	}
	return u, nil
}

func (m *MockUserRepo) GetByID(ctx context.Context, id int64) (*model.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.ByID[id]
	if !ok {
		return nil, utils.ErrNotFound
	}
	return u, nil
}

type MockProjectRepo struct {
	mu       sync.Mutex
	Projects map[int64]*model.Project
	Members  map[int64]map[int64]model.ProjectRole
	NextID   int64
}

func NewMockProjectRepo() *MockProjectRepo {
	return &MockProjectRepo{
		Projects: make(map[int64]*model.Project),
		Members:  make(map[int64]map[int64]model.ProjectRole),
		NextID:   1,
	}
}

func (m *MockProjectRepo) CreateWithOwner(ctx context.Context, p *model.Project) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p.ID = m.NextID
	m.NextID++
	m.Projects[p.ID] = p
	m.Members[p.ID] = map[int64]model.ProjectRole{p.OwnerID: model.RoleOwner}
	return p.ID, nil
}

func (m *MockProjectRepo) GetByID(ctx context.Context, id int64) (*model.Project, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if p, ok := m.Projects[id]; ok {
		return p, nil
	}
	return nil, utils.ErrNotFound
}

func (m *MockProjectRepo) ListForUser(ctx context.Context, userID int64, page, limit int) ([]model.Project, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []model.Project
	for pid, members := range m.Members {
		if _, ok := members[userID]; ok {
			out = append(out, *m.Projects[pid])
		}
	}
	return out, int64(len(out)), nil
}

func (m *MockProjectRepo) Update(ctx context.Context, id int64, name, description *string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.Projects[id]
	if !ok {
		return utils.ErrNotFound
	}
	if name != nil {
		p.Name = *name
	}
	if description != nil {
		p.Description = *description
	}
	return nil
}

func (m *MockProjectRepo) Delete(ctx context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.Projects[id]; !ok {
		return utils.ErrNotFound
	}
	delete(m.Projects, id)
	delete(m.Members, id)
	return nil
}

func (m *MockProjectRepo) AddMember(ctx context.Context, projectID, userID int64, role model.ProjectRole) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.Members[projectID]; !ok {
		m.Members[projectID] = map[int64]model.ProjectRole{}
	}
	if _, exists := m.Members[projectID][userID]; exists {
		return utils.ErrAlreadyExists
	}
	m.Members[projectID][userID] = role
	return nil
}

func (m *MockProjectRepo) ListMembers(ctx context.Context, projectID int64) ([]model.ProjectMember, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []model.ProjectMember
	for uid, role := range m.Members[projectID] {
		out = append(out, model.ProjectMember{
			ProjectID: projectID, UserID: uid, Role: role,
		})
	}
	return out, nil
}

func (m *MockProjectRepo) MemberRole(ctx context.Context, projectID, userID int64) (model.ProjectRole, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if members, ok := m.Members[projectID]; ok {
		if role, ok := members[userID]; ok {
			return role, nil
		}
	}
	return "", utils.ErrNotFound
}

type MockTaskRepo struct {
	mu     sync.Mutex
	Tasks  map[int64]*model.Task
	NextID int64
}

func NewMockTaskRepo() *MockTaskRepo {
	return &MockTaskRepo{Tasks: make(map[int64]*model.Task), NextID: 1}
}

func (m *MockTaskRepo) Create(ctx context.Context, t *model.Task) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t.ID = m.NextID
	m.NextID++
	m.Tasks[t.ID] = t
	return t.ID, nil
}

func (m *MockTaskRepo) GetByID(ctx context.Context, id int64) (*model.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if t, ok := m.Tasks[id]; ok {
		return t, nil
	}
	return nil, utils.ErrNotFound
}

func (m *MockTaskRepo) List(ctx context.Context, f dto.TaskFilter) ([]model.Task, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []model.Task
	for _, t := range m.Tasks {
		if f.ProjectID != nil && t.ProjectID != *f.ProjectID {
			continue
		}
		if f.AssigneeID != nil && (t.AssigneeID == nil || *t.AssigneeID != *f.AssigneeID) {
			continue
		}
		if f.Status != "" && string(t.Status) != f.Status {
			continue
		}
		if f.Priority != "" && string(t.Priority) != f.Priority {
			continue
		}
		out = append(out, *t)
	}
	return out, int64(len(out)), nil
}

func (m *MockTaskRepo) Update(ctx context.Context, id int64, patch repository.TaskPatch) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.Tasks[id]
	if !ok {
		return utils.ErrNotFound
	}
	if patch.Title != nil {
		t.Title = *patch.Title
	}
	if patch.Status != nil {
		t.Status = *patch.Status
	}
	if patch.Priority != nil {
		t.Priority = *patch.Priority
	}
	if patch.AssigneeID != nil {
		t.AssigneeID = *patch.AssigneeID
	}
	return nil
}

func (m *MockTaskRepo) Delete(ctx context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.Tasks[id]; !ok {
		return utils.ErrNotFound
	}
	delete(m.Tasks, id)
	return nil
}

var (
	_ repository.UserRepository    = (*MockUserRepo)(nil)
	_ repository.ProjectRepository = (*MockProjectRepo)(nil)
	_ repository.TaskRepository    = (*MockTaskRepo)(nil)
)
