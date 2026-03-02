package mailrify

import "time"

// Version is the SDK version used in the User-Agent header.
const Version = "0.2.0"

// ClientConfig defines client configuration.
type ClientConfig struct {
	APIKey  string
	BaseURL string
	Timeout time.Duration
}

// Recipient represents an email recipient or sender with optional display name.
type Recipient struct {
	Name  *string `json:"name,omitempty"`
	Email string  `json:"email"`
}

// Attachment represents an email attachment.
type Attachment struct {
	Filename    string `json:"filename"`
	Content     string `json:"content"`
	ContentType string `json:"contentType"`
}

// SendEmailParams contains fields for sending transactional emails.
type SendEmailParams struct {
	To      interface{} `json:"to"`
	From    interface{} `json:"from"`
	Subject *string     `json:"subject,omitempty"`
	Body    *string     `json:"body,omitempty"`
	// The plain text version of the message.
	// If not provided, the `body` will be used to generate a plain text version. You can opt out of this behavior by setting value to an empty string.
	Text        *string                `json:"text,omitempty"`
	Template    *string                `json:"template,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Headers     map[string]string      `json:"headers,omitempty"`
	Reply       *string                `json:"reply,omitempty"`
	Name        *string                `json:"name,omitempty"`
	Subscribed  *bool                  `json:"subscribed,omitempty"`
	Attachments []Attachment           `json:"attachments,omitempty"`
}

// VerifyEmailParams contains the email address to verify.
type VerifyEmailParams struct {
	Email string `json:"email"`
}

// VerifyEmailResult contains email verification analysis.
type VerifyEmailResult struct {
	Email           string   `json:"email"`
	Valid           bool     `json:"valid"`
	IsDisposable    bool     `json:"isDisposable"`
	IsAlias         bool     `json:"isAlias"`
	IsTypo          bool     `json:"isTypo"`
	IsPlusAddressed bool     `json:"isPlusAddressed"`
	IsRandomInput   bool     `json:"isRandomInput"`
	IsPersonalEmail bool     `json:"isPersonalEmail"`
	DomainExists    bool     `json:"domainExists"`
	HasWebsite      bool     `json:"hasWebsite"`
	HasMxRecords    bool     `json:"hasMxRecords"`
	SuggestedEmail  *string  `json:"suggestedEmail,omitempty"`
	Reasons         []string `json:"reasons"`
}

// SentEmailContact contains a contact reference for queued emails.
type SentEmailContact struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// SentEmail contains a queued email record.
type SentEmail struct {
	Contact SentEmailContact `json:"contact"`
	Email   string           `json:"email"`
}

// SendEmailResult contains send email response payload.
type SendEmailResult struct {
	Emails    []SentEmail `json:"emails"`
	Timestamp string      `json:"timestamp"`
}

// SendEmailResponse wraps send email responses.
type SendEmailResponse struct {
	Success bool            `json:"success"`
	Data    SendEmailResult `json:"data"`
}

// VerifyEmailResponse wraps verify email responses.
type VerifyEmailResponse struct {
	Success bool              `json:"success"`
	Data    VerifyEmailResult `json:"data"`
}

// TrackEventParams contains fields for tracking an event.
type TrackEventParams struct {
	Email      string                 `json:"email"`
	Event      string                 `json:"event"`
	Subscribed *bool                  `json:"subscribed,omitempty"`
	Data       map[string]interface{} `json:"data,omitempty"`
}

// TrackEventResult contains tracked event references.
type TrackEventResult struct {
	Contact   string `json:"contact"`
	Event     string `json:"event"`
	Timestamp string `json:"timestamp"`
}

// TrackEventResponse wraps event tracking responses.
type TrackEventResponse struct {
	Success bool             `json:"success"`
	Data    TrackEventResult `json:"data"`
}

// EventNamesResponse contains unique event names.
type EventNamesResponse struct {
	EventNames []string `json:"eventNames"`
}

// Contact contains a contact record.
type Contact struct {
	ID         string                 `json:"id"`
	Email      string                 `json:"email"`
	Subscribed bool                   `json:"subscribed"`
	Data       map[string]interface{} `json:"data"`
	CreatedAt  string                 `json:"createdAt"`
	UpdatedAt  string                 `json:"updatedAt"`
}

// ContactMeta contains metadata about upsert actions.
type ContactMeta struct {
	IsNew    bool `json:"isNew"`
	IsUpdate bool `json:"isUpdate"`
}

// ContactUpsertResponse is returned by create contact upserts.
type ContactUpsertResponse struct {
	Contact
	Meta ContactMeta `json:"_meta"`
}

// ListContactsParams controls list contacts filtering and pagination.
type ListContactsParams struct {
	Limit      *int    `json:"-"`
	Cursor     *string `json:"-"`
	Subscribed *bool   `json:"-"`
	Search     *string `json:"-"`
}

// ListContactsResponse contains cursor-paginated contacts.
type ListContactsResponse struct {
	Contacts []Contact `json:"contacts"`
	Cursor   *string   `json:"cursor"`
	HasMore  bool      `json:"hasMore"`
	Total    *int      `json:"total,omitempty"`
}

// CreateContactParams contains fields for creating a contact.
type CreateContactParams struct {
	Email      string                 `json:"email"`
	Subscribed *bool                  `json:"subscribed,omitempty"`
	Data       map[string]interface{} `json:"data,omitempty"`
}

// UpdateContactParams contains updatable contact fields.
type UpdateContactParams struct {
	Subscribed *bool                  `json:"subscribed,omitempty"`
	Data       map[string]interface{} `json:"data,omitempty"`
}

// Campaign contains a campaign record.
type Campaign struct {
	ID                string           `json:"id"`
	Name              string           `json:"name"`
	Description       *string          `json:"description,omitempty"`
	Subject           string           `json:"subject"`
	Body              string           `json:"body"`
	From              string           `json:"from"`
	FromName          *string          `json:"fromName,omitempty"`
	ReplyTo           *string          `json:"replyTo,omitempty"`
	AudienceType      string           `json:"audienceType"`
	AudienceCondition *FilterCondition `json:"audienceCondition,omitempty"`
	SegmentID         *string          `json:"segmentId,omitempty"`
	Status            string           `json:"status"`
	TotalRecipients   int              `json:"totalRecipients"`
	SentCount         int              `json:"sentCount"`
	DeliveredCount    int              `json:"deliveredCount"`
	OpenedCount       int              `json:"openedCount"`
	ClickedCount      int              `json:"clickedCount"`
	BouncedCount      int              `json:"bouncedCount"`
	ScheduledFor      *string          `json:"scheduledFor,omitempty"`
	SentAt            *string          `json:"sentAt,omitempty"`
	CreatedAt         string           `json:"createdAt"`
	UpdatedAt         string           `json:"updatedAt"`
	Segment           *Segment         `json:"segment,omitempty"`
}

// ListCampaignsParams controls campaign pagination and filtering.
type ListCampaignsParams struct {
	Page     *int    `json:"-"`
	PageSize *int    `json:"-"`
	Status   *string `json:"-"`
}

// ListCampaignsResponse contains paginated campaigns.
type ListCampaignsResponse struct {
	Data       []Campaign `json:"data"`
	Page       int        `json:"page"`
	PageSize   int        `json:"pageSize"`
	Total      int        `json:"total"`
	TotalPages int        `json:"totalPages"`
}

// CreateCampaignParams contains campaign creation fields.
type CreateCampaignParams struct {
	Name              string           `json:"name"`
	Subject           string           `json:"subject"`
	Body              string           `json:"body"`
	From              string           `json:"from"`
	AudienceType      string           `json:"audienceType"`
	Description       *string          `json:"description,omitempty"`
	FromName          *string          `json:"fromName,omitempty"`
	ReplyTo           *string          `json:"replyTo,omitempty"`
	SegmentID         *string          `json:"segmentId,omitempty"`
	AudienceCondition *FilterCondition `json:"audienceCondition,omitempty"`
}

// UpdateCampaignParams contains campaign update fields.
type UpdateCampaignParams struct {
	Name              *string          `json:"name,omitempty"`
	Description       *string          `json:"description,omitempty"`
	Subject           *string          `json:"subject,omitempty"`
	Body              *string          `json:"body,omitempty"`
	From              *string          `json:"from,omitempty"`
	FromName          *string          `json:"fromName,omitempty"`
	ReplyTo           *string          `json:"replyTo,omitempty"`
	AudienceType      *string          `json:"audienceType,omitempty"`
	SegmentID         *string          `json:"segmentId,omitempty"`
	AudienceCondition *FilterCondition `json:"audienceCondition,omitempty"`
}

// SendCampaignParams contains optional schedule data for sending campaigns.
type SendCampaignParams struct {
	ScheduledFor *string `json:"scheduledFor,omitempty"`
}

// CampaignEnvelope wraps common campaign responses.
type CampaignEnvelope struct {
	Success bool     `json:"success"`
	Data    Campaign `json:"data"`
}

// CancelCampaignResponse wraps campaign cancellation responses.
type CancelCampaignResponse struct {
	Success bool     `json:"success"`
	Data    Campaign `json:"data"`
	Message string   `json:"message"`
}

// TestCampaignResponse wraps campaign test-send responses.
type TestCampaignResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// CampaignStatsResponse wraps campaign stats responses.
type CampaignStatsResponse struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data"`
}

