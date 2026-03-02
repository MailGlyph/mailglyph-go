package mailrify

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

func conditionFixture() *FilterCondition {
	return &FilterCondition{
		Logic: "AND",
		Groups: []FilterGroup{
			{Filters: []SegmentFilter{{Field: "data.plan", Operator: "equals", Value: "premium"}}},
		},
	}
}

func TestSegmentsList_ReturnsSegments(t *testing.T) {
	client, server, captured := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"seg_1","name":"Premium","description":"Users on premium","condition":{"logic":"AND","groups":[{"filters":[{"field":"data.plan","operator":"equals","value":"premium"}]}]},"trackMembership":true,"memberCount":12}]`))
	}, "sk_test")
	defer server.Close()

	segments, err := client.Segments.List(ctx())
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(segments) != 1 || segments[0].ID != "seg_1" {
		t.Fatalf("unexpected response: %+v", segments)
	}
	if segments[0].Condition == nil || segments[0].Condition.Logic != "AND" {
		t.Fatalf("unexpected condition: %+v", segments[0].Condition)
	}
	if captured.Path != "/segments" {
		t.Fatalf("unexpected path: %s", captured.Path)
	}
}

func TestSegmentsCreate_WithConditions(t *testing.T) {
	client, server, captured := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"seg_1","name":"Premium","description":"Users on premium","condition":{"logic":"AND","groups":[{"filters":[{"field":"data.plan","operator":"equals","value":"premium"}]}]},"trackMembership":true,"memberCount":12}`))
	}, "sk_test")
	defer server.Close()

	segment, err := client.Segments.Create(ctx(), &CreateSegmentParams{
		Name:        "Premium",
		Condition:   conditionFixture(),
		Description: strPtr("Users on premium"),
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if segment.ID != "seg_1" {
		t.Fatalf("unexpected response: %+v", segment)
	}
	payload := decodeBody(t, captured.Body)
	if payload["name"] != "Premium" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	condition, ok := payload["condition"].(map[string]interface{})
	if !ok || condition["logic"] != "AND" {
		t.Fatalf("unexpected condition payload: %+v", payload)
	}
}

func TestSegmentsGet_ByID(t *testing.T) {
	client, server, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"seg_1","name":"Premium","condition":{"logic":"AND","groups":[{"filters":[{"field":"data.plan","operator":"equals","value":"premium"}]}]},"trackMembership":true,"memberCount":12}`))
	}, "sk_test")
	defer server.Close()

	segment, err := client.Segments.Get(ctx(), "seg_1")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if segment.ID != "seg_1" {
		t.Fatalf("unexpected response: %+v", segment)
	}
}

func TestSegmentsGet_NotFound(t *testing.T) {
	client, server, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}, "sk_test")
	defer server.Close()

	_, err := client.Segments.Get(ctx(), "missing")
	if err == nil {
		t.Fatal("expected error")
	}
	var notFoundErr *NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("expected NotFoundError, got %T", err)
	}
}

func TestSegmentsUpdate_Conditions(t *testing.T) {
	client, server, captured := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"seg_1","name":"Premium+","condition":{"logic":"OR","groups":[{"filters":[{"field":"data.plan","operator":"equals","value":"premium"}]}]},"trackMembership":true,"memberCount":20}`))
	}, "sk_test")
	defer server.Close()

	name := "Premium+"
	condition := &FilterCondition{
		Logic: "OR",
		Groups: []FilterGroup{
			{Filters: []SegmentFilter{{Field: "data.plan", Operator: "equals", Value: "premium"}}},
		},
	}
	segment, err := client.Segments.Update(ctx(), "seg_1", &UpdateSegmentParams{Name: &name, Condition: condition})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if segment.Name != "Premium+" {
		t.Fatalf("unexpected response: %+v", segment)
	}
	payload := decodeBody(t, captured.Body)
	if payload["name"] != "Premium+" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	conditionPayload := payload["condition"].(map[string]interface{})
	if conditionPayload["logic"] != "OR" {
		t.Fatalf("unexpected condition payload: %+v", conditionPayload)
	}
}

func TestSegmentsDelete_Success(t *testing.T) {
	client, server, captured := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}, "sk_test")
	defer server.Close()

	if err := client.Segments.Delete(ctx(), "seg_1"); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if captured.Method != http.MethodDelete || captured.Path != "/segments/seg_1" {
		t.Fatalf("unexpected request: %s %s", captured.Method, captured.Path)
	}
}

