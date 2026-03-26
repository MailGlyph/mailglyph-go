package mailglyph

import (
	"net/http"
	"time"
)

const (
	defaultBaseURL = "https://api.mailglyph.com"
	defaultTimeout = 30 * time.Second
)

// Option configures a MailGlyph client.
type Option func(*clientOptions)

type clientOptions struct {
	baseURL    string
	timeout    time.Duration
	httpClient *http.Client
}

// Client provides access to MailGlyph API resources.
type Client struct {
	config ClientConfig
	http   *httpClient

	Emails    *EmailsService
	Events    *EventsService
	Contacts  *ContactsService
	Templates *TemplatesService
	Campaigns *CampaignsService
	Segments  *SegmentsService
}

// New creates a new MailGlyph client.
func New(apiKey string, opts ...Option) *Client {
	options := &clientOptions{
		baseURL: defaultBaseURL,
		timeout: defaultTimeout,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(options)
		}
	}

	cfg := ClientConfig{
		APIKey:  apiKey,
		BaseURL: options.baseURL,
		Timeout: options.timeout,
	}

	transport := newHTTPClient(cfg, options.httpClient)
	client := &Client{
		config: cfg,
		http:   transport,
	}
	client.Emails = &EmailsService{client: client}
	client.Events = &EventsService{client: client}
	client.Contacts = &ContactsService{client: client}
	client.Templates = &TemplatesService{client: client}
	client.Campaigns = &CampaignsService{client: client}
	client.Segments = &SegmentsService{client: client}

	return client
}

// NewClient creates a new MailGlyph client.
func NewClient(apiKey string, opts ...Option) *Client {
	return New(apiKey, opts...)
}

// WithBaseURL overrides the default API base URL.
func WithBaseURL(baseURL string) Option {
	return func(options *clientOptions) {
		if baseURL != "" {
			options.baseURL = baseURL
		}
	}
}

// WithTimeout overrides the default request timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(options *clientOptions) {
		if timeout > 0 {
			options.timeout = timeout
		}
	}
}

// WithHTTPClient sets a custom http.Client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(options *clientOptions) {
		if httpClient != nil {
			options.httpClient = httpClient
		}
	}
}
