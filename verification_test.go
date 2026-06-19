package mailglyph

import (
	"bytes"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"
)

func TestVerificationValidate_EnhancedResult(t *testing.T) {
	client, server, captured := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"email":"user@gmail.com","valid":true,"validationMethod":"smtp","smtpStatus":"Valid","smtpDiagnosis":"Mailbox exists and can receive mail.","creditsConsumed":1,"isDisposable":false,"isAlias":false,"isTypo":false,"isPlusAddressed":false,"isRandomInput":false,"isPersonalEmail":true,"isCatchAll":false,"isGreylisted":false,"domainExists":true,"hasWebsite":true,"hasMxRecords":true,"reasons":[]}}`))
	}, "sk_test")
	defer server.Close()

	resp, err := client.Verification.Validate(ctx(), "user@gmail.com")
	if err != nil {
		t.Fatalf("validate failed: %v", err)
	}
	if !resp.Data.Valid || resp.Data.ValidationMethod != "smtp" || resp.Data.SMTPStatus == nil || *resp.Data.SMTPStatus != "Valid" || resp.Data.CreditsConsumed != 1 {
		t.Fatalf("unexpected validation payload: %+v", resp.Data)
	}
	if captured.Method != "POST" || captured.Path != "/v1/verify" {
		t.Fatalf("unexpected request: %s %s", captured.Method, captured.Path)
	}
	payload := decodeBody(t, captured.Body)
	if payload["email"] != "user@gmail.com" {
		t.Fatalf("unexpected email payload: %+v", payload)
	}
}

func TestVerificationCreateBulk_UploadsMultipartFile(t *testing.T) {
	client, server, captured := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"success":true,"data":{"id":"job_123","status":"QUEUED","originalFilename":"emails.csv","fileSizeBytes":35,"localEmailCount":3,"reservedCredits":3,"confirmedEmailCount":null,"creditUsed":null,"valid":0,"invalid":0,"unknown":0,"catchall":0,"duplicates":0,"spamTrap":0,"toxicDomains":0,"readyForDownload":false,"errorCode":null,"errorMessage":null,"lastValidationStatus":null,"createdAt":"2026-06-18T10:12:30.000Z","updatedAt":"2026-06-18T10:13:10.000Z","completedAt":null}}`))
	}, "sk_test")
	defer server.Close()

	resp, err := client.Verification.CreateBulk(ctx(), &CreateBulkEmailValidationParams{
		Filename: "emails.csv",
		Content:  strings.NewReader("email\none@example.com\ntwo@example.com\n"),
	})
	if err != nil {
		t.Fatalf("create bulk failed: %v", err)
	}
	if resp.Data.ID != "job_123" || resp.Data.Status != "QUEUED" {
		t.Fatalf("unexpected job response: %+v", resp.Data)
	}
	if captured.Method != "POST" || captured.Path != "/v1/verify/files" {
		t.Fatalf("unexpected request: %s %s", captured.Method, captured.Path)
	}
	if !strings.HasPrefix(captured.ContentType, "multipart/form-data") {
		t.Fatalf("expected multipart content type, got %q", captured.ContentType)
	}

	form := readCapturedMultipart(t, captured)
	files := form.File["file"]
	if len(files) != 1 || files[0].Filename != "emails.csv" {
		t.Fatalf("unexpected uploaded file metadata: %+v", files)
	}
	file, err := files[0].Open()
	if err != nil {
		t.Fatalf("open uploaded file: %v", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			t.Fatalf("close uploaded file: %v", closeErr)
		}
	}()
	content, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("read uploaded file: %v", err)
	}
	if !bytes.Contains(content, []byte("two@example.com")) {
		t.Fatalf("unexpected uploaded content: %q", content)
	}
}

