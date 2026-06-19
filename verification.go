package mailglyph

import (
	"context"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

// VerificationService provides methods for email validation and validation credits.
type VerificationService struct {
	client *Client
}

// Validate checks an email address for validity and deliverability signals.
func (s *VerificationService) Validate(ctx context.Context, email string) (*VerifyEmailResponse, error) {
	if email == "" {
		return nil, newValidationError("email is required")
	}

	response := &VerifyEmailResponse{}
	if err := s.client.http.do(ctx, http.MethodPost, "/v1/verify", nil, &VerifyEmailParams{Email: email}, response); err != nil {
		return nil, err
	}
	return response, nil
}

// CreateBulk creates a bulk email validation job from a TXT, CSV, or XLSX file stream.
func (s *VerificationService) CreateBulk(ctx context.Context, params *CreateBulkEmailValidationParams) (*BulkEmailValidationJobResponse, error) {
	if params == nil {
		return nil, newValidationError("create bulk email validation params are required")
	}
	if params.Filename == "" {
		return nil, newValidationError("filename is required")
	}
	if params.Content == nil {
		return nil, newValidationError("file content is required")
	}

	response := &BulkEmailValidationJobResponse{}
	if err := s.client.http.doMultipart(ctx, http.MethodPost, "/v1/verify/files", nil, "file", params.Filename, params.Content, response); err != nil {
		return nil, err
	}
	return response, nil
}

// CreateBulkFromFile creates a bulk email validation job from a local file.
func (s *VerificationService) CreateBulkFromFile(ctx context.Context, path string) (*BulkEmailValidationJobResponse, error) {
	if path == "" {
		return nil, newValidationError("file path is required")
	}

	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()

	return s.CreateBulk(ctx, &CreateBulkEmailValidationParams{
		Filename: filepath.Base(path),
		Content:  file,
	})
}

// ListBulk returns cursor-paginated bulk email validation jobs.
func (s *VerificationService) ListBulk(ctx context.Context, params *ListBulkEmailValidationsParams) (*ListBulkEmailValidationsResponse, error) {
	query := make(url.Values)
	if params != nil {
		if params.Limit != nil {
			query.Set("limit", intToString(*params.Limit))
		}
		if params.Cursor != nil {
			query.Set("cursor", *params.Cursor)
		}
		if params.Search != nil {
			query.Set("search", *params.Search)
		}
		if params.Status != nil {
			query.Set("status", *params.Status)
		}
	}

	response := &ListBulkEmailValidationsResponse{}
	if err := s.client.http.do(ctx, http.MethodGet, "/v1/verify/files", query, nil, response); err != nil {
		return nil, err
	}
	return response, nil
}

// GetBulk fetches a bulk email validation job by ID.
func (s *VerificationService) GetBulk(ctx context.Context, jobID string) (*BulkEmailValidationJobResponse, error) {
	if jobID == "" {
		return nil, newValidationError("job id is required")
	}

	response := &BulkEmailValidationJobResponse{}
	path := "/v1/verify/files/" + url.PathEscape(jobID)
	if err := s.client.http.do(ctx, http.MethodGet, path, nil, nil, response); err != nil {
		return nil, err
	}
	return response, nil
}

// ContinueBulk continues a queued or action-required bulk email validation job.
func (s *VerificationService) ContinueBulk(ctx context.Context, jobID string) (*BulkEmailValidationJobResponse, error) {
	if jobID == "" {
		return nil, newValidationError("job id is required")
	}

	response := &BulkEmailValidationJobResponse{}
	path := "/v1/verify/files/" + url.PathEscape(jobID) + "/continue"
	if err := s.client.http.do(ctx, http.MethodPost, path, nil, nil, response); err != nil {
		return nil, err
	}
	return response, nil
}

// DownloadBulk downloads completed bulk email validation results.
func (s *VerificationService) DownloadBulk(ctx context.Context, jobID string, params *DownloadBulkEmailValidationParams) (*BulkEmailValidationDownload, error) {
	if jobID == "" {
		return nil, newValidationError("job id is required")
	}

	query := make(url.Values)
	if params != nil {
		if params.Filter != nil {
			query.Set("filter", *params.Filter)
		}
		if params.Format != nil {
			query.Set("format", *params.Format)
		}
	}

	path := "/v1/verify/files/" + url.PathEscape(jobID) + "/download"
	content, headers, err := s.client.http.doRaw(ctx, http.MethodGet, path, query, nil)
	if err != nil {
		return nil, err
	}

	return &BulkEmailValidationDownload{
		Content:     content,
		ContentType: headers.Get("Content-Type"),
		Filename:    filenameFromContentDisposition(headers.Get("Content-Disposition")),
	}, nil
}

// DeleteBulk deletes a bulk email validation job.
func (s *VerificationService) DeleteBulk(ctx context.Context, jobID string) (*DeleteBulkEmailValidationResponse, error) {
	if jobID == "" {
		return nil, newValidationError("job id is required")
	}

	response := &DeleteBulkEmailValidationResponse{}
	path := "/v1/verify/files/" + url.PathEscape(jobID)
	if err := s.client.http.do(ctx, http.MethodDelete, path, nil, nil, response); err != nil {
		return nil, err
	}
	return response, nil
}

// GetCredits returns the current email validation credit balance.
func (s *VerificationService) GetCredits(ctx context.Context) (*VerificationCreditsResponse, error) {
	response := &VerificationCreditsResponse{}
	if err := s.client.http.do(ctx, http.MethodGet, "/v1/verification-credits", nil, nil, response); err != nil {
		return nil, err
	}
	return response, nil
}

// ListCreditLedger returns cursor-paginated email validation credit ledger entries.
func (s *VerificationService) ListCreditLedger(ctx context.Context, params *ListVerificationCreditLedgerParams) (*VerificationCreditLedgerResponse, error) {
	query := make(url.Values)
	if params != nil {
		if params.Limit != nil {
			query.Set("limit", intToString(*params.Limit))
		}
		if params.Cursor != nil {
			query.Set("cursor", *params.Cursor)
		}
	}

	response := &VerificationCreditLedgerResponse{}
	if err := s.client.http.do(ctx, http.MethodGet, "/v1/verification-credits/ledger", query, nil, response); err != nil {
		return nil, err
	}
	return response, nil
}

func filenameFromContentDisposition(value string) string {
	if value == "" {
		return ""
	}
	_, params, err := mime.ParseMediaType(value)
	if err != nil {
		return ""
	}
	return params["filename"]
}
