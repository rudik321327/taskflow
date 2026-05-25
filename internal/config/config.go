package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App     AppConfig
	DB      DBConfig
	Redis   RedisConfig
	JWT     JWTConfig
	GRPC    GRPCConfig
	Limiter LimiterConfig
	Worker  WorkerConfig
}

type AppConfig struct {
	Env             string
	Name            string
	Port            string
	ShutdownTimeout time.Duration
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	MaxConns int32
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret string
	TTL    time.Duration
}

type GRPCConfig struct {
	NotificationAddr string
	Port             string
}

type LimiterConfig struct {
	RequestsPerMinute int
}

type WorkerConfig struct {
	PoolSize  int
	QueueSize int
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		App: AppConfig{
			Env:             getEnv("APP_ENV", "development"),
			Name:            getEnv("APP_NAME", "taskflow"),
			Port:            getEnv("APP_PORT", "8080"),
			ShutdownTimeout: getEnvDuration("SHUTDOWN_TIMEOUT", 15*time.Second),
		},
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "taskflow"),
			Password: getEnv("DB_PASSWORD", "taskflow_secret"),
			Name:     getEnv("DB_NAME", "taskflow"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			MaxConns: int32(getEnvInt("DB_MAX_CONNS", 20)),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "dev_secret_change_me"),
			TTL:    getEnvDuration("JWT_TTL", 24*time.Hour),
		},
		GRPC: GRPCConfig{
			NotificationAddr: getEnv("GRPC_NOTIFICATION_ADDR", "localhost:50051"),
			Port:             getEnv("GRPC_PORT", "50051"),
		},
		Limiter: LimiterConfig{
			RequestsPerMinute: getEnvInt("RATE_LIMIT_RPM", 120),
		},
		Worker: WorkerConfig{
			PoolSize:  getEnvInt("WORKER_POOL_SIZE", 3),
			QueueSize: getEnvInt("WORKER_QUEUE_SIZE", 256),
		},
	}

	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("config: JWT_SECRET is required")
	}
	return cfg, nil
}

func (c DBConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Name, c.SSLMode,
	)
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getEnvDuration(key string, def time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