func TestVerificationListBulk_WithFilters(t *testing.T) {
	client, server, captured := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"items":[{"id":"job_123","status":"COMPLETED","originalFilename":"emails.csv","fileSizeBytes":35,"localEmailCount":3,"reservedCredits":2,"confirmedEmailCount":3,"creditUsed":2,"valid":1,"invalid":1,"unknown":0,"catchall":0,"duplicates":1,"spamTrap":0,"toxicDomains":0,"readyForDownload":true,"errorCode":null,"errorMessage":null,"lastValidationStatus":"finished","createdAt":"2026-06-18T10:12:30.000Z","updatedAt":"2026-06-18T10:14:05.000Z","completedAt":"2026-06-18T10:14:05.000Z"}],"nextCursor":"next_123"}}`))
	}, "sk_test")
	defer server.Close()

	resp, err := client.Verification.ListBulk(ctx(), &ListBulkEmailValidationsParams{
		Limit:  intPtr(10),
		Cursor: strPtr("cur_123"),
		Search: strPtr("emails"),
		Status: strPtr("COMPLETED"),
	})
	if err != nil {
		t.Fatalf("list bulk failed: %v", err)
	}
	if len(resp.Data.Items) != 1 || resp.Data.Items[0].CreditUsed == nil || *resp.Data.Items[0].CreditUsed != 2 {
		t.Fatalf("unexpected list response: %+v", resp.Data)
	}
	if captured.Query != "cursor=cur_123&limit=10&search=emails&status=COMPLETED" {
		t.Fatalf("unexpected query: %s", captured.Query)
	}
}

func TestVerificationGetContinueDownloadDeleteBulk(t *testing.T) {
	requests := 0
	client, server, captured := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requests++
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/verify/files/job_123":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true,"data":{"id":"job_123","status":"COMPLETED","originalFilename":"emails.csv","fileSizeBytes":35,"localEmailCount":3,"reservedCredits":2,"confirmedEmailCount":3,"creditUsed":2,"valid":1,"invalid":1,"unknown":0,"catchall":0,"duplicates":1,"spamTrap":0,"toxicDomains":0,"readyForDownload":true,"errorCode":null,"errorMessage":null,"lastValidationStatus":"finished","createdAt":"2026-06-18T10:12:30.000Z","updatedAt":"2026-06-18T10:14:05.000Z","completedAt":"2026-06-18T10:14:05.000Z"}}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v1/verify/files/job_123/continue":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true,"data":{"id":"job_123","status":"QUEUED","originalFilename":"emails.csv","fileSizeBytes":35,"localEmailCount":3,"reservedCredits":3,"confirmedEmailCount":null,"creditUsed":null,"valid":0,"invalid":0,"unknown":0,"catchall":0,"duplicates":0,"spamTrap":0,"toxicDomains":0,"readyForDownload":false,"errorCode":null,"errorMessage":null,"lastValidationStatus":null,"createdAt":"2026-06-18T10:12:30.000Z","updatedAt":"2026-06-18T10:13:10.000Z","completedAt":null}}`))
		case r.Method == http.MethodGet && r.URL.Path == "/v1/verify/files/job_123/download":
			w.Header().Set("Content-Type", "text/csv")
			w.Header().Set("Content-Disposition", `attachment; filename="job_123.csv"`)
			_, _ = w.Write([]byte("email,status\none@example.com,Valid\n"))
		case r.Method == http.MethodDelete && r.URL.Path == "/v1/verify/files/job_123":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true,"data":{"refundedCredits":1}}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}, "sk_test")
	defer server.Close()

	got, err := client.Verification.GetBulk(ctx(), "job_123")
	if err != nil {
		t.Fatalf("get bulk failed: %v", err)
	}
	if got.Data.Status != "COMPLETED" {
		t.Fatalf("unexpected get response: %+v", got.Data)
	}

	continued, err := client.Verification.ContinueBulk(ctx(), "job_123")
	if err != nil {
		t.Fatalf("continue bulk failed: %v", err)
	}
	if continued.Data.Status != "QUEUED" {
		t.Fatalf("unexpected continue response: %+v", continued.Data)
	}

	download, err := client.Verification.DownloadBulk(ctx(), "job_123", &DownloadBulkEmailValidationParams{Filter: strPtr("valid"), Format: strPtr("csv")})
	if err != nil {
		t.Fatalf("download bulk failed: %v", err)
	}
	if download.ContentType != "text/csv" || download.Filename != "job_123.csv" || !bytes.Contains(download.Content, []byte("one@example.com")) {
		t.Fatalf("unexpected download: %+v", download)
	}
	if captured.Query != "filter=valid&format=csv" {
		t.Fatalf("unexpected download query: %s", captured.Query)
	}

	deleted, err := client.Verification.DeleteBulk(ctx(), "job_123")
	if err != nil {
		t.Fatalf("delete bulk failed: %v", err)
	}
	if deleted.Data.RefundedCredits != 1 {
		t.Fatalf("unexpected delete response: %+v", deleted.Data)
	}
	if requests != 4 {
		t.Fatalf("expected 4 requests, got %d", requests)
	}
}

