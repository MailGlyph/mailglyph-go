package mailrify

import (
	"fmt"
	"net/http"
)

// MailrifyError is the base error type returned by the Mailrify API.
type MailrifyError struct {
	StatusCode int
	Code       int
	Type       string
	Message    string
	Time       int64
	RawBody    string
}

// Error returns a human-readable error message.
func (e *MailrifyError) Error() string {
	if e == nil {
		return "mailrify: unknown error"
	}
	if e.Message != "" {
		if e.Type != "" {
			return fmt.Sprintf("mailrify: %s (%d): %s", e.Type, e.StatusCode, e.Message)
		}
		return fmt.Sprintf("mailrify: %d: %s", e.StatusCode, e.Message)
	}
	if e.Type != "" {
		return fmt.Sprintf("mailrify: %s (%d)", e.Type, e.StatusCode)
	}
	return fmt.Sprintf("mailrify: request failed with status %d", e.StatusCode)
}

// AuthenticationError represents HTTP 401 authentication failures.
type AuthenticationError struct {
	*MailrifyError
}

// Error returns the underlying Mailrify error string.
func (e *AuthenticationError) Error() string {
	if e == nil || e.MailrifyError == nil {
		return "mailrify: authentication failed"
	}
	return e.MailrifyError.Error()
}

// Unwrap returns the underlying MailrifyError.
func (e *AuthenticationError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.MailrifyError
}

// ValidationError represents HTTP 400 validation failures.
type ValidationError struct {
	*MailrifyError
}

// Error returns the underlying Mailrify error string.
func (e *ValidationError) Error() string {
	if e == nil || e.MailrifyError == nil {
		return "mailrify: validation failed"
	}
	return e.MailrifyError.Error()
}

// Unwrap returns the underlying MailrifyError.
func (e *ValidationError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.MailrifyError
}

// NotFoundError represents HTTP 404 not found failures.
type NotFoundError struct {
	*MailrifyError
}

// Error returns the underlying Mailrify error string.
func (e *NotFoundError) Error() string {
	if e == nil || e.MailrifyError == nil {
		return "mailrify: resource not found"
	}
	return e.MailrifyError.Error()
}

// Unwrap returns the underlying MailrifyError.
func (e *NotFoundError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.MailrifyError
}

// RateLimitError represents HTTP 429 rate limit failures.
type RateLimitError struct {
	*MailrifyError
	RetryAfterSeconds int
}

// Error returns the underlying Mailrify error string.
func (e *RateLimitError) Error() string {
	if e == nil || e.MailrifyError == nil {
		return "mailrify: rate limit exceeded"
	}
	return e.MailrifyError.Error()
}

// Unwrap returns the underlying MailrifyError.
func (e *RateLimitError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.MailrifyError
}

// ApiError represents HTTP 5xx server failures.
type ApiError struct {
	*MailrifyError
}

// Error returns the underlying Mailrify error string.
func (e *ApiError) Error() string {
	if e == nil || e.MailrifyError == nil {
		return "mailrify: api error"
	}
	return e.MailrifyError.Error()
}

// Unwrap returns the underlying MailrifyError.
func (e *ApiError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.MailrifyError
}

func newValidationError(message string) error {
	return &ValidationError{
		MailrifyError: &MailrifyError{
			StatusCode: http.StatusBadRequest,
			Type:       "validation_error",
			Message:    message,
		},
	}
}
