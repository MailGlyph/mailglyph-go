package mailrify

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type capturedRequest struct {
	Method      string
	Path        string
	Query       string
	Auth        string
	UserAgent   string
	ContentType string
	Body        []byte
}

func newTestClient(t *testing.T, handler http.HandlerFunc, apiKey string) (*Client, *httptest.Server, *capturedRequest) {
	t.Helper()

	captured := &capturedRequest{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured.Method = r.Method
		captured.Path = r.URL.Path
		captured.Query = r.URL.RawQuery
		captured.Auth = r.Header.Get("Authorization")
		captured.UserAgent = r.Header.Get("User-Agent")
		captured.ContentType = r.Header.Get("Content-Type")
		if r.Body != nil {
			defer func() {
				if err := r.Body.Close(); err != nil {
					t.Fatalf("close request body: %v", err)
				}
			}()
			b, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("read request body: %v", err)
			}
			captured.Body = b
		}
		handler(w, r)
	}))

	client := New(apiKey, WithBaseURL(server.URL), WithHTTPClient(server.Client()), WithTimeout(2*time.Second))
	return client, server, captured
}

func decodeBody(t *testing.T, body []byte) map[string]interface{} {
	t.Helper()
	if len(body) == 0 {
		return map[string]interface{}{}
	}
	result := map[string]interface{}{}
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&result); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	return result
}

func strPtr(v string) *string { return &v }

func boolPtr(v bool) *bool { return &v }

func intPtr(v int) *int { return &v }

func ctx() context.Context { return context.Background() }
