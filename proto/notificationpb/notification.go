package notificationpb

import (
	"context"

	"google.golang.org/grpc"
)

type SendNotificationRequest struct {
	UserId  int64  `json:"user_id"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

type SendNotificationResponse struct {
	Id        int64 `json:"id"`
	Delivered bool  `json:"delivered"`
}

type GetNotificationsRequest struct {
	UserId int64 `json:"user_id"`
	Page   int32 `json:"page"`
	Limit  int32 `json:"limit"`
}

type Notification struct {
	Id        int64  `json:"id"`
	UserId    int64  `json:"user_id"`
	Type      string `json:"type"`
	Message   string `json:"message"`
	IsRead    bool   `json:"is_read"`
	CreatedAt int64  `json:"created_at"`
}

type GetNotificationsResponse struct {
	Items []*Notification `json:"items"`
	Total int64           `json:"total"`
}

type NotificationServiceServer interface {
	SendNotification(context.Context, *SendNotificationRequest) (*SendNotificationResponse, error)
	GetNotifications(context.Context, *GetNotificationsRequest) (*GetNotificationsResponse, error)
}

type NotificationServiceClient interface {
	SendNotification(ctx context.Context, in *SendNotificationRequest, opts ...grpc.CallOption) (*SendNotificationResponse, error)
	GetNotifications(ctx context.Context, in *GetNotificationsRequest, opts ...grpc.CallOption) (*GetNotificationsResponse, error)
}

const (
	ServiceName       = "notification.NotificationService"
	MethodSend        = "/notification.NotificationService/SendNotification"
	MethodGet         = "/notification.NotificationService/GetNotifications"
)

var ServiceDesc = grpc.ServiceDesc{
	ServiceName: ServiceName,
	HandlerType: (*NotificationServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "SendNotification", Handler: handleSend},
		{MethodName: "GetNotifications", Handler: handleGet},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/notification.proto",
}

func RegisterNotificationServiceServer(s grpc.ServiceRegistrar, srv NotificationServiceServer) {
	s.RegisterService(&ServiceDesc, srv)
}

func handleSend(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendNotificationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotificationServiceServer).SendNotification(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: MethodSend}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotificationServiceServer).SendNotification(ctx, req.(*SendNotificationRequest))
	})
}

func handleGet(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetNotificationsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotificationServiceServer).GetNotifications(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: MethodGet}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotificationServiceServer).GetNotifications(ctx, req.(*GetNotificationsRequest))
	})
}

type notificationServiceClient struct{ cc grpc.ClientConnInterface }

func NewNotificationServiceClient(cc grpc.ClientConnInterface) NotificationServiceClient {
	return &notificationServiceClient{cc: cc}
}

func (c *notificationServiceClient) SendNotification(ctx context.Context, in *SendNotificationRequest, opts ...grpc.CallOption) (*SendNotificationResponse, error) {
	out := new(SendNotificationResponse)
	if err := c.cc.Invoke(ctx, MethodSend, in, out, opts...); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *notificationServiceClient) GetNotifications(ctx context.Context, in *GetNotificationsRequest, opts ...grpc.CallOption) (*GetNotificationsResponse, error) {
	out := new(GetNotificationsResponse)
	if err := c.cc.Invoke(ctx, MethodGet, in, out, opts...); err != nil {
		return nil, err
	}
	return out, nil
}
