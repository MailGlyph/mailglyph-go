package mailglyph

import (
	"context"
	"net/url"
)

// ContactsService provides methods for contact endpoints.
type ContactsService struct {
	client *Client
}

// List returns cursor-paginated contacts.
func (s *ContactsService) List(ctx context.Context, params *ListContactsParams) (*ListContactsResponse, error) {
	query := make(url.Values)
	if params != nil {
		if params.Limit != nil {
			query.Set("limit", intToString(*params.Limit))
		}
		if params.Cursor != nil {
			query.Set("cursor", *params.Cursor)
		}
		if params.Subscribed != nil {
			query.Set("subscribed", boolToString(*params.Subscribed))
		}
		if params.Search != nil {
			query.Set("search", *params.Search)
		}
	}

	response := &ListContactsResponse{}
	if err := s.client.http.do(ctx, "GET", "/contacts", query, nil, response); err != nil {
		return nil, err
	}
	return response, nil
}

// Create creates or updates a contact by email.
func (s *ContactsService) Create(ctx context.Context, params *CreateContactParams) (*ContactUpsertResponse, error) {
	if params == nil {
		return nil, newValidationError("create contact params are required")
	}
	if params.Email == "" {
		return nil, newValidationError("email is required")
	}

	response := &ContactUpsertResponse{}
	if err := s.client.http.do(ctx, "POST", "/contacts", nil, params, response); err != nil {
		return nil, err
	}
	return response, nil
}

// Get fetches a contact by ID.
func (s *ContactsService) Get(ctx context.Context, id string) (*Contact, error) {
	if id == "" {
		return nil, newValidationError("id is required")
	}

	response := &Contact{}
	path := "/contacts/" + url.PathEscape(id)
	if err := s.client.http.do(ctx, "GET", path, nil, nil, response); err != nil {
		return nil, err
	}
	return response, nil
}

// Update updates a contact by ID.
func (s *ContactsService) Update(ctx context.Context, id string, params *UpdateContactParams) (*Contact, error) {
	if id == "" {
		return nil, newValidationError("id is required")
	}
	if params == nil {
		return nil, newValidationError("update contact params are required")
	}

	response := &Contact{}
	path := "/contacts/" + url.PathEscape(id)
	if err := s.client.http.do(ctx, "PATCH", path, nil, params, response); err != nil {
		return nil, err
	}
	return response, nil
}

// Delete deletes a contact by ID.
func (s *ContactsService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return newValidationError("id is required")
	}

	path := "/contacts/" + url.PathEscape(id)
	return s.client.http.do(ctx, "DELETE", path, nil, nil, nil)
}

// Count returns the total number of contacts for optional filters.
func (s *ContactsService) Count(ctx context.Context, params *ListContactsParams) (int, error) {
	queryParams := &ListContactsParams{}
	if params != nil {
		queryParams.Subscribed = params.Subscribed
		queryParams.Search = params.Search
	}
	limit := 1
	queryParams.Limit = &limit

	response, err := s.List(ctx, queryParams)
	if err != nil {
		return 0, err
	}
	if response.Total != nil {
		return *response.Total, nil
	}
	return len(response.Contacts), nil
}
