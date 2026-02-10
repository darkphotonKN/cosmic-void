package notification

import (
	"context"
	"log/slog"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/notification"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

type Handler struct {
	pb.UnimplementedNotificationServiceServer
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// GetNotification implements the gRPC method to get notifications for a user
func (h *Handler) GetNotification(ctx context.Context, req *pb.NotificationRequest) (*pb.NotificationResponse, error) {
	// Parse user ID
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		slog.Error("Invalid user ID format", "user_id", req.UserId, "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID format: %v", err)
	}

	// Build query request
	queryReq := &QueryNotifications{
		UserID: userID,
	}

	// Handle optional pagination parameters
	if req.Limit != nil {
		limit := int(*req.Limit)
		queryReq.Limit = &limit
	}

	if req.Offset != nil {
		offset := int(*req.Offset)
		queryReq.Offset = &offset
	}

	// Get notifications from service
	response, err := h.service.GetNotification(ctx, *queryReq)
	if err != nil {
		slog.Error("Failed to get notifications", "user_id", req.UserId, "error", err)
		return nil, status.Errorf(codes.Internal, "failed to get notifications: %v", err)
	}

	// Convert to protobuf response
	pbNotifications := make([]*pb.NotificationList, 0, len(response.Notifications))
	for _, notif := range response.Notifications {
		// Convert data map to protobuf Struct
		dataStruct, err := structpb.NewStruct(notif.Data)
		if err != nil {
			slog.Warn("Failed to convert notification data to struct", "notification_id", notif.ID, "error", err)
			dataStruct = &structpb.Struct{} // Use empty struct on error
		}

		pbNotifications = append(pbNotifications, &pb.NotificationList{
			UserId:  notif.UserID.String(),
			Title:   notif.Title,
			Message: notif.Message,
			Data:    dataStruct,
		})
	}

	return &pb.NotificationResponse{
		Notifications: pbNotifications,
		Total:         int32(response.Total),
	}, nil
}
