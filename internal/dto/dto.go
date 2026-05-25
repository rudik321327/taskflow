package dto

import "time"

type RegisterRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=72"`
	Name     string `json:"name"     binding:"required,min=1,max=120"`
}

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      UserView  `json:"user"`
}

type UserView struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type CreateProjectRequest struct {
	Name        string `json:"name"        binding:"required,min=1,max=200"`
	Description string `json:"description" binding:"max=2000"`
}

type UpdateProjectRequest struct {
	Name        *string `json:"name"        binding:"omitempty,min=1,max=200"`
	Description *string `json:"description" binding:"omitempty,max=2000"`
}

type AddMemberRequest struct {
	UserID int64  `json:"user_id" binding:"required"`
	Role   string `json:"role"    binding:"omitempty,oneof=admin member"`
}

type CreateTaskRequest struct {
	ProjectID   int64      `json:"project_id"  binding:"required"`
	Title       string     `json:"title"       binding:"required,min=1,max=200"`
	Description string     `json:"description" binding:"max=4000"`
	AssigneeID  *int64     `json:"assignee_id"`
	Priority    string     `json:"priority"    binding:"omitempty,oneof=low medium high critical"`
	DueDate     *time.Time `json:"due_date"`
}

type UpdateTaskRequest struct {
	Title       *string    `json:"title"       binding:"omitempty,min=1,max=200"`
	Description *string    `json:"description" binding:"omitempty,max=4000"`
	AssigneeID  *int64     `json:"assignee_id"`
	Status      *string    `json:"status"      binding:"omitempty,oneof=todo in_progress review done"`
	Priority    *string    `json:"priority"    binding:"omitempty,oneof=low medium high critical"`
	DueDate     *time.Time `json:"due_date"`
}

type TaskFilter struct {
	ProjectID  *int64
	AssigneeID *int64
	Status     string
	Priority   string
	Page       int
	Limit      int
	Sort       string
}

type CreateCommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=4000"`
}

type Pagination struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

type ListResponse[T any] struct {
	Items      []T        `json:"items"`
	Pagination Pagination `json:"pagination"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
