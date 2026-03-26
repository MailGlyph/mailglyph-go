//go:build integration

package mailglyph

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type integrationConfig struct {
	APIKey      string
	PublicKey   string
	BaseURL     string
	TestDomain  string
	MemberEmail string
}

// TestIntegrationLocalAPI runs end-to-end integration scenarios against a local MailGlyph API.
func TestIntegrationLocalAPI(t *testing.T) {
	cfg, ok := loadIntegrationConfig(t)
	if !ok {
		return
	}

	ctx := context.Background()
	secretClient := New(cfg.APIKey, WithBaseURL(cfg.BaseURL))
	publicClient := New(cfg.PublicKey, WithBaseURL(cfg.BaseURL))

	unique := fmt.Sprintf("%d", time.Now().UnixNano())
	createdContactID := ""
	createdSegmentID := ""
	createdCampaignID := ""

	defer func() {
		if createdContactID != "" {
			t.Logf("cleanup: deleting contact %s", createdContactID)
			if err := secretClient.Contacts.Delete(ctx, createdContactID); err != nil {
				t.Logf("cleanup warning (contact): %s", formatIntegrationErr(err))
			}
		}
		if createdSegmentID != "" {
			t.Logf("cleanup: deleting segment %s", createdSegmentID)
			if err := secretClient.Segments.Delete(ctx, createdSegmentID); err != nil {
				t.Logf("cleanup warning (segment): %s", formatIntegrationErr(err))
			}
		}
		if createdCampaignID != "" {
			// There is no campaign delete endpoint in the SDK/OpenAPI. Best effort cleanup via cancel.
			t.Logf("cleanup: attempting campaign cancel for %s", createdCampaignID)
			if _, err := secretClient.Campaigns.Cancel(ctx, createdCampaignID); err != nil {
				t.Logf("cleanup warning (campaign cancel): %s", formatIntegrationErr(err))
			}
		}
	}()

	t.Log("step 1: Email — Send")
	subject := "SDK Integration Test"
	body := "<p>Test</p>"
	sendResp, err := secretClient.Emails.Send(ctx, &SendEmailParams{
		To:      cfg.MemberEmail,
		From:    "sdk-test@" + cfg.TestDomain,
		Subject: &subject,
		Body:    &body,
	})
	requireNoStepError(t, "1. Email — Send", err)
	if !sendResp.Success || len(sendResp.Data.Emails) == 0 || sendResp.Data.Emails[0].Contact.ID == "" {
		t.Fatalf("step 1. Email — Send failed: expected success response with contact ID, got %+v", sendResp)
	}
	t.Logf("step 1 ok: contact ID=%s", sendResp.Data.Emails[0].Contact.ID)

	t.Log("step 2: Email — Verify")
	verifyInput := "test@" + cfg.TestDomain
	verifyResp, err := secretClient.Emails.Verify(ctx, verifyInput)
	requireNoStepError(t, "2. Email — Verify", err)
	if verifyResp.Data.Email != verifyInput {
		t.Fatalf("step 2. Email — Verify failed: expected email=%q, got %+v", verifyInput, verifyResp.Data)
	}
	t.Logf("step 2 ok: valid=%v reasons=%v", verifyResp.Data.Valid, verifyResp.Data.Reasons)

	t.Log("step 3: Events — Track (pk_* key)")
	eventName := "sdk_test_event"
	trackResp, err := publicClient.Events.Track(ctx, &TrackEventParams{
		Email: verifyInput,
		Event: eventName,
	})
	requireNoStepError(t, "3. Events — Track", err)
	if !trackResp.Success {
		t.Fatalf("step 3. Events — Track failed: expected success response, got %+v", trackResp)
	}
	t.Logf("step 3 ok: event ID=%s", trackResp.Data.Event)

	t.Log("step 4: Events — Get Names (sk_* key)")
	namesResp, err := secretClient.Events.GetNames(ctx)
	requireNoStepError(t, "4. Events — Get Names", err)
	if len(namesResp.EventNames) == 0 {
		t.Fatalf("step 4. Events — Get Names failed: expected non-empty array")
	}
	if !containsString(namesResp.EventNames, eventName) {
		t.Fatalf("step 4. Events — Get Names failed: expected %q in %v", eventName, namesResp.EventNames)
	}
	t.Logf("step 4 ok: names=%v", namesResp.EventNames)

	t.Log("step 5: Contacts — Full CRUD lifecycle")
	contactEmail := fmt.Sprintf("sdk-integration-%s@%s", unique, cfg.TestDomain)
	createdContact, err := secretClient.Contacts.Create(ctx, &CreateContactParams{
		Email: contactEmail,
		Data:  map[string]interface{}{"source": "sdk-test"},
	})
	requireNoStepError(t, "5. Contacts — Create", err)
	createdContactID = createdContact.ID
	if createdContact.ID == "" || createdContact.Email != contactEmail {
		t.Fatalf("step 5. Contacts — Create failed: unexpected contact %+v", createdContact)
	}

	fetchedContact, err := secretClient.Contacts.Get(ctx, createdContactID)
	requireNoStepError(t, "5. Contacts — Get", err)
	if fetchedContact.ID != createdContactID || fetchedContact.Email != contactEmail {
		t.Fatalf("step 5. Contacts — Get failed: got %+v", fetchedContact)
	}

	updatedContact, err := secretClient.Contacts.Update(ctx, createdContactID, &UpdateContactParams{
		Data: map[string]interface{}{"source": "sdk-test", "updated": true},
	})
	requireNoStepError(t, "5. Contacts — Update", err)
	if updatedContact.Data["updated"] != true {
		t.Fatalf("step 5. Contacts — Update failed: expected updated=true, got %+v", updatedContact.Data)
	}

	contactsList, err := secretClient.Contacts.List(ctx, &ListContactsParams{Limit: intPtr(20)})
	requireNoStepError(t, "5. Contacts — List", err)
	listTotal := len(contactsList.Data)
	if contactsList.Total != nil {
		listTotal = *contactsList.Total
	}
	if listTotal <= 0 {
		t.Fatalf("step 5. Contacts — List failed: expected total > 0, got response %+v", contactsList)
	}

	count, err := secretClient.Contacts.Count(ctx, nil)
	requireNoStepError(t, "5. Contacts — Count", err)
	if count <= 0 {
		t.Fatalf("step 5. Contacts — Count failed: expected count > 0, got %d", count)
	}

	err = secretClient.Contacts.Delete(ctx, createdContactID)
	requireNoStepError(t, "5. Contacts — Delete", err)
	deletedContactID := createdContactID
	createdContactID = ""

	_, err = secretClient.Contacts.Get(ctx, deletedContactID)
	if err == nil {
		t.Fatalf("step 5. Contacts — Get Deleted failed: expected NotFoundError for id=%s", deletedContactID)
	}
	var notFoundErr *NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("step 5. Contacts — Get Deleted failed: expected NotFoundError, got %s", formatIntegrationErr(err))
	}
	t.Log("step 5 ok")

	t.Log("step 6: Campaigns — Full lifecycle")
	campaign, err := secretClient.Campaigns.Create(ctx, &CreateCampaignParams{
		Name:         "SDK Test Campaign " + unique,
		Subject:      "Test",
		Body:         "<p>Test</p>",
		From:         "sdk-test@" + cfg.TestDomain,
		AudienceType: "ALL",
	})
	requireNoStepError(t, "6. Campaigns — Create", err)
	createdCampaignID = campaign.ID
	if campaign.ID == "" {
		t.Fatalf("step 6. Campaigns — Create failed: missing campaign ID")
	}

	campaignByID, err := secretClient.Campaigns.Get(ctx, createdCampaignID)
	requireNoStepError(t, "6. Campaigns — Get", err)
	if campaignByID.Status != "DRAFT" {
		t.Fatalf("step 6. Campaigns — Get failed: expected DRAFT status, got %q", campaignByID.Status)
	}

	updatedSubject := "Updated Test"
	updatedCampaign, err := secretClient.Campaigns.Update(ctx, createdCampaignID, &UpdateCampaignParams{Subject: &updatedSubject})
	requireNoStepError(t, "6. Campaigns — Update", err)
	if updatedCampaign.Subject != updatedSubject {
		t.Fatalf("step 6. Campaigns — Update failed: expected subject=%q, got %+v", updatedSubject, updatedCampaign)
	}

	testResp, err := secretClient.Campaigns.Test(ctx, createdCampaignID, cfg.MemberEmail)
	requireNoStepError(t, "6. Campaigns — Test Email", err)
	if !testResp.Success {
		t.Fatalf("step 6. Campaigns — Test Email failed: expected success response, got %+v", testResp)
	}

	statsResp, err := secretClient.Campaigns.Stats(ctx, createdCampaignID)
	requireNoStepError(t, "6. Campaigns — Stats", err)
	if statsResp.Data == nil {
		t.Fatalf("step 6. Campaigns — Stats failed: expected stats object, got %+v", statsResp)
	}

	// Delete is not available in API spec/SDK. Attempt cancel as best-effort cleanup.
	if _, err := secretClient.Campaigns.Cancel(ctx, createdCampaignID); err != nil {
		t.Logf("step 6 cleanup: campaign cancel unavailable/not applicable: %s", formatIntegrationErr(err))
	} else {
		t.Log("step 6 cleanup: campaign cancel succeeded")
	}
	t.Log("step 6 ok")

	t.Log("step 7: Segments — Full CRUD lifecycle")
	segment, err := secretClient.Segments.Create(ctx, &CreateSegmentParams{
		Name: "SDK Test Segment " + unique,
		Condition: &FilterCondition{
			Logic: "AND",
			Groups: []FilterGroup{{
				Filters: []SegmentFilter{{Field: "email", Operator: "contains", Value: "@" + cfg.TestDomain}},
			}},
		},
	})
	requireNoStepError(t, "7. Segments — Create", err)
	createdSegmentID = segment.ID
	if segment.ID == "" {
		t.Fatalf("step 7. Segments — Create failed: missing segment ID")
	}

	segmentByID, err := secretClient.Segments.Get(ctx, createdSegmentID)
	requireNoStepError(t, "7. Segments — Get", err)
	if segmentByID.Name != "SDK Test Segment "+unique {
		t.Fatalf("step 7. Segments — Get failed: expected name %q, got %q", "SDK Test Segment "+unique, segmentByID.Name)
	}

	updatedSegmentName := "Updated SDK Test Segment " + unique
	updatedSegment, err := secretClient.Segments.Update(ctx, createdSegmentID, &UpdateSegmentParams{Name: &updatedSegmentName})
	requireNoStepError(t, "7. Segments — Update", err)
	if updatedSegment.Name != updatedSegmentName {
		t.Fatalf("step 7. Segments — Update failed: expected name %q, got %+v", updatedSegmentName, updatedSegment)
	}

	segments, err := secretClient.Segments.List(ctx)
	requireNoStepError(t, "7. Segments — List", err)
	foundSegment := false
	for _, s := range segments {
		if s.ID == createdSegmentID {
			foundSegment = true
			break
		}
	}
	if !foundSegment {
		t.Fatalf("step 7. Segments — List failed: segment %q not found", createdSegmentID)
	}

	segmentContacts, err := secretClient.Segments.ListContacts(ctx, createdSegmentID, &ListSegmentContactsParams{Page: intPtr(1), PageSize: intPtr(20)})
	requireNoStepError(t, "7. Segments — List Contacts", err)
	if segmentContacts.Page <= 0 || segmentContacts.PageSize <= 0 || segmentContacts.TotalPages <= 0 {
		t.Fatalf("step 7. Segments — List Contacts failed: expected paginated response, got %+v", segmentContacts)
	}

	err = secretClient.Segments.Delete(ctx, createdSegmentID)
	requireNoStepError(t, "7. Segments — Delete", err)
	createdSegmentID = ""
	t.Log("step 7 ok")
}

