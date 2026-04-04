package notification

import (
	"context"
	"fmt"
	"log/slog"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/notification"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service interface {
	GetNotification(ctx context.Context, payload QueryNotifications) (*NotificationListResponse, error)
	MarkNotificationAsRead(ctx context.Context, payload *UpdateNotification) error
	MarkAllNotificationsAsRead(ctx context.Context, userID uuid.UUID) (int64, error) 
}

type Handler struct {
	pb.UnimplementedNotificationServiceServer
	service Service
}

func NewHandler(service Service) *Handler {
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
			Id:               notif.ID.String(),
			UserId:           notif.UserID.String(),
			Title:            notif.Title,
			Message:          notif.Message,
			NotificationType: notif.NotificationType,
			EventType:        notif.EventType,
			Read:             notif.Read,
			Data:             dataStruct,
			CreatedAt:        timestamppb.New(notif.CreatedAt),
			UpdatedAt:        timestamppb.New(notif.UpdatedAt),
		})
	}

	return &pb.NotificationResponse{
		Notifications: pbNotifications,
		Total:         int32(response.Total),
	}, nil
}

// single read
func (h *Handler) MarkNotificationAsRead(ctx context.Context, req *pb.MarkNotificationAsReadRequest) (*pb.MarkNotificationAsReadResponse, error) {
	notificationId, err := uuid.Parse(req.NotificationId)
	if err != nil {
		slog.Error("Invalid notification_id UUID format", "notification_id", req.NotificationId, "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "notification_id is invalid UUID format")
	}
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		slog.Error("Invalid user_id UUID format", "user_id", req.UserId, "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "user_id is invalid UUID format")
	}

	data := &UpdateNotification{
		ID:     notificationId,
		UserID: userId,
		Read:   true,
	}
	err = h.service.MarkNotificationAsRead(ctx, data)
	if err != nil {
		slog.Error("Failed to mark notification as read", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to mark notification as read: %v", err)
	}

	return &pb.MarkNotificationAsReadResponse{
		Success: true,
		Message: "Notification marked as read successfully",
	}, nil
}

func (h *Handler) MarkAllNotificationsAsRead(ctx context.Context, req *pb.MarkAllNotificationsAsReadRequest) (*pb.MarkAllNotificationsAsReadResponse, error) {
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		slog.Error("Invalid user_id UUID format", "user_id", req.UserId, "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "user_id is invalid UUID format")
	}

	updatedCount, err := h.service.MarkAllNotificationsAsRead(ctx, userId)
	if err != nil {
		slog.Error("Failed to mark all notifications as read", "user_id", userId, "error", err)
		return nil, status.Errorf(codes.Internal, "failed to mark all notifications as read: %v", err)
	}

	// 3. 返回成功 response
	return &pb.MarkAllNotificationsAsReadResponse{
		Success:      true,
		Message:      fmt.Sprintf("Successfully marked %d notifications as read", updatedCount),
		UpdatedCount: int32(updatedCount),
	}, nil

}
