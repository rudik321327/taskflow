package server

import (
	"context"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/taskflow/taskflow/internal/model"
	"github.com/taskflow/taskflow/internal/repository"
	"github.com/taskflow/taskflow/proto/notificationpb"
)

type NotificationServer struct {
	repo repository.NotificationRepository
	log  *zap.Logger
}

func NewNotificationServer(repo repository.NotificationRepository, log *zap.Logger) *NotificationServer {
	return &NotificationServer{repo: repo, log: log}
}

func (s *NotificationServer) SendNotification(
	ctx context.Context, req *notificationpb.SendNotificationRequest,
) (*notificationpb.SendNotificationResponse, error) {
	n := &model.Notification{
		UserID:  req.UserId,
		Type:    req.Type,
		Message: req.Message,
	}
	id, err := s.repo.Create(ctx, n)
	if err != nil {
		s.log.Error("notification persist failed",
			zap.Int64("user_id", req.UserId),
			zap.String("type", req.Type),
			zap.Error(err),
		)
		return nil, err
	}
	s.log.Info("notification delivered",
		zap.Int64("id", id),
		zap.Int64("user_id", req.UserId),
		zap.String("type", req.Type),
	)
	return &notificationpb.SendNotificationResponse{Id: id, Delivered: true}, nil
}

func (s *NotificationServer) GetNotifications(
	ctx context.Context, req *notificationpb.GetNotificationsRequest,
) (*notificationpb.GetNotificationsResponse, error) {
	items, total, err := s.repo.ListByUser(ctx, req.UserId, int(req.Page), int(req.Limit))
	if err != nil {
		return nil, err
	}
	out := make([]*notificationpb.Notification, 0, len(items))
	for _, n := range items {
		out = append(out, &notificationpb.Notification{
			Id:        n.ID,
			UserId:    n.UserID,
			Type:      n.Type,
			Message:   n.Message,
			IsRead:    n.IsRead,
			CreatedAt: n.CreatedAt.Unix(),
		})
	}
	return &notificationpb.GetNotificationsResponse{Items: out, Total: total}, nil
}

func Serve(ctx context.Context, addr string, srv *NotificationServer, log *zap.Logger) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	notificationpb.RegisterNotificationServiceServer(grpcServer, srv)

	go func() {
		<-ctx.Done()
		log.Info("gRPC server stopping (context cancelled)")
		grpcServer.GracefulStop()
	}()

	log.Info("gRPC server listening", zap.String("addr", addr))
	return grpcServer.Serve(lis)
}