func TestSegmentsListContacts_Paginated(t *testing.T) {
	client, server, captured := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"c_1","email":"a@example.com","subscribed":true,"data":{},"createdAt":"2026-01-01T00:00:00Z","updatedAt":"2026-01-01T00:00:00Z"}],"total":1,"page":1,"pageSize":20,"totalPages":1}`))
	}, "sk_test")
	defer server.Close()

	resp, err := client.Segments.ListContacts(ctx(), "seg_1", &ListSegmentContactsParams{Page: intPtr(1), PageSize: intPtr(20)})
	if err != nil {
		t.Fatalf("list contacts failed: %v", err)
	}
	if len(resp.Data) != 1 || resp.Total != 1 {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if !strings.Contains(captured.Query, "page=1") || !strings.Contains(captured.Query, "pageSize=20") {
		t.Fatalf("unexpected query: %s", captured.Query)
	}
}

func TestSegmentsListContacts_NotFound(t *testing.T) {
	client, server, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}, "sk_test")
	defer server.Close()

	_, err := client.Segments.ListContacts(ctx(), "missing", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var notFoundErr *NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("expected NotFoundError, got %T", err)
	}
}

func TestSegmentsAddMembers_Success(t *testing.T) {
	client, server, captured := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"added":1,"notFound":["missing@example.com"]}`))
	}, "sk_test")
	defer server.Close()

	resp, err := client.Segments.AddMembers(ctx(), "seg_static", &StaticSegmentMembersParams{
		Emails: []string{"member@example.com", "missing@example.com"},
	})
	if err != nil {
		t.Fatalf("add members failed: %v", err)
	}
	if resp.Added != 1 || len(resp.NotFound) != 1 || resp.NotFound[0] != "missing@example.com" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if captured.Method != http.MethodPost || captured.Path != "/segments/seg_static/members" {
		t.Fatalf("unexpected request: %s %s", captured.Method, captured.Path)
	}
	payload := decodeBody(t, captured.Body)
	emails := payload["emails"].([]interface{})
	if len(emails) != 2 || emails[0] != "member@example.com" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestSegmentsAddMembers_ValidationError(t *testing.T) {
	client, _, _ := newTestClient(t, func(_ http.ResponseWriter, _ *http.Request) {}, "sk_test")

	_, err := client.Segments.AddMembers(ctx(), "", &StaticSegmentMembersParams{Emails: []string{"a@example.com"}})
	if err == nil {
		t.Fatal("expected error for missing id")
	}
	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError for missing id, got %T", err)
	}

	_, err = client.Segments.AddMembers(ctx(), "seg_static", nil)
	if err == nil {
		t.Fatal("expected error for nil params")
	}
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError for nil params, got %T", err)
	}

	_, err = client.Segments.AddMembers(ctx(), "seg_static", &StaticSegmentMembersParams{})
	if err == nil {
		t.Fatal("expected error for empty emails")
	}
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError for empty emails, got %T", err)
	}
}

func TestSegmentsRemoveMembers_Success(t *testing.T) {
	client, server, captured := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"removed":2}`))
	}, "sk_test")
	defer server.Close()

	resp, err := client.Segments.RemoveMembers(ctx(), "seg_static", &StaticSegmentMembersParams{
		Emails: []string{"member1@example.com", "member2@example.com"},
	})
	if err != nil {
		t.Fatalf("remove members failed: %v", err)
	}
	if resp.Removed != 2 {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if captured.Method != http.MethodDelete || captured.Path != "/segments/seg_static/members" {
		t.Fatalf("unexpected request: %s %s", captured.Method, captured.Path)
	}
	payload := decodeBody(t, captured.Body)
	emails := payload["emails"].([]interface{})
	if len(emails) != 2 || emails[1] != "member2@example.com" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestSegmentsRemoveMembers_NotFound(t *testing.T) {
	client, server, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}, "sk_test")
	defer server.Close()

	_, err := client.Segments.RemoveMembers(ctx(), "missing", &StaticSegmentMembersParams{Emails: []string{"a@example.com"}})
	if err == nil {
		t.Fatal("expected error")
	}
	var notFoundErr *NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("expected NotFoundError, got %T", err)
	}
}

func TestSegmentsRemoveMembers_ValidationError(t *testing.T) {
	client, _, _ := newTestClient(t, func(_ http.ResponseWriter, _ *http.Request) {}, "sk_test")

	_, err := client.Segments.RemoveMembers(ctx(), "", &StaticSegmentMembersParams{Emails: []string{"a@example.com"}})
	if err == nil {
		t.Fatal("expected error for missing id")
	}
	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError for missing id, got %T", err)
	}

	_, err = client.Segments.RemoveMembers(ctx(), "seg_static", nil)
	if err == nil {
		t.Fatal("expected error for nil params")
	}
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError for nil params, got %T", err)
	}

	_, err = client.Segments.RemoveMembers(ctx(), "seg_static", &StaticSegmentMembersParams{})
	if err == nil {
		t.Fatal("expected error for empty emails")
	}
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError for empty emails, got %T", err)
	}
}
