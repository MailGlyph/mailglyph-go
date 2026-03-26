package mailglyph

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

func campaignFixtureJSON(name, status string) string {
	return `{"id":"cmp_1","name":"` + name + `","description":"Launch campaign","subject":"Hello","body":"<p>Body</p>","from":"hello@example.com","fromName":"Team","replyTo":"reply@example.com","audienceType":"ALL","status":"` + status + `","totalRecipients":100,"sentCount":50,"deliveredCount":45,"openedCount":10,"clickedCount":5,"bouncedCount":1,"scheduledFor":null,"sentAt":null,"createdAt":"2026-01-01T00:00:00Z","updatedAt":"2026-01-01T00:00:00Z"}`
}

func TestCampaignsList_ReturnsPaginated(t *testing.T) {
	client, server, captured := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[` + campaignFixtureJSON("Launch", "DRAFT") + `],"page":1,"pageSize":20,"total":1,"totalPages":1}`))
	}, "sk_test")
	defer server.Close()

	resp, err := client.Campaigns.List(ctx(), nil)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(resp.Data) != 1 || resp.Total != 1 {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if captured.Path != "/campaigns" {
		t.Fatalf("unexpected path: %s", captured.Path)
	}
}

func TestCampaignsList_WithStatusFilter(t *testing.T) {
	client, server, captured := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[],"page":2,"pageSize":10,"total":0,"totalPages":0}`))
	}, "sk_test")
	defer server.Close()

	status := "CANCELLED"
	page := 2
	pageSize := 10
	_, err := client.Campaigns.List(ctx(), &ListCampaignsParams{Page: &page, PageSize: &pageSize, Status: &status})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if !strings.Contains(captured.Query, "status=CANCELLED") || !strings.Contains(captured.Query, "page=2") || !strings.Contains(captured.Query, "pageSize=10") {
		t.Fatalf("unexpected query: %s", captured.Query)
	}
}

func TestCampaignsCreate_RequiredFields(t *testing.T) {
	client, server, captured := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"success":true,"data":` + campaignFixtureJSON("Launch", "DRAFT") + `}`))
	}, "sk_test")
	defer server.Close()

	campaign, err := client.Campaigns.Create(ctx(), &CreateCampaignParams{
		Name:         "Launch",
		Subject:      "Hello",
		Body:         "<p>Body</p>",
		From:         "hello@example.com",
		AudienceType: "ALL",
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if campaign.ID != "cmp_1" || campaign.AudienceType != "ALL" {
		t.Fatalf("unexpected campaign: %+v", campaign)
	}
	payload := decodeBody(t, captured.Body)
	if payload["subject"] != "Hello" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestCampaignsCreate_FilteredAudienceCondition(t *testing.T) {
	client, server, captured := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"success":true,"data":` + campaignFixtureJSON("Filtered", "DRAFT") + `}`))
	}, "sk_test")
	defer server.Close()

	condition := &FilterCondition{
		Logic: "AND",
		Groups: []FilterGroup{{
			Filters: []SegmentFilter{{Field: "data.plan", Operator: "equals", Value: "premium"}},
		}},
	}
	_, err := client.Campaigns.Create(ctx(), &CreateCampaignParams{
		Name:              "Filtered",
		Subject:           "Hello",
		Body:              "<p>Body</p>",
		From:              "hello@example.com",
		AudienceType:      "FILTERED",
		AudienceCondition: condition,
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	payload := decodeBody(t, captured.Body)
	conditionValue, ok := payload["audienceCondition"].(map[string]interface{})
	if !ok || conditionValue["logic"] != "AND" {
		t.Fatalf("unexpected audienceCondition payload: %+v", payload)
	}
}

func TestCampaignsCreate_ValidationError(t *testing.T) {
	client := New("sk_test")
	_, err := client.Campaigns.Create(ctx(), &CreateCampaignParams{Name: "Launch"})
	if err == nil {
		t.Fatal("expected error")
	}
	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}
}

func TestCampaignsGet_ByID(t *testing.T) {
	client, server, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":` + campaignFixtureJSON("Launch", "DRAFT") + `}`))
	}, "sk_test")
	defer server.Close()

	campaign, err := client.Campaigns.Get(ctx(), "cmp_1")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if campaign.ID != "cmp_1" {
		t.Fatalf("unexpected campaign: %+v", campaign)
	}
}

