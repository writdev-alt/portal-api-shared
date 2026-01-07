package responses

// ErrorResponse represents error response
type ErrorResponse struct {
	Error   string            `json:"error"`
	Errors  map[string]string `json:"errors,omitempty"`
	Message string            `json:"message,omitempty"`
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(err error) ErrorResponse {
	return ErrorResponse{
		Error: err.Error(),
	}
}

// NewValidationErrorResponse creates a validation error response
func NewValidationErrorResponse(errors map[string]string) ErrorResponse {
	return ErrorResponse{
		Error:  "Validation failed",
		Errors: errors,
	}
}

// NewMessageResponse creates a new message response
func NewMessageResponse(message string) MessageResponse {
	return MessageResponse{
		Message: message,
	}
}
