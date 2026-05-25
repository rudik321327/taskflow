package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/taskflow/taskflow/internal/model"
	"github.com/taskflow/taskflow/internal/utils"
)

type commentRepo struct{ db *pgxpool.Pool }

func NewCommentRepository(db *pgxpool.Pool) CommentRepository { return &commentRepo{db: db} }

func (r *commentRepo) Create(ctx context.Context, c *model.Comment) (int64, error) {
	const q = `
		INSERT INTO task_comments (task_id, author_id, content)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`
	if err := r.db.QueryRow(ctx, q, c.TaskID, c.AuthorID, c.Content).
		Scan(&c.ID, &c.CreatedAt); err != nil {
		return 0, err
	}
	return c.ID, nil
}

func (r *commentRepo) ListByTask(ctx context.Context, taskID int64, page, limit int) ([]model.Comment, int64, error) {
	_, _, offset := utils.NormalizePagination(page, limit)

	var total int64
	if err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM task_comments WHERE task_id = $1`, taskID,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	const q = `
		SELECT c.id, c.task_id, c.author_id, u.name, c.content, c.created_at
		FROM task_comments c
		INNER JOIN users u ON u.id = c.author_id
		WHERE c.task_id = $1
		ORDER BY c.created_at ASC
		LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, q, taskID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.Comment, 0, limit)
	for rows.Next() {
		var c model.Comment
		if err := rows.Scan(&c.ID, &c.TaskID, &c.AuthorID, &c.AuthorName, &c.Content, &c.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, c)
	}
	return items, total, rows.Err()
}
