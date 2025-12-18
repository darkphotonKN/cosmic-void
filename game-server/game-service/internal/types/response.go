package types

import (
	"fmt"
	"log"

	"github.com/darkphotonKN/cosmic-void-server/game-service/common/constants"
)

type MessageWriter interface {
	WriteJSON(v interface{}) error
}

type ResponseBuilder struct{}

func NewResponseBuilder() *ResponseBuilder {
	return &ResponseBuilder{}
}

func (rb *ResponseBuilder) Success(writer MessageWriter, action string, payload map[string]interface{}) error {
	response := ServerResponse{
		Action:  action,
		Payload: payload,
		Success: true,
	}

	return rb.send(writer, response, action)
}

func (rb *ResponseBuilder) Error(writer MessageWriter, action string, errCode constants.ErrorCode, message string) error {
	errorResponse := ServerResponse{
		Action:  string(action),
		Success: false,
		Error: &ErrorResponse{
			Code:    string(errCode),
			Message: message,
		},
	}

	return rb.send(writer, errorResponse, action)
}

func (rb *ResponseBuilder) send(writer MessageWriter, response ServerResponse, action string) error {
	if writer == nil {
		log.Printf("[ResponseBuilder] Warning: nil writer for action '%s', skipping send", action)
		return nil
	}

	if err := writer.WriteJSON(response); err != nil {
		log.Printf("[ResponseBuilder] Failed to send response for action '%s': %v", action, err)
		return fmt.Errorf("failed to send response: %w", err)
	}

	return nil
}
