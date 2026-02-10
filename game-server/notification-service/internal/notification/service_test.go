package notification

import (
	"context"
	"testing"
	"time"

	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockRepository struct {
	createCalled      bool
	createParams      *CreateNotification
	createReturn      *Notification
	createErr         error
	getByUserIDCalled bool
	getByUserIDParams *QueryNotifications
	getByUserIDReturn []Notification
	getByUserIDErr    error
	updateCalled      bool
	updateParams      *UpdateNotification
	updateErr         error
}

func (m *mockRepository) Create(ctx context.Context, createNotification *CreateNotification) (*Notification, error) {
	m.createCalled = true
	m.createParams = createNotification
	return m.createReturn, m.createErr
}

func (m *mockRepository) GetByUserID(ctx context.Context, request *QueryNotifications) ([]Notification, error) {
	m.getByUserIDCalled = true
	m.getByUserIDParams = request
	return m.getByUserIDReturn, m.getByUserIDErr
}

func (m *mockRepository) Update(ctx context.Context, request *UpdateNotification) error {
	m.updateCalled = true
	m.updateParams = request
	return m.updateErr
}

func (m *mockRepository) TestProcessMemberSignedUp_Success(t *testing.T) {
	// ===== Arrange =====
	mockRepo := &mockRepository{
		createReturn: &Notification{
			ID:               uuid.New(),
			UserID:           uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
			NotificationType: "in_app",
			EventType:        "member.signedup",
			Title:            "Welcome to Cosmic Void!",
			Message:          "Welcome John! Thank you for joining Cosmic Void.",
			Data: map[string]any{
				"Name":  "kiki_test",
				"Email": "kiki_test@example.com",
			},
			Read:      false,
			Sent:      false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	service := NewService(mockRepo)

	payload := &commonconstants.MemberSignedUpEventPayload{
		UserID: "550e8400-e29b-41d4-a716-446655440001",
		Name:   "kiki_test",
		Email:  "kiki_test@example.com",
	}

	// ===== Act =====
	err := service.ProcessMemberSignedUp(context.Background(), payload)

	// ===== Assert =====
	assert.NoError(t, err)
	assert.True(t, mockRepo.createCalled, "Create should be called")
	assert.NotNil(t, mockRepo.createParams)
	assert.Equal(t, uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"), mockRepo.createParams.UserID)
	assert.Equal(t, "member.signedup", mockRepo.createParams.EventType)
	assert.Equal(t, "kiki", mockRepo.createParams.Data["Name"])
	assert.Equal(t, "kiki_test@example.com", mockRepo.createParams.Data["Email"])
}

func (m *mockRepository) TestProcessMemberSignedUp_InvalidUUID(t *testing.T) {
	// ===== Arrange =====
	mockRepo := &mockRepository{}
	service := NewService(mockRepo)
	payload := &commonconstants.MemberSignedUpEventPayload{
		UserID: "invalid-uuid",
		Name:   "kiki_test",
		Email:  "kiki_test@example.com",
	}

	// ===== Act =====
	err := service.ProcessMemberSignedUp(context.Background(), payload)

	// ===== Assert =====
	assert.NoError(t, err)
	assert.False(t, mockRepo.createCalled, "Create should not be called when UUID is invalid")
}
