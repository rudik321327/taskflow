package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/taskflow/taskflow/internal/model"
	"github.com/taskflow/taskflow/internal/utils"
)

type userRepo struct{ db *pgxpool.Pool }

func NewUserRepository(db *pgxpool.Pool) UserRepository { return &userRepo{db: db} }

func (r *userRepo) Create(ctx context.Context, u *model.User) (int64, error) {
	const q = `
		INSERT INTO users (email, password_hash, name)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`
	row := r.db.QueryRow(ctx, q, strings.ToLower(u.Email), u.PasswordHash, u.Name)
	if err := row.Scan(&u.ID, &u.CreatedAt); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, utils.ErrAlreadyExists
		}
		return 0, err
	}
	return u.ID, nil
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	const q = `
		SELECT id, email, password_hash, name, created_at
		FROM users
		WHERE LOWER(email) = LOWER($1)`
	var u model.User
	err := r.db.QueryRow(ctx, q, email).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.CreatedAt)
	if err != nil {
		if IsNoRows(err) {
			return nil, utils.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepo) GetByID(ctx context.Context, id int64) (*model.User, error) {
	const q = `
		SELECT id, email, password_hash, name, created_at
		FROM users
		WHERE id = $1`
	var u model.User
	err := r.db.QueryRow(ctx, q, id).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.CreatedAt)
	if err != nil {
		if IsNoRows(err) {
			return nil, utils.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}
