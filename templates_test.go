package mailglyph

import (
	"net/http"
	"strings"
	"testing"
)

func TestTemplatesList_ReturnsPaginatedData(t *testing.T) {
	client, server, captured := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"tpl_1","name":"Welcome","description":"Welcome flow","subject":"Welcome","body":"<p>Hi</p>","text":"Hi","from":"hello@example.com","fromName":"Team","replyTo":"reply@example.com","type":"TRANSACTIONAL","projectId":"proj_1","createdAt":"2026-01-01T00:00:00Z","updatedAt":"2026-01-02T00:00:00Z"}],"total":1,"page":1,"pageSize":20,"totalPages":1}`))
	}, "sk_test")
	defer server.Close()

	limit := 20
	templateType := "TRANSACTIONAL"
	search := "welcome"
	resp, err := client.Templates.List(ctx(), &ListTemplatesParams{
		Limit:  &limit,
		Type:   &templateType,
		Search: &search,
	})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}

	if len(resp.Data) != 1 || resp.Total != 1 || resp.Page != 1 || resp.PageSize != 20 || resp.TotalPages != 1 {
		t.Fatalf("unexpected response: %+v", resp)
	}
	tpl := resp.Data[0]
	if tpl.ID != "tpl_1" || tpl.ProjectID != "proj_1" || tpl.UpdatedAt == "" {
		t.Fatalf("unexpected template: %+v", tpl)
	}
	if tpl.Description == nil || *tpl.Description != "Welcome flow" {
		t.Fatalf("expected description, got %+v", tpl.Description)
	}
	if !strings.Contains(captured.Query, "limit=20") || !strings.Contains(captured.Query, "type=TRANSACTIONAL") || !strings.Contains(captured.Query, "search=welcome") {
		t.Fatalf("unexpected query: %s", captured.Query)
	}
	if captured.Path != "/templates" {
		t.Fatalf("unexpected path: %s", captured.Path)
	}
}

func TestTemplatesList_WithCursor(t *testing.T) {
	client, server, captured := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[],"total":0,"page":1,"pageSize":10,"totalPages":0}`))
	}, "sk_test")
	defer server.Close()

	cursor := "next_1"
	_, err := client.Templates.List(ctx(), &ListTemplatesParams{Cursor: &cursor})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if !strings.Contains(captured.Query, "cursor=next_1") {
		t.Fatalf("unexpected query: %s", captured.Query)
	}
}