func TestVerificationDownloadBulk_MapsJSONError(t *testing.T) {
	client, server, _ := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"code":400,"error":"validation_error","message":"results are not ready"}`))
	}, "sk_test")
	defer server.Close()

	_, err := client.Verification.DownloadBulk(ctx(), "job_123", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}
}

func TestVerificationCreditsAndLedger(t *testing.T) {
	requests := 0
	client, server, captured := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/verification-credits":
			_, _ = w.Write([]byte(`{"success":true,"data":{"balance":4820,"lowCredits":false}}`))
		case r.Method == http.MethodGet && r.URL.Path == "/v1/verification-credits/ledger":
			_, _ = w.Write([]byte(`{"success":true,"data":{"items":[{"id":"ledger_123","seq":9182,"type":"CONSUME","creditsDelta":-1,"balanceAfter":4820,"source":"single_api","status":"Valid","createdAt":"2026-06-17T10:15:30.000Z"}],"nextCursor":"9000"}}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}, "sk_test")
	defer server.Close()

	credits, err := client.Verification.GetCredits(ctx())
	if err != nil {
		t.Fatalf("get credits failed: %v", err)
	}
	if credits.Data.Balance != 4820 || credits.Data.LowCredits {
		t.Fatalf("unexpected credits response: %+v", credits.Data)
	}

	ledger, err := client.Verification.ListCreditLedger(ctx(), &ListVerificationCreditLedgerParams{Limit: intPtr(1), Cursor: strPtr("9182")})
	if err != nil {
		t.Fatalf("list credit ledger failed: %v", err)
	}
	if len(ledger.Data.Items) != 1 || ledger.Data.Items[0].CreditsDelta != -1 || ledger.Data.NextCursor == nil || *ledger.Data.NextCursor != "9000" {
		t.Fatalf("unexpected ledger response: %+v", ledger.Data)
	}
	if captured.Query != "cursor=9182&limit=1" {
		t.Fatalf("unexpected ledger query: %s", captured.Query)
	}
	if requests != 2 {
		t.Fatalf("expected 2 requests, got %d", requests)
	}
}

func readCapturedMultipart(t *testing.T, captured *capturedRequest) *multipart.Form {
	t.Helper()

	_, params, err := mime.ParseMediaType(captured.ContentType)
	if err != nil {
		t.Fatalf("parse multipart content type: %v", err)
	}
	reader := multipart.NewReader(bytes.NewReader(captured.Body), params["boundary"])
	form, err := reader.ReadForm(1024)
	if err != nil {
		t.Fatalf("read multipart form: %v", err)
	}
	t.Cleanup(func() {
		if err := form.RemoveAll(); err != nil {
			t.Fatalf("remove multipart temp files: %v", err)
		}
	})
	return form
}
