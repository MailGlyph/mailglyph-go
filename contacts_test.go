package mailrify

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestContactsList_PaginatedCursor(t *testing.T) {
	client, server, captured := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"contacts":[{"id":"c_1","email":"a@example.com","subscribed":true,"data":{},"createdAt":"2026-01-01T00:00:00Z","updatedAt":"2026-01-01T00:00:00Z"}],"cursor":"next_1","hasMore":true,"total":12}`))
	}, "sk_test")
	defer server.Close()

	limit := 10
	cursor := "cur_1"
	resp, err := client.Contacts.List(ctx(), &ListContactsParams{Limit: &limit, Cursor: &cursor})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(resp.Contacts) != 1 || resp.Cursor == nil || *resp.Cursor != "next_1" || !resp.HasMore {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if !strings.Contains(captured.Query, "limit=10") || !strings.Contains(captured.Query, "cursor=cur_1") {
		t.Fatalf("unexpected query: %s", captured.Query)
	}
}

func TestContactsList_WithFilters(t *testing.T) {
	client, server, captured := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"contacts":[],"hasMore":false}`))
	}, "sk_test")
	defer server.Close()

	subscribed := true
	search := "john"
	_, err := client.Contacts.List(ctx(), &ListContactsParams{Subscribed: &subscribed, Search: &search})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if !strings.Contains(captured.Query, "subscribed=true") || !strings.Contains(captured.Query, "search=john") {
		t.Fatalf("unexpected query: %s", captured.Query)
	}
}

func TestContactsGet_ByID(t *testing.T) {
	client, server, captured := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"c_1","email":"a@example.com","subscribed":true,"data":{"plan":"pro"},"createdAt":"2026-01-01T00:00:00Z","updatedAt":"2026-01-01T00:00:00Z"}`))
	}, "sk_test")
	defer server.Close()

	contact, err := client.Contacts.Get(ctx(), "c_1")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if contact.ID != "c_1" {
		t.Fatalf("unexpected contact: %+v", contact)
	}
	if captured.Path != "/contacts/c_1" {
		t.Fatalf("unexpected path: %s", captured.Path)
	}
}

func TestContactsGet_NotFound(t *testing.T) {
	client, server, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}, "sk_test")
	defer server.Close()

	_, err := client.Contacts.Get(ctx(), "missing")
	if err == nil {
		t.Fatal("expected error")
	}
	var notFoundErr *NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("expected NotFoundError, got %T", err)
	}
}

func TestContactsCreate_New(t *testing.T) {
	client, server, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"c_1","email":"new@example.com","subscribed":true,"data":{},"createdAt":"2026-01-01T00:00:00Z","updatedAt":"2026-01-01T00:00:00Z","_meta":{"isNew":true,"isUpdate":false}}`))
	}, "sk_test")
	defer server.Close()

	resp, err := client.Contacts.Create(ctx(), &CreateContactParams{Email: "new@example.com"})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if !resp.Meta.IsNew || resp.Meta.IsUpdate {
		t.Fatalf("unexpected meta: %+v", resp.Meta)
	}
}

func TestContactsCreate_Upsert(t *testing.T) {
	client, server, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"c_1","email":"existing@example.com","subscribed":true,"data":{},"createdAt":"2026-01-01T00:00:00Z","updatedAt":"2026-01-01T00:00:00Z","_meta":{"isNew":false,"isUpdate":true}}`))
	}, "sk_test")
	defer server.Close()

	resp, err := client.Contacts.Create(ctx(), &CreateContactParams{Email: "existing@example.com"})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if resp.Meta.IsNew || !resp.Meta.IsUpdate {
		t.Fatalf("unexpected meta: %+v", resp.Meta)
	}
}

func TestContactsUpdate_SubscribedOnly(t *testing.T) {
	client, server, captured := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"c_1","email":"a@example.com","subscribed":false,"data":{},"createdAt":"2026-01-01T00:00:00Z","updatedAt":"2026-01-01T00:00:00Z"}`))
	}, "sk_test")
	defer server.Close()

	resp, err := client.Contacts.Update(ctx(), "c_1", &UpdateContactParams{Subscribed: boolPtr(false)})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if resp.Subscribed {
		t.Fatalf("expected unsubscribed, got %+v", resp)
	}
	payload := decodeBody(t, captured.Body)
	if payload["subscribed"] != false {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestContactsUpdate_CustomData(t *testing.T) {
	client, server, captured := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"c_1","email":"a@example.com","subscribed":true,"data":{"plan":"pro"},"createdAt":"2026-01-01T00:00:00Z","updatedAt":"2026-01-01T00:00:00Z"}`))
	}, "sk_test")
	defer server.Close()

	_, err := client.Contacts.Update(ctx(), "c_1", &UpdateContactParams{Data: map[string]interface{}{"plan": "pro"}})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	payload := decodeBody(t, captured.Body)
	data := payload["data"].(map[string]interface{})
	if data["plan"] != "pro" {
		t.Fatalf("unexpected payload data: %+v", data)
	}
}

func TestContactsDelete_Success(t *testing.T) {
	client, server, captured := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}, "sk_test")
	defer server.Close()

	if err := client.Contacts.Delete(ctx(), "c_1"); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if captured.Method != http.MethodDelete || captured.Path != "/contacts/c_1" {
		t.Fatalf("unexpected request: %s %s", captured.Method, captured.Path)
	}
}

func TestContactsDelete_NotFound(t *testing.T) {
	client, server, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}, "sk_test")
	defer server.Close()

	err := client.Contacts.Delete(ctx(), "missing")
	if err == nil {
		t.Fatal("expected error")
	}
	var notFoundErr *NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("expected NotFoundError, got %T", err)
	}
}

func TestContactsCount_UsesTotal(t *testing.T) {
	client, server, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"contacts":[],"hasMore":true,"total":42}`))
	}, "sk_test")
	defer server.Close()

	count, err := client.Contacts.Count(ctx(), nil)
	if err != nil {
		t.Fatalf("count failed: %v", err)
	}
	if count != 42 {
		t.Fatalf("expected count 42, got %d", count)
	}
}
