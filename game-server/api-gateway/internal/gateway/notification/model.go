package notification

import (
	"context"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/notification"
)

type NotificationClient interface {
	GetNotification(ctx context.Context, req *pb.NotificationRequest) (*pb.NotificationResponse, error)
	MarkNotificationAsRead(ctx context.Context, req *pb.MarkNotificationAsReadRequest) (*pb.MarkNotificationAsReadResponse, error)
	MarkAllNotificationsAsRead(ctx context.Context, req *pb.MarkAllNotificationsAsReadRequest) (*pb.MarkAllNotificationsAsReadResponse, error)
}
