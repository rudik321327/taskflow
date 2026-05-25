package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/taskflow/taskflow/internal/dto"
	"github.com/taskflow/taskflow/internal/model"
	"github.com/taskflow/taskflow/internal/utils"
)

type taskRepo struct{ db *pgxpool.Pool }

func NewTaskRepository(db *pgxpool.Pool) TaskRepository { return &taskRepo{db: db} }

var sortable = map[string]string{
	"created_at":      "t.created_at ASC",
	"created_at_desc": "t.created_at DESC",
	"due_date":        "t.due_date ASC NULLS LAST",
	"priority":        "t.priority DESC",
}

func (r *taskRepo) Create(ctx context.Context, t *model.Task) (int64, error) {
	const q = `
		INSERT INTO tasks (project_id, creator_id, assignee_id, title, description, status, priority, due_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`
	err := r.db.QueryRow(ctx, q,
		t.ProjectID, t.CreatorID, t.AssigneeID, t.Title, t.Description,
		t.Status, t.Priority, t.DueDate,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return 0, err
	}
	return t.ID, nil
}

func (r *taskRepo) GetByID(ctx context.Context, id int64) (*model.Task, error) {
	const q = `
		SELECT t.id, t.project_id, t.creator_id, t.assignee_id, u.name,
		       t.title, t.description, t.status, t.priority, t.due_date,
		       t.created_at, t.updated_at
		FROM tasks t
		LEFT JOIN users u ON u.id = t.assignee_id
		WHERE t.id = $1`
	var t model.Task
	err := r.db.QueryRow(ctx, q, id).Scan(
		&t.ID, &t.ProjectID, &t.CreatorID, &t.AssigneeID, &t.AssigneeName,
		&t.Title, &t.Description, &t.Status, &t.Priority, &t.DueDate,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if IsNoRows(err) {
			return nil, utils.ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *taskRepo) List(ctx context.Context, f dto.TaskFilter) ([]model.Task, int64, error) {
	_, limit, offset := utils.NormalizePagination(f.Page, f.Limit)

	var (
		where []string
		args  []any
		idx   = 1
	)
	add := func(clause string, val any) {
		where = append(where, fmt.Sprintf(clause, idx))
		args = append(args, val)
		idx++
	}
	if f.ProjectID != nil {
		add("t.project_id = $%d", *f.ProjectID)
	}
	if f.AssigneeID != nil {
		add("t.assignee_id = $%d", *f.AssigneeID)
	}
	if f.Status != "" {
		add("t.status = $%d", f.Status)
	}
	if f.Priority != "" {
		add("t.priority = $%d", f.Priority)
	}

	whereSQL := ""
	if len(where) > 0 {
		whereSQL = "WHERE " + strings.Join(where, " AND ")
	}

	countQ := "SELECT COUNT(*) FROM tasks t " + whereSQL
	var total int64
	if err := r.db.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	order := utils.SortClause(f.Sort, sortable, "t.created_at DESC")
	listQ := fmt.Sprintf(`
		SELECT t.id, t.project_id, t.creator_id, t.assignee_id, u.name,
		       t.title, t.description, t.status, t.priority, t.due_date,
		       t.created_at, t.updated_at
		FROM tasks t
		LEFT JOIN users u ON u.id = t.assignee_id
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d`, whereSQL, order, idx, idx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, listQ, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.Task, 0, limit)
	for rows.Next() {
		var t model.Task
		if err := rows.Scan(
			&t.ID, &t.ProjectID, &t.CreatorID, &t.AssigneeID, &t.AssigneeName,
			&t.Title, &t.Description, &t.Status, &t.Priority, &t.DueDate,
			&t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, t)
	}
	return items, total, rows.Err()
}

func (r *taskRepo) Update(ctx context.Context, id int64, patch TaskPatch) error {
	var (
		sets []string
		args []any
		idx  = 1
	)
	add := func(col string, val any) {
		sets = append(sets, fmt.Sprintf("%s = $%d", col, idx))
		args = append(args, val)
		idx++
	}
	if patch.Title != nil {
		add("title", *patch.Title)
	}
	if patch.Description != nil {
		add("description", *patch.Description)
	}
	if patch.AssigneeID != nil {
		add("assignee_id", *patch.AssigneeID)
	}
	if patch.Status != nil {
		add("status", *patch.Status)
	}
	if patch.Priority != nil {
		add("priority", *patch.Priority)
	}
	if patch.DueDate != nil {
		add("due_date", *patch.DueDate)
	}
	if len(sets) == 0 {
		return nil
	}

	q := fmt.Sprintf("UPDATE tasks SET %s WHERE id = $%d", strings.Join(sets, ", "), idx)
	args = append(args, id)

	tag, err := r.db.Exec(ctx, q, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return utils.ErrNotFound
	}
	return nil
}

func (r *taskRepo) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return utils.ErrNotFound
	}
	return nil
}
