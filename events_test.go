package mailglyph

import (
	"errors"
	"net/http"
	"testing"
)

func TestEventsTrack_SimpleEvent(t *testing.T) {
	client, server, captured := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"contact":"c_1","event":"e_1","timestamp":"2026-01-01T00:00:00Z"}}`))
	}, "pk_test")
	defer server.Close()

	resp, err := client.Events.Track(ctx(), &TrackEventParams{Email: "user@example.com", Event: "signup"})
	if err != nil {
		t.Fatalf("track failed: %v", err)
	}
	if resp.Data.Event != "e_1" {
		t.Fatalf("unexpected event id: %s", resp.Data.Event)
	}
	payload := decodeBody(t, captured.Body)
	if payload["event"] != "signup" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestEventsTrack_WithData(t *testing.T) {
	client, server, captured := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"contact":"c_1","event":"e_2","timestamp":"2026-01-01T00:00:00Z"}}`))
	}, "pk_test")
	defer server.Close()

	_, err := client.Events.Track(ctx(), &TrackEventParams{Email: "user@example.com", Event: "purchase", Data: map[string]interface{}{"amount": 99}})
	if err != nil {
		t.Fatalf("track failed: %v", err)
	}
	payload := decodeBody(t, captured.Body)
	data := payload["data"].(map[string]interface{})
	if data["amount"] != float64(99) {
		t.Fatalf("unexpected data: %+v", data)
	}
}

func TestEventsTrack_RejectsSecretKey(t *testing.T) {
	client := New("sk_test")
	_, err := client.Events.Track(ctx(), &TrackEventParams{Email: "user@example.com", Event: "signup"})
	if err == nil {
		t.Fatal("expected authentication error")
	}
	var authErr *AuthenticationError
	if !errors.As(err, &authErr) {
		t.Fatalf("expected AuthenticationError, got %T", err)
	}
}

func TestEventsTrack_WorksWithPublicKey(t *testing.T) {
	client, server, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"contact":"c_1","event":"e_1","timestamp":"2026-01-01T00:00:00Z"}}`))
	}, "pk_test")
	defer server.Close()

	_, err := client.Events.Track(ctx(), &TrackEventParams{Email: "user@example.com", Event: "signup"})
	if err != nil {
		t.Fatalf("track failed: %v", err)
	}
}

func TestEventsGetNames_ReturnsNames(t *testing.T) {
	client, server, captured := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"eventNames":["signup","purchase"]}`))
	}, "sk_test")
	defer server.Close()

	resp, err := client.Events.GetNames(ctx())
	if err != nil {
		t.Fatalf("get names failed: %v", err)
	}
	if len(resp.EventNames) != 2 {
		t.Fatalf("unexpected names response: %+v", resp.EventNames)
	}
	if captured.Path != "/events/names" || captured.Method != http.MethodGet {
		t.Fatalf("unexpected request: %s %s", captured.Method, captured.Path)
	}
}

func TestEventsGetNames_RejectsPublicKey(t *testing.T) {
	client := New("pk_test")
	_, err := client.Events.GetNames(ctx())
	if err == nil {
		t.Fatal("expected authentication error")
	}
	var authErr *AuthenticationError
	if !errors.As(err, &authErr) {
		t.Fatalf("expected AuthenticationError, got %T", err)
	}
}
