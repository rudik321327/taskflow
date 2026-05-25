package client

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/taskflow/taskflow/proto/notificationpb"
)

type NotificationClient struct {
	conn   *grpc.ClientConn
	stub   notificationpb.NotificationServiceClient
	log    *zap.Logger
}

func NewNotificationClient(addr string, log *zap.Logger) (*NotificationClient, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.CallContentSubtype(notificationpb.CodecName)),
	)
	if err != nil {
		return nil, err
	}
	return &NotificationClient{
		conn: conn,
		stub: notificationpb.NewNotificationServiceClient(conn),
		log:  log,
	}, nil
}

func (c *NotificationClient) Close() error { return c.conn.Close() }

func (c *NotificationClient) Send(ctx context.Context, userID int64, ntype, message string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	_, err := c.stub.SendNotification(ctx, &notificationpb.SendNotificationRequest{
		UserId:  userID,
		Type:    ntype,
		Message: message,
	})
	if err != nil {
		c.log.Warn("gRPC SendNotification failed",
			zap.Int64("user_id", userID),
			zap.String("type", ntype),
			zap.Error(err),
		)
	}
	return err
}

func (c *NotificationClient) List(ctx context.Context, userID int64, page, limit int) (*notificationpb.GetNotificationsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return c.stub.GetNotifications(ctx, &notificationpb.GetNotificationsRequest{
		UserId: userID,
		Page:   int32(page),
		Limit:  int32(limit),
	})
}
