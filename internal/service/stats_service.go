package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/taskflow/taskflow/internal/cache"
	"github.com/taskflow/taskflow/internal/model"
	"github.com/taskflow/taskflow/internal/repository"
	"github.com/taskflow/taskflow/internal/utils"
)

type statsService struct {
	repo     repository.StatsRepository
	projects repository.ProjectRepository
	cache    cache.Cache
	ttl      time.Duration
}

func NewStatsService(repo repository.StatsRepository, projects repository.ProjectRepository, c cache.Cache) StatsService {
	return &statsService{repo: repo, projects: projects, cache: c, ttl: 30 * time.Second}
}

func (s *statsService) Project(ctx context.Context, userID, projectID int64) (*model.ProjectStats, error) {
	if _, err := s.projects.MemberRole(ctx, projectID, userID); err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return nil, utils.ErrForbidden
		}
		return nil, err
	}

	key := fmt.Sprintf("stats:project:%d", projectID)
	var cached model.ProjectStats
	if err := s.cache.Get(ctx, key, &cached); err == nil {
		return &cached, nil
	}

	stats, err := s.repo.ProjectStats(ctx, projectID)
	if err != nil {
		return nil, err
	}
	_ = s.cache.Set(ctx, key, stats, s.ttl)
	return stats, nil
}

func (s *statsService) User(ctx context.Context, callerID, targetID int64) (*model.UserStats, error) {
	if callerID != targetID {
		return nil, utils.ErrForbidden
	}

	key := fmt.Sprintf("stats:user:%d", targetID)
	var cached model.UserStats
	if err := s.cache.Get(ctx, key, &cached); err == nil {
		return &cached, nil
	}

	stats, err := s.repo.UserStats(ctx, targetID)
	if err != nil {
		return nil, err
	}
	_ = s.cache.Set(ctx, key, stats, s.ttl)
	return stats, nil
}