func loadIntegrationConfig(t *testing.T) (integrationConfig, bool) {
	t.Helper()
	loadEnvFromFile(t, ".env")

	cfg := integrationConfig{
		APIKey:      strings.TrimSpace(os.Getenv("MAILGLYPH_API_KEY")),
		PublicKey:   strings.TrimSpace(os.Getenv("MAILGLYPH_PUBLIC_KEY")),
		BaseURL:     strings.TrimSpace(os.Getenv("MAILGLYPH_BASE_URL")),
		TestDomain:  strings.TrimSpace(os.Getenv("MAILGLYPH_TEST_DOMAIN")),
		MemberEmail: strings.TrimSpace(os.Getenv("MAILGLYPH_TEST_MEMBER_EMAIL")),
	}

	if cfg.APIKey == "" {
		t.Skip("MAILGLYPH_API_KEY is required for integration tests")
		return integrationConfig{}, false
	}

	if cfg.PublicKey == "" {
		t.Fatal("MAILGLYPH_PUBLIC_KEY is required when MAILGLYPH_API_KEY is set")
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://localhost:8081"
	}
	if cfg.TestDomain == "" {
		t.Fatal("MAILGLYPH_TEST_DOMAIN is required when MAILGLYPH_API_KEY is set")
	}
	if cfg.MemberEmail == "" {
		t.Fatal("MAILGLYPH_TEST_MEMBER_EMAIL is required when MAILGLYPH_API_KEY is set")
	}

	return cfg, true
}

func requireNoStepError(t *testing.T, step string, err error) {
	t.Helper()
	if err == nil {
		return
	}
	t.Fatalf("step %s failed: %s", step, formatIntegrationErr(err))
}

func formatIntegrationErr(err error) string {
	if err == nil {
		return ""
	}
	var mailglyphErr *MailGlyphError
	if errors.As(err, &mailglyphErr) {
		return fmt.Sprintf("%v (status=%d, body=%q)", err, mailglyphErr.StatusCode, mailglyphErr.RawBody)
	}
	return err.Error()
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func loadEnvFromFile(t *testing.T, name string) {
	t.Helper()

	path := filepath.Clean(name)
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return
		}
		t.Fatalf("open %s: %v", path, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			t.Fatalf("close %s: %v", path, closeErr)
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" {
			continue
		}

		// Keep explicitly exported env vars as source-of-truth.
		if _, exists := os.LookupEnv(key); exists {
			continue
		}

		value = strings.Trim(value, `"'`)
		if err := os.Setenv(key, value); err != nil {
			t.Fatalf("set env %s from %s: %v", key, path, err)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
}
