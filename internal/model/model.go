package model

import "time"

type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusReview     TaskStatus = "review"
	TaskStatusDone       TaskStatus = "done"
)

func (s TaskStatus) Valid() bool {
	switch s {
	case TaskStatusTodo, TaskStatusInProgress, TaskStatusReview, TaskStatusDone:
		return true
	}
	return false
}

type TaskPriority string

const (
	PriorityLow      TaskPriority = "low"
	PriorityMedium   TaskPriority = "medium"
	PriorityHigh     TaskPriority = "high"
	PriorityCritical TaskPriority = "critical"
)

func (p TaskPriority) Valid() bool {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh, PriorityCritical:
		return true
	}
	return false
}

type ProjectRole string

const (
	RoleOwner  ProjectRole = "owner"
	RoleAdmin  ProjectRole = "admin"
	RoleMember ProjectRole = "member"
)

type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"created_at"`
}

type Project struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     int64     `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type ProjectMember struct {
	ID        int64       `json:"id"`
	ProjectID int64       `json:"project_id"`
	UserID    int64       `json:"user_id"`
	UserName  string      `json:"user_name,omitempty"`
	UserEmail string      `json:"user_email,omitempty"`
	Role      ProjectRole `json:"role"`
	JoinedAt  time.Time   `json:"joined_at"`
}

type Task struct {
	ID           int64        `json:"id"`
	ProjectID    int64        `json:"project_id"`
	CreatorID    int64        `json:"creator_id"`
	AssigneeID   *int64       `json:"assignee_id,omitempty"`
	AssigneeName *string      `json:"assignee_name,omitempty"`
	Title        string       `json:"title"`
	Description  string       `json:"description"`
	Status       TaskStatus   `json:"status"`
	Priority     TaskPriority `json:"priority"`
	DueDate      *time.Time   `json:"due_date,omitempty"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

type Comment struct {
	ID         int64     `json:"id"`
	TaskID     int64     `json:"task_id"`
	AuthorID   int64     `json:"author_id"`
	AuthorName string    `json:"author_name,omitempty"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
}

type Notification struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

type ProjectStats struct {
	ProjectID    int64          `json:"project_id"`
	TotalTasks   int64          `json:"total_tasks"`
	TasksByStatus map[string]int64 `json:"tasks_by_status"`
	MemberCount  int64          `json:"member_count"`
}

type UserStats struct {
	UserID          int64 `json:"user_id"`
	AssignedTasks   int64 `json:"assigned_tasks"`
	CreatedTasks    int64 `json:"created_tasks"`
	CompletedTasks  int64 `json:"completed_tasks"`
	ProjectsOwned   int64 `json:"projects_owned"`
	ProjectsJoined  int64 `json:"projects_joined"`
}