// Segment contains an audience segment record.
type Segment struct {
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	Description     *string          `json:"description,omitempty"`
	Condition       *FilterCondition `json:"condition,omitempty"`
	TrackMembership bool             `json:"trackMembership"`
	MemberCount     int              `json:"memberCount"`
}

// CreateSegmentParams contains segment creation fields.
type CreateSegmentParams struct {
	Name            string           `json:"name"`
	Description     *string          `json:"description,omitempty"`
	Condition       *FilterCondition `json:"condition"`
	TrackMembership *bool            `json:"trackMembership,omitempty"`
}

// UpdateSegmentParams contains segment update fields.
type UpdateSegmentParams struct {
	Name            *string          `json:"name,omitempty"`
	Description     *string          `json:"description,omitempty"`
	Condition       *FilterCondition `json:"condition,omitempty"`
	TrackMembership *bool            `json:"trackMembership,omitempty"`
}

// SegmentFilter is a single filter rule used in segment and campaign conditions.
type SegmentFilter struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value,omitempty"`
	Unit     *string     `json:"unit,omitempty"`
}

// FilterGroup combines filters and optional nested conditions.
type FilterGroup struct {
	Filters    []SegmentFilter  `json:"filters"`
	Conditions *FilterCondition `json:"conditions,omitempty"`
}

// FilterCondition is the shared logical condition tree used by segments and campaigns.
type FilterCondition struct {
	Logic  string        `json:"logic"`
	Groups []FilterGroup `json:"groups"`
}

// ListSegmentContactsParams controls page-based segment contacts pagination.
type ListSegmentContactsParams struct {
	Page     *int `json:"-"`
	PageSize *int `json:"-"`
}

// ListSegmentContactsResponse contains paginated contacts in a segment.
type ListSegmentContactsResponse struct {
	Data       []Contact `json:"data"`
	Total      int       `json:"total"`
	Page       int       `json:"page"`
	PageSize   int       `json:"pageSize"`
	TotalPages int       `json:"totalPages"`
}

// StaticSegmentMembersParams contains email addresses for static segment membership updates.
type StaticSegmentMembersParams struct {
	Emails []string `json:"emails"`
}

// AddStaticSegmentMembersResponse contains add-members results.
type AddStaticSegmentMembersResponse struct {
	Added    int      `json:"added"`
	NotFound []string `json:"notFound"`
}

// RemoveStaticSegmentMembersResponse contains remove-members results.
type RemoveStaticSegmentMembersResponse struct {
	Removed int `json:"removed"`
}
