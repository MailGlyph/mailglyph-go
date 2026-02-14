package mailrify

import (
	"errors"
	"net/http"
	"testing"
)

func TestEmailsSend_SimpleStringToFrom(t *testing.T) {
	client, server, captured := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"emails":[{"contact":{"id":"c_1","email":"to@example.com"},"email":"e_1"}],"timestamp":"2026-01-01T00:00:00Z"}}`))
	}, "sk_test")
	defer server.Close()

	subject := "Hello"
	body := "<p>Test</p>"
	resp, err := client.Emails.Send(ctx(), &SendEmailParams{To: "to@example.com", From: "from@example.com", Subject: &subject, Body: &body})
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}
	if resp.Data.Emails[0].Contact.Email != "to@example.com" {
		t.Fatalf("unexpected contact email: %+v", resp.Data.Emails[0])
	}
	if captured.Method != "POST" || captured.Path != "/v1/send" {
		t.Fatalf("unexpected request: %s %s", captured.Method, captured.Path)
	}
	payload := decodeBody(t, captured.Body)
	if payload["to"] != "to@example.com" {
		t.Fatalf("unexpected to value: %+v", payload["to"])
	}
}

func TestEmailsSend_ObjectToFrom(t *testing.T) {
	client, server, captured := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"emails":[],"timestamp":"2026-01-01T00:00:00Z"}}`))
	}, "sk_test")
	defer server.Close()

	recipientName := "Jane"
	senderName := "Mailrify"
	_, err := client.Emails.Send(ctx(), &SendEmailParams{
		To:   &Recipient{Name: &recipientName, Email: "jane@example.com"},
		From: &Recipient{Name: &senderName, Email: "hello@example.com"},
	})
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}
	payload := decodeBody(t, captured.Body)
	toObj := payload["to"].(map[string]interface{})
	if toObj["email"] != "jane@example.com" {
		t.Fatalf("unexpected recipient: %+v", toObj)
	}
}

func TestEmailsSend_ArrayRecipients(t *testing.T) {
	client, server, captured := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"emails":[],"timestamp":"2026-01-01T00:00:00Z"}}`))
	}, "sk_test")
	defer server.Close()

	_, err := client.Emails.Send(ctx(), &SendEmailParams{
		To:   []interface{}{"a@example.com", map[string]interface{}{"email": "b@example.com"}},
		From: "from@example.com",
	})
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}
	payload := decodeBody(t, captured.Body)
	toArr := payload["to"].([]interface{})
	if len(toArr) != 2 {
		t.Fatalf("expected 2 recipients, got %d", len(toArr))
	}
}

func TestEmailsSend_TemplateWithData(t *testing.T) {
	client, server, captured := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"emails":[],"timestamp":"2026-01-01T00:00:00Z"}}`))
	}, "sk_test")
	defer server.Close()

	tpl := "tpl_123"
	_, err := client.Emails.Send(ctx(), &SendEmailParams{To: "user@example.com", From: "from@example.com", Template: &tpl, Data: map[string]interface{}{"name": "John"}})
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}
	payload := decodeBody(t, captured.Body)
	if payload["template"] != "tpl_123" {
		t.Fatalf("unexpected template: %+v", payload["template"])
	}
}

func TestEmailsSend_WithAttachments(t *testing.T) {
	client, server, captured := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"emails":[],"timestamp":"2026-01-01T00:00:00Z"}}`))
	}, "sk_test")
	defer server.Close()

	_, err := client.Emails.Send(ctx(), &SendEmailParams{
		To:   "user@example.com",
		From: "from@example.com",
		Attachments: []Attachment{{
			Filename:    "invoice.pdf",
			Content:     "JVBERi0x",
			ContentType: "application/pdf",
		}},
	})
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}
	payload := decodeBody(t, captured.Body)
	attachments := payload["attachments"].([]interface{})
	if len(attachments) != 1 {
		t.Fatalf("expected one attachment, got %d", len(attachments))
	}
}

func TestEmailsSend_ValidationError(t *testing.T) {
	client, server, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"code":400,"error":"validation_error","message":"to is required","time":1}`))
	}, "sk_test")
	defer server.Close()

	_, err := client.Emails.Send(ctx(), &SendEmailParams{From: "from@example.com"})
	if err == nil {
		t.Fatal("expected error")
	}
	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}
}

func TestEmailsVerify_ValidEmail(t *testing.T) {
	client, server, captured := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"email":"user@gmail.com","valid":true,"isDisposable":false,"isAlias":false,"isTypo":false,"isPlusAddressed":false,"isRandomInput":false,"isPersonalEmail":true,"domainExists":true,"hasWebsite":true,"hasMxRecords":true,"reasons":["Email appears to be valid"]}}`))
	}, "sk_test")
	defer server.Close()

	resp, err := client.Emails.Verify(ctx(), "user@gmail.com")
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if !resp.Data.Valid || resp.Data.IsRandomInput {
		t.Fatalf("unexpected verification payload: %+v", resp.Data)
	}
	if captured.Path != "/v1/verify" {
		t.Fatalf("unexpected path: %s", captured.Path)
	}
}

func TestEmailsVerify_TypoSuggestion(t *testing.T) {
	client, server, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"email":"user@gmial.com","valid":false,"isDisposable":false,"isAlias":false,"isTypo":true,"isPlusAddressed":false,"isRandomInput":false,"isPersonalEmail":false,"domainExists":false,"hasWebsite":false,"hasMxRecords":false,"suggestedEmail":"user@gmail.com","reasons":["Possible typo detected"]}}`))
	}, "sk_test")
	defer server.Close()

	resp, err := client.Emails.Verify(ctx(), "user@gmial.com")
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if resp.Data.SuggestedEmail == nil || *resp.Data.SuggestedEmail != "user@gmail.com" {
		t.Fatalf("unexpected suggestion: %+v", resp.Data.SuggestedEmail)
	}
}

func TestEmailsVerify_ValidationError(t *testing.T) {
	client, server, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"code":400,"error":"validation_error","message":"invalid email"}`))
	}, "sk_test")
	defer server.Close()

	_, err := client.Emails.Verify(ctx(), "invalid")
	if err == nil {
		t.Fatal("expected error")
	}
	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}
}
