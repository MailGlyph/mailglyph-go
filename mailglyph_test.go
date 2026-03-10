package mailglyph

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestNew_InitializesServices(t *testing.T) {
	client := New("sk_test")
	if client == nil {
		t.Fatal("expected client")
	}
	if client.Emails == nil || client.Events == nil || client.Contacts == nil || client.Campaigns == nil || client.Segments == nil {
		t.Fatal("expected all services to be initialized")
	}
	if client.config.BaseURL != defaultBaseURL {
		t.Fatalf("expected default base url %q, got %q", defaultBaseURL, client.config.BaseURL)
	}
}

func TestNew_OptionsApplied(t *testing.T) {
	hc := &http.Client{}
	client := New("sk_test", WithBaseURL("https://example.com"), WithTimeout(5*time.Second), WithHTTPClient(hc))
	if client.config.BaseURL != "https://example.com" {
		t.Fatalf("unexpected base url: %s", client.config.BaseURL)
	}
	if client.config.Timeout != 5*time.Second {
		t.Fatalf("unexpected timeout: %s", client.config.Timeout)
	}
	if client.http.client != hc {
		t.Fatal("expected custom http client to be used")
	}
	if client.http.client.Timeout != 5*time.Second {
		t.Fatalf("expected timeout to be applied to custom client, got %s", client.http.client.Timeout)
	}
}

func TestHTTPHeaders_BearerAndUserAgent(t *testing.T) {
	client, server, captured := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"email":"user@example.com","valid":true,"isDisposable":false,"isAlias":false,"isTypo":false,"isPlusAddressed":false,"isRandomInput":false,"isPersonalEmail":true,"domainExists":true,"hasWebsite":true,"hasMxRecords":true,"reasons":["ok"]}}`))
	}, "sk_test")
	defer server.Close()

	_, err := client.Emails.Verify(ctx(), "user@example.com")
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}

	if captured.Auth != "Bearer sk_test" {
		t.Fatalf("unexpected auth header: %q", captured.Auth)
	}
	if !strings.HasPrefix(captured.UserAgent, "mailglyph-go/") {
		t.Fatalf("unexpected user agent: %q", captured.UserAgent)
	}
	if captured.ContentType != "application/json" {
		t.Fatalf("unexpected content type: %q", captured.ContentType)
	}
}

func TestErrorMapping_StatusCodes(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
		check  func(error) bool
	}{
		{
			name:   "400 validation",
			status: http.StatusBadRequest,
			body:   `{"code":400,"error":"validation_error","message":"bad input","time":1}`,
			check: func(err error) bool {
				var target *ValidationError
				return errors.As(err, &target)
			},
		},
		{
			name:   "401 auth",
			status: http.StatusUnauthorized,
			body:   `{"code":401,"error":"auth_error","message":"unauthorized","time":1}`,
			check: func(err error) bool {
				var target *AuthenticationError
				return errors.As(err, &target)
			},
		},
		{
			name:   "404 not found",
			status: http.StatusNotFound,
			body:   `{"message":"missing"}`,
			check: func(err error) bool {
				var target *NotFoundError
				return errors.As(err, &target)
			},
		},
		{
			name:   "429 rate limit",
			status: http.StatusTooManyRequests,
			body:   `{"message":"slow down"}`,
			check: func(err error) bool {
				var target *RateLimitError
				return errors.As(err, &target)
			},
		},
		{
			name:   "500 api",
			status: http.StatusInternalServerError,
			body:   `{"message":"server"}`,
			check: func(err error) bool {
				var target *ApiError
				return errors.As(err, &target)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, server, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}, "sk_test")
			defer server.Close()

			_, err := client.Emails.Verify(ctx(), "user@example.com")
			if err == nil {
				t.Fatal("expected error")
			}
			if !tt.check(err) {
				t.Fatalf("unexpected error type: %T (%v)", err, err)
			}
		})
	}
}

func TestRetry_429ThenSuccess(t *testing.T) {
	var calls int32
	client, server, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		count := atomic.AddInt32(&calls, 1)
		if count == 1 {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"message":"rate limited"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"email":"user@example.com","valid":true,"isDisposable":false,"isAlias":false,"isTypo":false,"isPlusAddressed":false,"isRandomInput":false,"isPersonalEmail":true,"domainExists":true,"hasWebsite":true,"hasMxRecords":true,"reasons":["ok"]}}`))
	}, "sk_test")
	defer server.Close()

	_, err := client.Emails.Verify(ctx(), "user@example.com")
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", calls)
	}
}

func TestTimeoutHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"email":"user@example.com","valid":true,"isDisposable":false,"isAlias":false,"isTypo":false,"isPlusAddressed":false,"isRandomInput":false,"isPersonalEmail":true,"domainExists":true,"hasWebsite":true,"hasMxRecords":true,"reasons":["ok"]}}`))
	}))
	defer server.Close()

	client := New("sk_test", WithBaseURL(server.URL), WithHTTPClient(server.Client()), WithTimeout(10*time.Millisecond))
	_, err := client.Emails.Verify(context.Background(), "user@example.com")
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestAuthRules_PublicKeyRestrictions(t *testing.T) {
	pkClient := New("pk_test")
	_, err := pkClient.Emails.Verify(ctx(), "user@example.com")
	if err == nil {
		t.Fatal("expected authentication error")
	}
	var authErr *AuthenticationError
	if !errors.As(err, &authErr) {
		t.Fatalf("expected AuthenticationError, got %T", err)
	}

	skClient := New("sk_test")
	_, err = skClient.Events.Track(ctx(), &TrackEventParams{Email: "user@example.com", Event: "signup"})
	if err == nil {
		t.Fatal("expected authentication error")
	}
	if !errors.As(err, &authErr) {
		t.Fatalf("expected AuthenticationError, got %T", err)
	}
}
