package mailglyph

import (
	"context"
	"net/url"
)

// SegmentsService provides methods for segment endpoints.
type SegmentsService struct {
	client *Client
}

// List returns all segments.
func (s *SegmentsService) List(ctx context.Context) ([]Segment, error) {
	response := make([]Segment, 0)
	if err := s.client.http.do(ctx, "GET", "/segments", nil, nil, &response); err != nil {
		return nil, err
	}
	return response, nil
}

// Create creates a new segment.
func (s *SegmentsService) Create(ctx context.Context, params *CreateSegmentParams) (*Segment, error) {
	if params == nil {
		return nil, newValidationError("create segment params are required")
	}
	if params.Name == "" {
		return nil, newValidationError("name is required")
	}
	if params.Condition == nil {
		return nil, newValidationError("condition is required")
	}

	response := &Segment{}
	if err := s.client.http.do(ctx, "POST", "/segments", nil, params, response); err != nil {
		return nil, err
	}
	return response, nil
}

// Get fetches a segment by ID.
func (s *SegmentsService) Get(ctx context.Context, id string) (*Segment, error) {
	if id == "" {
		return nil, newValidationError("id is required")
	}

	response := &Segment{}
	path := "/segments/" + url.PathEscape(id)
	if err := s.client.http.do(ctx, "GET", path, nil, nil, response); err != nil {
		return nil, err
	}
	return response, nil
}

// Update updates a segment by ID.
func (s *SegmentsService) Update(ctx context.Context, id string, params *UpdateSegmentParams) (*Segment, error) {
	if id == "" {
		return nil, newValidationError("id is required")
	}
	if params == nil {
		return nil, newValidationError("update segment params are required")
	}

	response := &Segment{}
	path := "/segments/" + url.PathEscape(id)
	if err := s.client.http.do(ctx, "PATCH", path, nil, params, response); err != nil {
		return nil, err
	}
	return response, nil
}

// Delete deletes a segment by ID.
func (s *SegmentsService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return newValidationError("id is required")
	}

	path := "/segments/" + url.PathEscape(id)
	return s.client.http.do(ctx, "DELETE", path, nil, nil, nil)
}

// ListContacts returns page-based contacts for a segment.
func (s *SegmentsService) ListContacts(ctx context.Context, id string, params *ListSegmentContactsParams) (*ListSegmentContactsResponse, error) {
	if id == "" {
		return nil, newValidationError("id is required")
	}

	query := make(url.Values)
	if params != nil {
		if params.Page != nil {
			query.Set("page", intToString(*params.Page))
		}
		if params.PageSize != nil {
			query.Set("pageSize", intToString(*params.PageSize))
		}
	}

	response := &ListSegmentContactsResponse{}
	path := "/segments/" + url.PathEscape(id) + "/contacts"
	if err := s.client.http.do(ctx, "GET", path, query, nil, response); err != nil {
		return nil, err
	}
	return response, nil
}

// AddMembers adds contacts to a static segment by email.
func (s *SegmentsService) AddMembers(ctx context.Context, id string, params *StaticSegmentMembersParams) (*AddStaticSegmentMembersResponse, error) {
	if id == "" {
		return nil, newValidationError("id is required")
	}
	if params == nil {
		return nil, newValidationError("add members params are required")
	}
	if len(params.Emails) == 0 {
		return nil, newValidationError("emails is required")
	}

	response := &AddStaticSegmentMembersResponse{}
	path := "/segments/" + url.PathEscape(id) + "/members"
	if err := s.client.http.do(ctx, "POST", path, nil, params, response); err != nil {
		return nil, err
	}
	return response, nil
}

// RemoveMembers removes contacts from a static segment by email.
func (s *SegmentsService) RemoveMembers(ctx context.Context, id string, params *StaticSegmentMembersParams) (*RemoveStaticSegmentMembersResponse, error) {
	if id == "" {
		return nil, newValidationError("id is required")
	}
	if params == nil {
		return nil, newValidationError("remove members params are required")
	}
	if len(params.Emails) == 0 {
		return nil, newValidationError("emails is required")
	}

	response := &RemoveStaticSegmentMembersResponse{}
	path := "/segments/" + url.PathEscape(id) + "/members"
	if err := s.client.http.do(ctx, "DELETE", path, nil, params, response); err != nil {
		return nil, err
	}
	return response, nil
}
