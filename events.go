package mailglyph

import "context"

// EventsService provides methods for event endpoints.
type EventsService struct {
	client *Client
}

// Track tracks an event for a contact.
func (s *EventsService) Track(ctx context.Context, params *TrackEventParams) (*TrackEventResponse, error) {
	if params == nil {
		return nil, newValidationError("track event params are required")
	}
	if params.Email == "" {
		return nil, newValidationError("email is required")
	}
	if params.Event == "" {
		return nil, newValidationError("event is required")
	}

	response := &TrackEventResponse{}
	if err := s.client.http.do(ctx, "POST", "/v1/track", nil, params, response); err != nil {
		return nil, err
	}
	return response, nil
}

// GetNames lists unique tracked event names.
func (s *EventsService) GetNames(ctx context.Context) (*EventNamesResponse, error) {
	response := &EventNamesResponse{}
	if err := s.client.http.do(ctx, "GET", "/events/names", nil, nil, response); err != nil {
		return nil, err
	}
	return response, nil
}

// ListNames lists unique tracked event names.
func (s *EventsService) ListNames(ctx context.Context) (*EventNamesResponse, error) {
	return s.GetNames(ctx)
}
