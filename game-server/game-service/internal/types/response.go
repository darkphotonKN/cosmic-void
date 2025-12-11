package types

import (
	"github.com/darkphotonKN/cosmic-void-server/game-service/common/constants"
)

type MessageWriter interface {
	WriteJSON(v interface{}) error
}

type ResponseBuilder struct {
	writer MessageWriter
}

func NewResponseBuilder(writer MessageWriter) *ResponseBuilder {
	return &ResponseBuilder{
		writer: writer,
	}
}

func (rb *ResponseBuilder) Success(action string, payload map[string]interface{}) error {
	response := ServerResponse{
		Action:  action,
		Payload: payload,
		Success: true,
	}
	return rb.writer.WriteJSON(response)
}

func (rb *ResponseBuilder) Error(action string, errCode constants.ErrorCode, message string) error {
	errorResponse := ServerResponse{
		Action:  action,
		Success: false,
		Error: &ErrorResponse{
			Code:    string(errCode),
			Message: message,
		},
	}
	return rb.writer.WriteJSON(errorResponse)
}
