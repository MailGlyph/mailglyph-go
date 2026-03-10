package mailglyph

import (
	"fmt"
	"net/http"
)

// MailGlyphError is the base error type returned by the MailGlyph API.
type MailGlyphError struct {
	StatusCode int
	Code       int
	Type       string
	Message    string
	Time       int64
	RawBody    string
}

// Error returns a human-readable error message.
func (e *MailGlyphError) Error() string {
	if e == nil {
		return "mailglyph: unknown error"
	}
	if e.Message != "" {
		if e.Type != "" {
			return fmt.Sprintf("mailglyph: %s (%d): %s", e.Type, e.StatusCode, e.Message)
		}
		return fmt.Sprintf("mailglyph: %d: %s", e.StatusCode, e.Message)
	}
	if e.Type != "" {
		return fmt.Sprintf("mailglyph: %s (%d)", e.Type, e.StatusCode)
	}
	return fmt.Sprintf("mailglyph: request failed with status %d", e.StatusCode)
}

// AuthenticationError represents HTTP 401 authentication failures.
type AuthenticationError struct {
	*MailGlyphError
}

// Error returns the underlying MailGlyph error string.
func (e *AuthenticationError) Error() string {
	if e == nil || e.MailGlyphError == nil {
		return "mailglyph: authentication failed"
	}
	return e.MailGlyphError.Error()
}

// Unwrap returns the underlying MailGlyphError.
func (e *AuthenticationError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.MailGlyphError
}

// ValidationError represents HTTP 400 validation failures.
type ValidationError struct {
	*MailGlyphError
}

// Error returns the underlying MailGlyph error string.
func (e *ValidationError) Error() string {
	if e == nil || e.MailGlyphError == nil {
		return "mailglyph: validation failed"
	}
	return e.MailGlyphError.Error()
}

// Unwrap returns the underlying MailGlyphError.
func (e *ValidationError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.MailGlyphError
}

// NotFoundError represents HTTP 404 not found failures.
type NotFoundError struct {
	*MailGlyphError
}

// Error returns the underlying MailGlyph error string.
func (e *NotFoundError) Error() string {
	if e == nil || e.MailGlyphError == nil {
		return "mailglyph: resource not found"
	}
	return e.MailGlyphError.Error()
}

// Unwrap returns the underlying MailGlyphError.
func (e *NotFoundError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.MailGlyphError
}

// RateLimitError represents HTTP 429 rate limit failures.
type RateLimitError struct {
	*MailGlyphError
	RetryAfterSeconds int
}

// Error returns the underlying MailGlyph error string.
func (e *RateLimitError) Error() string {
	if e == nil || e.MailGlyphError == nil {
		return "mailglyph: rate limit exceeded"
	}
	return e.MailGlyphError.Error()
}

// Unwrap returns the underlying MailGlyphError.
func (e *RateLimitError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.MailGlyphError
}

// ApiError represents HTTP 5xx server failures.
type ApiError struct {
	*MailGlyphError
}

// Error returns the underlying MailGlyph error string.
func (e *ApiError) Error() string {
	if e == nil || e.MailGlyphError == nil {
		return "mailglyph: api error"
	}
	return e.MailGlyphError.Error()
}

// Unwrap returns the underlying MailGlyphError.
func (e *ApiError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.MailGlyphError
}

func newValidationError(message string) error {
	return &ValidationError{
		MailGlyphError: &MailGlyphError{
			StatusCode: http.StatusBadRequest,
			Type:       "validation_error",
			Message:    message,
		},
	}
}
