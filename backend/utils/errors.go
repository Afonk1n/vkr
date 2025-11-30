package utils

import (
	"net/http"
)

// APIError represents an API error
type APIError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	return e.Message
}

// NewError creates a new API error
func NewError(message string, code int) *APIError {
	return &APIError{
		Message: message,
		Code:    code,
	}
}

// Common error types
var (
	ErrNotFound      = NewError("Resource not found", http.StatusNotFound)
	ErrUnauthorized  = NewError("Unauthorized", http.StatusUnauthorized)
	ErrForbidden     = NewError("Forbidden", http.StatusForbidden)
	ErrBadRequest    = NewError("Bad request", http.StatusBadRequest)
	ErrInternalError = NewError("Internal server error", http.StatusInternalServerError)
	ErrValidation    = NewError("Validation error", http.StatusBadRequest)
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code"`
}

// HandleError handles errors and returns appropriate HTTP response
func HandleError(err error) (int, ErrorResponse) {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Code, ErrorResponse{
			Error:   http.StatusText(apiErr.Code),
			Message: apiErr.Message,
			Code:    apiErr.Code,
		}
	}

	return http.StatusInternalServerError, ErrorResponse{
		Error:   http.StatusText(http.StatusInternalServerError),
		Message: err.Error(),
		Code:    http.StatusInternalServerError,
	}
}

