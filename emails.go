package mailrify

import "context"

// EmailsService provides methods for email endpoints.
type EmailsService struct {
	client *Client
}

// Send queues a transactional email.
func (s *EmailsService) Send(ctx context.Context, params *SendEmailParams) (*SendEmailResponse, error) {
	if params == nil {
		return nil, newValidationError("send params are required")
	}

	response := &SendEmailResponse{}
	if err := s.client.http.do(ctx, "POST", "/v1/send", nil, params, response); err != nil {
		return nil, err
	}
	return response, nil
}

// Verify checks an email address for validity and quality signals.
func (s *EmailsService) Verify(ctx context.Context, email string) (*VerifyEmailResponse, error) {
	if email == "" {
		return nil, newValidationError("email is required")
	}

	response := &VerifyEmailResponse{}
	if err := s.client.http.do(ctx, "POST", "/v1/verify", nil, &VerifyEmailParams{Email: email}, response); err != nil {
		return nil, err
	}
	return response, nil
}
