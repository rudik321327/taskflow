package service

import (
	"context"
	"errors"

	"github.com/taskflow/taskflow/internal/auth"
	"github.com/taskflow/taskflow/internal/dto"
	"github.com/taskflow/taskflow/internal/model"
	"github.com/taskflow/taskflow/internal/repository"
	"github.com/taskflow/taskflow/internal/utils"
)

type authService struct {
	users  repository.UserRepository
	issuer auth.Issuer
}

func NewAuthService(users repository.UserRepository, issuer auth.Issuer) AuthService {
	return &authService{users: users, issuer: issuer}
}

func (s *authService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error) {
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	u := &model.User{
		Email:        req.Email,
		PasswordHash: hash,
		Name:         req.Name,
	}
	if _, err := s.users.Create(ctx, u); err != nil {
		return nil, err
	}
	return s.tokenFor(u)
}

func (s *authService) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	u, err := s.users.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return nil, utils.ErrInvalidCreds
		}
		return nil, err
	}
	if err := auth.ComparePassword(u.PasswordHash, req.Password); err != nil {
		return nil, utils.ErrInvalidCreds
	}
	return s.tokenFor(u)
}

func (s *authService) Me(ctx context.Context, userID int64) (*model.User, error) {
	return s.users.GetByID(ctx, userID)
}

func (s *authService) tokenFor(u *model.User) (*dto.AuthResponse, error) {
	token, exp, err := s.issuer.Issue(u.ID, u.Email)
	if err != nil {
		return nil, err
	}
	return &dto.AuthResponse{
		Token:     token,
		ExpiresAt: exp,
		User:      dto.UserView{ID: u.ID, Email: u.Email, Name: u.Name},
	}, nil
}
