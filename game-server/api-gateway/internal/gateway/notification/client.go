package notification

import (
	"context"
	"fmt"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/notification"
	"github.com/darkphotonKN/cosmic-void-server/common/discovery"
)

const (
	serviceName = "notification"
)

type Client struct {
	registry discovery.Registry
}

func NewClient(registry discovery.Registry) NotificationClient {
	return &Client{
		registry: registry,
	}
}

func (c *Client) GetNotification(ctx context.Context, req *pb.NotificationRequest) (*pb.NotificationResponse, error) {
	conn, err := discovery.ServiceConnection(ctx, serviceName, c.registry)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to notification service: %w", err)
	}
	defer conn.Close()

	client := pb.NewNotificationServiceClient(conn)

	response, err := client.GetNotification(ctx, req)
	return response, err
}
