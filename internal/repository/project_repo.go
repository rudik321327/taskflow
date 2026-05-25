package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/taskflow/taskflow/internal/model"
	"github.com/taskflow/taskflow/internal/utils"
)

type projectRepo struct{ db *pgxpool.Pool }

func NewProjectRepository(db *pgxpool.Pool) ProjectRepository { return &projectRepo{db: db} }

func (r *projectRepo) CreateWithOwner(ctx context.Context, p *model.Project) (int64, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}

	defer func() { _ = tx.Rollback(ctx) }()

	const insertProject = `
		INSERT INTO projects (name, description, owner_id)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`
	if err := tx.QueryRow(ctx, insertProject, p.Name, p.Description, p.OwnerID).
		Scan(&p.ID, &p.CreatedAt); err != nil {
		return 0, fmt.Errorf("insert project: %w", err)
	}

	const insertMember = `
		INSERT INTO project_members (project_id, user_id, role)
		VALUES ($1, $2, 'owner')`
	if _, err := tx.Exec(ctx, insertMember, p.ID, p.OwnerID); err != nil {
		return 0, fmt.Errorf("insert owner membership: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}
	return p.ID, nil
}

func (r *projectRepo) GetByID(ctx context.Context, id int64) (*model.Project, error) {
	const q = `
		SELECT id, name, description, owner_id, created_at
		FROM projects
		WHERE id = $1`
	var p model.Project
	err := r.db.QueryRow(ctx, q, id).
		Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt)
	if err != nil {
		if IsNoRows(err) {
			return nil, utils.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *projectRepo) ListForUser(ctx context.Context, userID int64, page, limit int) ([]model.Project, int64, error) {
	_, _, offset := utils.NormalizePagination(page, limit)

	const countQ = `
		SELECT COUNT(*)
		FROM projects p
		INNER JOIN project_members pm ON pm.project_id = p.id
		WHERE pm.user_id = $1`
	var total int64
	if err := r.db.QueryRow(ctx, countQ, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	const q = `
		SELECT p.id, p.name, p.description, p.owner_id, p.created_at
		FROM projects p
		INNER JOIN project_members pm ON pm.project_id = p.id
		WHERE pm.user_id = $1
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, q, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.Project, 0, limit)
	for rows.Next() {
		var p model.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, p)
	}
	return items, total, rows.Err()
}

func (r *projectRepo) Update(ctx context.Context, id int64, name, description *string) error {
	const q = `
		UPDATE projects
		SET name        = COALESCE($2, name),
		    description = COALESCE($3, description)
		WHERE id = $1`
	tag, err := r.db.Exec(ctx, q, id, name, description)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return utils.ErrNotFound
	}
	return nil
}

func (r *projectRepo) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM projects WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return utils.ErrNotFound
	}
	return nil
}

func (r *projectRepo) AddMember(ctx context.Context, projectID, userID int64, role model.ProjectRole) error {
	const q = `
		INSERT INTO project_members (project_id, user_id, role)
		VALUES ($1, $2, $3)`
	if _, err := r.db.Exec(ctx, q, projectID, userID, role); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return utils.ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (r *projectRepo) ListMembers(ctx context.Context, projectID int64) ([]model.ProjectMember, error) {
	const q = `
		SELECT pm.id, pm.project_id, pm.user_id, u.name, u.email, pm.role, pm.joined_at
		FROM project_members pm
		INNER JOIN users u ON u.id = pm.user_id
		WHERE pm.project_id = $1
		ORDER BY pm.joined_at ASC`
	rows, err := r.db.Query(ctx, q, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]model.ProjectMember, 0)
	for rows.Next() {
		var m model.ProjectMember
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.UserID, &m.UserName, &m.UserEmail, &m.Role, &m.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (r *projectRepo) MemberRole(ctx context.Context, projectID, userID int64) (model.ProjectRole, error) {
	const q = `SELECT role FROM project_members WHERE project_id = $1 AND user_id = $2`
	var role model.ProjectRole
	if err := r.db.QueryRow(ctx, q, projectID, userID).Scan(&role); err != nil {
		if IsNoRows(err) {
			return "", utils.ErrNotFound
		}
		return "", err
	}
	return role, nil
}
