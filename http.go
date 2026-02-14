package mailrify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type apiKeyType int

const (
	apiKeyTypeUnknown apiKeyType = iota
	apiKeyTypeSecret
	apiKeyTypePublic
)

type httpClient struct {
	apiKey     string
	keyType    apiKeyType
	baseURL    string
	client     *http.Client
	userAgent  string
	maxRetries int
}

type apiErrorPayload struct {
	Code    int    `json:"code"`
	Type    string `json:"error"`
	Message string `json:"message"`
	Time    int64  `json:"time"`
}

func newHTTPClient(cfg ClientConfig, customClient *http.Client) *httpClient {
	client := customClient
	if client == nil {
		client = &http.Client{Timeout: cfg.Timeout}
	} else if cfg.Timeout > 0 {
		client.Timeout = cfg.Timeout
	}

	return &httpClient{
		apiKey:     cfg.APIKey,
		keyType:    detectAPIKeyType(cfg.APIKey),
		baseURL:    strings.TrimRight(cfg.BaseURL, "/"),
		client:     client,
		userAgent:  "mailrify-go/" + Version,
		maxRetries: 3,
	}
}

func detectAPIKeyType(apiKey string) apiKeyType {
	switch {
	case strings.HasPrefix(apiKey, "sk_"):
		return apiKeyTypeSecret
	case strings.HasPrefix(apiKey, "pk_"):
		return apiKeyTypePublic
	default:
		return apiKeyTypeUnknown
	}
}

func (h *httpClient) do(ctx context.Context, method, path string, query url.Values, body interface{}, out interface{}) error {
	if err := h.validateAPIKey(path); err != nil {
		return err
	}

	requestURL, err := h.buildURL(path, query)
	if err != nil {
		return err
	}

	var payload []byte
	if body != nil {
		payload, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("mailrify: encode request body: %w", err)
		}
	}

	for attempt := 0; ; attempt++ {
		var bodyReader io.Reader
		if len(payload) > 0 {
			bodyReader = bytes.NewReader(payload)
		}

		req, err := http.NewRequestWithContext(ctx, method, requestURL, bodyReader)
		if err != nil {
			return fmt.Errorf("mailrify: create request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+h.apiKey)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", h.userAgent)

		resp, err := h.client.Do(req)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return err
			}
			return fmt.Errorf("mailrify: request failed: %w", err)
		}

		responseBody, readErr := io.ReadAll(resp.Body)
		closeErr := resp.Body.Close()
		if readErr != nil {
			return fmt.Errorf("mailrify: read response body: %w", readErr)
		}
		if closeErr != nil {
			return fmt.Errorf("mailrify: close response body: %w", closeErr)
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if out != nil && resp.StatusCode != http.StatusNoContent && len(responseBody) > 0 {
				if err := json.Unmarshal(responseBody, out); err != nil {
					return fmt.Errorf("mailrify: decode response body: %w", err)
				}
			}
			return nil
		}

		mappedErr := mapHTTPError(resp.StatusCode, responseBody, resp.Header)
		if shouldRetry(resp.StatusCode) && attempt < h.maxRetries {
			delay := retryDelay(attempt, resp.Header.Get("Retry-After"))
			if err := sleepWithContext(ctx, delay); err != nil {
				return err
			}
			continue
		}
		return mappedErr
	}
}

func (h *httpClient) validateAPIKey(path string) error {
	if path == "/v1/track" {
		switch h.keyType {
		case apiKeyTypePublic:
			return nil
		case apiKeyTypeSecret:
			return &AuthenticationError{MailrifyError: &MailrifyError{
				StatusCode: http.StatusUnauthorized,
				Type:       "authentication_error",
				Message:    "public API key (pk_*) is required for /v1/track",
			}}
		default:
			return &AuthenticationError{MailrifyError: &MailrifyError{
				StatusCode: http.StatusUnauthorized,
				Type:       "authentication_error",
				Message:    "invalid API key format: expected pk_* for /v1/track",
			}}
		}
	}

	switch h.keyType {
	case apiKeyTypeSecret:
		return nil
	case apiKeyTypePublic:
		return &AuthenticationError{MailrifyError: &MailrifyError{
			StatusCode: http.StatusUnauthorized,
			Type:       "authentication_error",
			Message:    "secret API key (sk_*) is required for this endpoint",
		}}
	default:
		return &AuthenticationError{MailrifyError: &MailrifyError{
			StatusCode: http.StatusUnauthorized,
			Type:       "authentication_error",
			Message:    "invalid API key format: expected sk_* or pk_*",
		}}
	}
}

func (h *httpClient) buildURL(path string, query url.Values) (string, error) {
	base, err := url.Parse(h.baseURL)
	if err != nil {
		return "", fmt.Errorf("mailrify: invalid base url: %w", err)
	}
	rel, err := url.Parse(path)
	if err != nil {
		return "", fmt.Errorf("mailrify: invalid path: %w", err)
	}
	full := base.ResolveReference(rel)
	if len(query) > 0 {
		full.RawQuery = query.Encode()
	}
	return full.String(), nil
}

func mapHTTPError(statusCode int, responseBody []byte, headers http.Header) error {
	payload := apiErrorPayload{}
	if len(responseBody) > 0 {
		_ = json.Unmarshal(responseBody, &payload)
	}

	message := payload.Message
	if message == "" {
		message = http.StatusText(statusCode)
	}
	if message == "" {
		message = "request failed"
	}

	baseErr := &MailrifyError{
		StatusCode: statusCode,
		Code:       payload.Code,
		Type:       payload.Type,
		Message:    message,
		Time:       payload.Time,
		RawBody:    string(responseBody),
	}

	switch {
	case statusCode == http.StatusBadRequest:
		return &ValidationError{MailrifyError: baseErr}
	case statusCode == http.StatusUnauthorized:
		return &AuthenticationError{MailrifyError: baseErr}
	case statusCode == http.StatusNotFound:
		return &NotFoundError{MailrifyError: baseErr}
	case statusCode == http.StatusTooManyRequests:
		retryAfterSeconds := parseRetryAfterSeconds(headers.Get("Retry-After"))
		return &RateLimitError{MailrifyError: baseErr, RetryAfterSeconds: retryAfterSeconds}
	case statusCode >= 500 && statusCode <= 599:
		return &ApiError{MailrifyError: baseErr}
	default:
		return baseErr
	}
}

func shouldRetry(statusCode int) bool {
	return statusCode == http.StatusTooManyRequests || (statusCode >= 500 && statusCode <= 599)
}

func retryDelay(attempt int, retryAfter string) time.Duration {
	if delay := parseRetryAfter(retryAfter); delay > 0 {
		return delay
	}
	base := time.Second * time.Duration(1<<attempt)
	jitter := time.Duration(rand.Int63n(int64(250 * time.Millisecond)))
	return base + jitter
}

func parseRetryAfterSeconds(retryAfter string) int {
	if retryAfter == "" {
		return 0
	}
	if seconds, err := strconv.Atoi(strings.TrimSpace(retryAfter)); err == nil && seconds > 0 {
		return seconds
	}
	t, err := http.ParseTime(retryAfter)
	if err != nil {
		return 0
	}
	delta := time.Until(t)
	if delta <= 0 {
		return 0
	}
	return int(delta.Round(time.Second).Seconds())
}

func parseRetryAfter(retryAfter string) time.Duration {
	if seconds := parseRetryAfterSeconds(retryAfter); seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	return 0
}

func sleepWithContext(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
