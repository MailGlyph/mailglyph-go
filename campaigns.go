package mailglyph

import (
	"context"
	"net/url"
)

// CampaignsService provides methods for campaign endpoints.
type CampaignsService struct {
	client *Client
}

// List returns paginated campaigns.
func (s *CampaignsService) List(ctx context.Context, params *ListCampaignsParams) (*ListCampaignsResponse, error) {
	query := make(url.Values)
	if params != nil {
		if params.Page != nil {
			query.Set("page", intToString(*params.Page))
		}
		if params.PageSize != nil {
			query.Set("pageSize", intToString(*params.PageSize))
		}
		if params.Status != nil {
			query.Set("status", *params.Status)
		}
	}

	response := &ListCampaignsResponse{}
	if err := s.client.http.do(ctx, "GET", "/campaigns", query, nil, response); err != nil {
		return nil, err
	}
	return response, nil
}

// Create creates a campaign.
func (s *CampaignsService) Create(ctx context.Context, params *CreateCampaignParams) (*Campaign, error) {
	if params == nil {
		return nil, newValidationError("create campaign params are required")
	}
	if params.Name == "" {
		return nil, newValidationError("name is required")
	}
	if params.Subject == "" {
		return nil, newValidationError("subject is required")
	}
	if params.Body == "" {
		return nil, newValidationError("body is required")
	}
	if params.From == "" {
		return nil, newValidationError("from is required")
	}
	if params.AudienceType == "" {
		return nil, newValidationError("audienceType is required")
	}

	env := &CampaignEnvelope{}
	if err := s.client.http.do(ctx, "POST", "/campaigns", nil, params, env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// Get fetches a campaign by ID.
func (s *CampaignsService) Get(ctx context.Context, id string) (*Campaign, error) {
	if id == "" {
		return nil, newValidationError("id is required")
	}

	env := &CampaignEnvelope{}
	path := "/campaigns/" + url.PathEscape(id)
	if err := s.client.http.do(ctx, "GET", path, nil, nil, env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// Update updates a campaign by ID.
func (s *CampaignsService) Update(ctx context.Context, id string, params *UpdateCampaignParams) (*Campaign, error) {
	if id == "" {
		return nil, newValidationError("id is required")
	}
	if params == nil {
		return nil, newValidationError("update campaign params are required")
	}

	env := &CampaignEnvelope{}
	path := "/campaigns/" + url.PathEscape(id)
	if err := s.client.http.do(ctx, "PUT", path, nil, params, env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// Send sends or schedules a campaign.
func (s *CampaignsService) Send(ctx context.Context, id string, params *SendCampaignParams) error {
	if id == "" {
		return newValidationError("id is required")
	}

	path := "/campaigns/" + url.PathEscape(id) + "/send"
	var payload interface{}
	if params != nil {
		payload = params
	}
	return s.client.http.do(ctx, "POST", path, nil, payload, nil)
}

// Cancel cancels a scheduled campaign.
func (s *CampaignsService) Cancel(ctx context.Context, id string) (*CancelCampaignResponse, error) {
	if id == "" {
		return nil, newValidationError("id is required")
	}

	response := &CancelCampaignResponse{}
	path := "/campaigns/" + url.PathEscape(id) + "/cancel"
	if err := s.client.http.do(ctx, "POST", path, nil, nil, response); err != nil {
		return nil, err
	}
	return response, nil
}

// Test sends a test email for a campaign.
func (s *CampaignsService) Test(ctx context.Context, id, email string) (*TestCampaignResponse, error) {
	if id == "" {
		return nil, newValidationError("id is required")
	}
	if email == "" {
		return nil, newValidationError("email is required")
	}

	response := &TestCampaignResponse{}
	path := "/campaigns/" + url.PathEscape(id) + "/test"
	payload := map[string]string{"email": email}
	if err := s.client.http.do(ctx, "POST", path, nil, payload, response); err != nil {
		return nil, err
	}
	return response, nil
}

// Stats returns analytics for a campaign.
func (s *CampaignsService) Stats(ctx context.Context, id string) (*CampaignStatsResponse, error) {
	if id == "" {
		return nil, newValidationError("id is required")
	}

	response := &CampaignStatsResponse{}
	path := "/campaigns/" + url.PathEscape(id) + "/stats"
	if err := s.client.http.do(ctx, "GET", path, nil, nil, response); err != nil {
		return nil, err
	}
	return response, nil
}
