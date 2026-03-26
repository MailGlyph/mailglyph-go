package mailglyph

import (
	"context"
	"net/url"
)

// TemplatesService provides methods for template endpoints.
type TemplatesService struct {
	client *Client
}

// List returns paginated templates.
func (s *TemplatesService) List(ctx context.Context, params *ListTemplatesParams) (*ListTemplatesResponse, error) {
	query := make(url.Values)
	if params != nil {
		if params.Limit != nil {
			query.Set("limit", intToString(*params.Limit))
		}
		if params.Cursor != nil {
			query.Set("cursor", *params.Cursor)
		}
		if params.Type != nil {
			query.Set("type", *params.Type)
		}
		if params.Search != nil {
			query.Set("search", *params.Search)
		}
	}

	response := &ListTemplatesResponse{}
	if err := s.client.http.do(ctx, "GET", "/templates", query, nil, response); err != nil {
		return nil, err
	}
	return response, nil
}
