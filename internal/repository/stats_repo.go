package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/taskflow/taskflow/internal/model"
)

type statsRepo struct{ db *pgxpool.Pool }

func NewStatsRepository(db *pgxpool.Pool) StatsRepository { return &statsRepo{db: db} }

func (r *statsRepo) ProjectStats(ctx context.Context, projectID int64) (*model.ProjectStats, error) {
	stats := &model.ProjectStats{
		ProjectID:     projectID,
		TasksByStatus: make(map[string]int64),
	}

	const tasksQ = `
		SELECT status, COUNT(*) AS cnt
		FROM tasks
		WHERE project_id = $1
		GROUP BY status`
	rows, err := r.db.Query(ctx, tasksQ, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var status string
		var cnt int64
		if err := rows.Scan(&status, &cnt); err != nil {
			return nil, err
		}
		stats.TasksByStatus[status] = cnt
		stats.TotalTasks += cnt
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	const memberQ = `SELECT COUNT(*) FROM project_members WHERE project_id = $1`
	if err := r.db.QueryRow(ctx, memberQ, projectID).Scan(&stats.MemberCount); err != nil {
		return nil, err
	}
	return stats, nil
}

func (r *statsRepo) UserStats(ctx context.Context, userID int64) (*model.UserStats, error) {
	stats := &model.UserStats{UserID: userID}

	const q = `
		SELECT
		  (SELECT COUNT(*) FROM tasks         WHERE assignee_id = $1)                          AS assigned,
		  (SELECT COUNT(*) FROM tasks         WHERE creator_id  = $1)                          AS created,
		  (SELECT COUNT(*) FROM tasks         WHERE assignee_id = $1 AND status = 'done')      AS completed,
		  (SELECT COUNT(*) FROM projects      WHERE owner_id    = $1)                          AS owned,
		  (SELECT COUNT(*) FROM project_members WHERE user_id   = $1)                          AS joined`

	if err := r.db.QueryRow(ctx, q, userID).Scan(
		&stats.AssignedTasks,
		&stats.CreatedTasks,
		&stats.CompletedTasks,
		&stats.ProjectsOwned,
		&stats.ProjectsJoined,
	); err != nil {
		return nil, err
	}
	return stats, nil
}