func TestCampaignsGet_NotFound(t *testing.T) {
	client, server, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}, "sk_test")
	defer server.Close()

	_, err := client.Campaigns.Get(ctx(), "missing")
	if err == nil {
		t.Fatal("expected error")
	}
	var notFoundErr *NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("expected NotFoundError, got %T", err)
	}
}

func TestCampaignsUpdate_Partial(t *testing.T) {
	client, server, captured := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":` + campaignFixtureJSON("Updated", "DRAFT") + `}`))
	}, "sk_test")
	defer server.Close()

	name := "Updated"
	campaign, err := client.Campaigns.Update(ctx(), "cmp_1", &UpdateCampaignParams{Name: &name})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if campaign.Name != "Updated" {
		t.Fatalf("unexpected campaign: %+v", campaign)
	}
	payload := decodeBody(t, captured.Body)
	if payload["name"] != "Updated" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestCampaignsSend_Immediate(t *testing.T) {
	client, server, captured := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"data":` + campaignFixtureJSON("Launch", "SENDING") + `,"message":"Campaign is sending"}`))
	}, "sk_test")
	defer server.Close()

	resp, err := client.Campaigns.Send(ctx(), "cmp_1", nil)
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}
	if captured.Path != "/campaigns/cmp_1/send" {
		t.Fatalf("unexpected path: %s", captured.Path)
	}
	if !resp.Success || resp.Data.ID != "cmp_1" || resp.Message == "" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestCampaignsSend_Scheduled(t *testing.T) {
	client, server, captured := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"data":` + campaignFixtureJSON("Launch", "SCHEDULED") + `,"message":"Campaign scheduled"}`))
	}, "sk_test")
	defer server.Close()

	scheduled := "2026-03-01T10:00:00Z"
	resp, err := client.Campaigns.Send(ctx(), "cmp_1", &SendCampaignParams{ScheduledFor: &scheduled})
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}
	payload := decodeBody(t, captured.Body)
	if payload["scheduledFor"] != scheduled {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	if !resp.Success || resp.Data.Status != "SCHEDULED" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestCampaignsCancel_Success(t *testing.T) {
	client, server, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":` + campaignFixtureJSON("Launch", "CANCELLED") + `,"message":"Cancelled"}`))
	}, "sk_test")
	defer server.Close()

	resp, err := client.Campaigns.Cancel(ctx(), "cmp_1")
	if err != nil {
		t.Fatalf("cancel failed: %v", err)
	}
	if !resp.Success || resp.Data.Status != "CANCELLED" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestCampaignsCancel_NotFound(t *testing.T) {
	client, server, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}, "sk_test")
	defer server.Close()

	_, err := client.Campaigns.Cancel(ctx(), "missing")
	if err == nil {
		t.Fatal("expected error")
	}
	var notFoundErr *NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("expected NotFoundError, got %T", err)
	}
}

func TestCampaignsTest_SendTestEmail(t *testing.T) {
	client, server, captured := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"message":"Test email sent"}`))
	}, "sk_test")
	defer server.Close()

	resp, err := client.Campaigns.Test(ctx(), "cmp_1", "preview@example.com")
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
	if !resp.Success {
		t.Fatalf("unexpected response: %+v", resp)
	}
	payload := decodeBody(t, captured.Body)
	if payload["email"] != "preview@example.com" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestCampaignsTest_ValidationError(t *testing.T) {
	client := New("sk_test")
	_, err := client.Campaigns.Test(ctx(), "cmp_1", "")
	if err == nil {
		t.Fatal("expected error")
	}
	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}
}

func TestCampaignsStats_ReturnsAnalytics(t *testing.T) {
	client, server, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"sent":100,"openRate":0.42}}`))
	}, "sk_test")
	defer server.Close()

	resp, err := client.Campaigns.Stats(ctx(), "cmp_1")
	if err != nil {
		t.Fatalf("stats failed: %v", err)
	}
	if !resp.Success {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if resp.Data["sent"] != float64(100) {
		t.Fatalf("unexpected stats: %+v", resp.Data)
	}
}

func TestCampaignsStats_NotFound(t *testing.T) {
	client, server, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}, "sk_test")
	defer server.Close()

	_, err := client.Campaigns.Stats(ctx(), "missing")
	if err == nil {
		t.Fatal("expected error")
	}
	var notFoundErr *NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("expected NotFoundError, got %T", err)
	}
}
