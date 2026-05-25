package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrCacheMiss = errors.New("cache: miss")

type Cache interface {
	Get(ctx context.Context, key string, dst any) error
	Set(ctx context.Context, key string, val any, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error

	Incr(ctx context.Context, key string, ttl time.Duration) (int64, error)
}

type redisCache struct{ rdb *redis.Client }

func NewRedis(rdb *redis.Client) Cache { return &redisCache{rdb: rdb} }

func (c *redisCache) Get(ctx context.Context, key string, dst any) error {
	raw, err := c.rdb.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrCacheMiss
		}
		return err
	}
	return json.Unmarshal(raw, dst)
}

func (c *redisCache) Set(ctx context.Context, key string, val any, ttl time.Duration) error {
	payload, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, key, payload, ttl).Err()
}

func (c *redisCache) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return c.rdb.Del(ctx, keys...).Err()
}

func (c *redisCache) Incr(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	pipe := c.rdb.TxPipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, ttl)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}
	return incr.Val(), nil
}

func NewClient(addr, password string, db int) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		return nil, err
	}
	return rdb, nil
}
