package errors

import "net/http"

// APIError defines uniform error payload.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func (a APIError) Error() string { return a.Message }

// New builds APIError with code + message.
func New(code, message string) APIError {
	return APIError{Code: code, Message: message}
}

// FromError converts unknown errors to APIError.
func FromError(err error) APIError {
	if err == nil {
		return APIError{}
	}
	if e, ok := err.(APIError); ok {
		return e
	}
	return APIError{Code: "internal_error", Message: err.Error()}
}

// StatusCode maps code to HTTP status.
func StatusCode(err APIError) int {
	switch err.Code {
	case "bad_request":
		return http.StatusBadRequest
	case "unauthorized":
		return http.StatusUnauthorized
	case "forbidden":
		return http.StatusForbidden
	case "not_found":
		return http.StatusNotFound
	case "conflict":
		return http.StatusConflict
	case "bad_gateway", "upstream_error":
		return http.StatusBadGateway
	default:
		return http.StatusInternalServerError
	}
}
