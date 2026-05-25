package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/taskflow/taskflow/internal/dto"
	"github.com/taskflow/taskflow/internal/model"
	"github.com/taskflow/taskflow/internal/service"
	"github.com/taskflow/taskflow/internal/utils"
)

func TestTaskService_CreateAndList(t *testing.T) {
	projects := NewMockProjectRepo()
	tasks := NewMockTaskRepo()
	pool := noopPool(t)

	projSvc := service.NewProjectService(projects, pool)
	taskSvc := service.NewTaskService(tasks, projects, pool)

	ctx := context.Background()

	p, err := projSvc.Create(ctx, 1, dto.CreateProjectRequest{Name: "P1"})
	require.NoError(t, err)
	require.NoError(t, projects.AddMember(ctx, p.ID, 2, model.RoleMember))

	t.Run("member can create", func(t *testing.T) {
		assignee := int64(2)
		task, err := taskSvc.Create(ctx, 2, dto.CreateTaskRequest{
			ProjectID:  p.ID,
			Title:      "Write tests",
			AssigneeID: &assignee,
			Priority:   "high",
		})
		require.NoError(t, err)
		require.Equal(t, model.TaskStatusTodo, task.Status)
		require.Equal(t, model.TaskPriority("high"), task.Priority)
	})

	t.Run("non-member cannot create", func(t *testing.T) {
		_, err := taskSvc.Create(ctx, 9, dto.CreateTaskRequest{
			ProjectID: p.ID, Title: "Intrude",
		})
		require.ErrorIs(t, err, utils.ErrForbidden)
	})

	t.Run("list filters by project membership", func(t *testing.T) {
		items, total, err := taskSvc.List(ctx, 2, dto.TaskFilter{ProjectID: &p.ID, Page: 1, Limit: 10})
		require.NoError(t, err)
		require.EqualValues(t, 1, total)
		require.Len(t, items, 1)
	})

	t.Run("list without project defaults to assignee = caller", func(t *testing.T) {
		items, _, err := taskSvc.List(ctx, 2, dto.TaskFilter{Page: 1, Limit: 10})
		require.NoError(t, err)

		require.Len(t, items, 1)
	})
}

func TestTaskService_Update_StatusTransition(t *testing.T) {
	projects := NewMockProjectRepo()
	tasks := NewMockTaskRepo()
	pool := noopPool(t)
	projSvc := service.NewProjectService(projects, pool)
	taskSvc := service.NewTaskService(tasks, projects, pool)

	ctx := context.Background()
	p, _ := projSvc.Create(ctx, 1, dto.CreateProjectRequest{Name: "X"})
	task, err := taskSvc.Create(ctx, 1, dto.CreateTaskRequest{ProjectID: p.ID, Title: "T"})
	require.NoError(t, err)

	statuses := []string{"in_progress", "review", "done"}
	for _, st := range statuses {
		st := st
		t.Run("transition to "+st, func(t *testing.T) {
			updated, err := taskSvc.Update(ctx, 1, task.ID, dto.UpdateTaskRequest{Status: &st})
			require.NoError(t, err)
			require.EqualValues(t, st, updated.Status)
		})
	}
}
