package main

import (
	"context"
	"flag"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/taskflow/taskflow/internal/auth"
	"github.com/taskflow/taskflow/internal/cache"
	"github.com/taskflow/taskflow/internal/config"
	"github.com/taskflow/taskflow/internal/grpc/client"
	grpcserver "github.com/taskflow/taskflow/internal/grpc/server"
	"github.com/taskflow/taskflow/internal/handler"
	"github.com/taskflow/taskflow/internal/logger"
	"github.com/taskflow/taskflow/internal/repository"
	"github.com/taskflow/taskflow/internal/service"
	"github.com/taskflow/taskflow/internal/worker"
)

func main() {
	mode := flag.String("mode", "api", "process mode: api | grpc")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	log, err := logger.New(cfg.App.Env)
	if err != nil {
		panic(err)
	}
	defer func() { _ = log.Sync() }()

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	switch *mode {
	case "api":
		runAPI(rootCtx, cfg, log)
	case "grpc":
		runGRPC(rootCtx, cfg, log)
	default:
		log.Fatal("unknown mode", zap.String("mode", *mode))
	}
}

func runAPI(ctx context.Context, cfg *config.Config, log *zap.Logger) {
	db, err := repository.NewPool(ctx, cfg.DB.DSN(), cfg.DB.MaxConns)
	if err != nil {
		log.Fatal("postgres connect", zap.Error(err))
	}
	defer db.Close()

	rdb, err := cache.NewClient(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Fatal("redis connect", zap.Error(err))
	}
	defer func() { _ = rdb.Close() }()
	c := cache.NewRedis(rdb)

	notifyClient, err := client.NewNotificationClient(cfg.GRPC.NotificationAddr, log)
	if err != nil {
		log.Fatal("notification gRPC client", zap.Error(err))
	}
	defer func() { _ = notifyClient.Close() }()

	users := repository.NewUserRepository(db)
	projects := repository.NewProjectRepository(db)
	tasks := repository.NewTaskRepository(db)
	comments := repository.NewCommentRepository(db)
	stats := repository.NewStatsRepository(db)

	pool := worker.NewPool(cfg.Worker.PoolSize, cfg.Worker.QueueSize, log,
		func(hctx context.Context, e worker.Event) error {
			return notifyClient.Send(hctx, e.UserID, string(e.Type), e.Message)
		},
	)
	pool.Start(ctx)
	defer pool.Stop()

	issuer := auth.NewIssuer(cfg.JWT.Secret, cfg.JWT.TTL, cfg.App.Name)
	authSvc := service.NewAuthService(users, issuer)
	projSvc := service.NewProjectService(projects, pool)
	taskSvc := service.NewTaskService(tasks, projects, pool)
	cmtSvc := service.NewCommentService(comments, tasks, projects, pool)
	statSvc := service.NewStatsService(stats, projects, c)

	router := handler.NewRouter(cfg, log, issuer, c, handler.Handlers{
		Auth:    handler.NewAuthHandler(authSvc),
		Project: handler.NewProjectHandler(projSvc),
		Task:    handler.NewTaskHandler(taskSvc),
		Comment: handler.NewCommentHandler(cmtSvc),
		Stats:   handler.NewStatsHandler(statSvc),
	})

	srv := &http.Server{
		Addr:              ":" + cfg.App.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Info("HTTP server listening", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("http listen", zap.Error(err))
		}
	}()

	<-ctx.Done()
	log.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.App.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("http shutdown", zap.Error(err))
	}
	log.Info("HTTP server stopped")
}

func runGRPC(ctx context.Context, cfg *config.Config, log *zap.Logger) {
	db, err := repository.NewPool(ctx, cfg.DB.DSN(), cfg.DB.MaxConns)
	if err != nil {
		log.Fatal("postgres connect", zap.Error(err))
	}
	defer db.Close()

	srv := grpcserver.NewNotificationServer(repository.NewNotificationRepository(db), log)
	if err := grpcserver.Serve(ctx, ":"+cfg.GRPC.Port, srv, log); err != nil {
		log.Fatal("grpc serve", zap.Error(err))
	}
}
