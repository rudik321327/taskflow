package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/taskflow/taskflow/internal/model"
	"github.com/taskflow/taskflow/internal/utils"
)

type notificationRepo struct{ db *pgxpool.Pool }

func NewNotificationRepository(db *pgxpool.Pool) NotificationRepository {
	return &notificationRepo{db: db}
}

func (r *notificationRepo) Create(ctx context.Context, n *model.Notification) (int64, error) {
	const q = `
		INSERT INTO notifications (user_id, type, message)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`
	if err := r.db.QueryRow(ctx, q, n.UserID, n.Type, n.Message).
		Scan(&n.ID, &n.CreatedAt); err != nil {
		return 0, err
	}
	return n.ID, nil
}

func (r *notificationRepo) ListByUser(ctx context.Context, userID int64, page, limit int) ([]model.Notification, int64, error) {
	_, _, offset := utils.NormalizePagination(page, limit)

	var total int64
	if err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM notifications WHERE user_id = $1`, userID,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	const q = `
		SELECT id, user_id, type, message, is_read, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, q, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.Notification, 0, limit)
	for rows.Next() {
		var n model.Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Message, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, n)
	}
	return items, total, rows.Err()
}
