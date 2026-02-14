# Mailrify Go SDK Plan

> Shared spec: [sdk-plan.md](./sdk-plan.md) · Repo: [Mailrify/mailrify-go](https://github.com/Mailrify/mailrify-go) · Registry: Go modules · Min: Go 1.21+

---

## Tech Stack

| Concern | Choice |
|---------|--------|
| Language | Go 1.21+ |
| HTTP | `net/http` (stdlib) |
| JSON | `encoding/json` (stdlib) |
| Testing | `testing` (stdlib) + `testify` |
| Linting | `golangci-lint` |

---

## Repository Structure

```
mailrify-go/
├── mailrify.go               # Client struct + NewClient()
├── http.go                    # HTTP transport
├── errors.go                  # Error types
├── types.go                   # All struct definitions
├── emails.go                  # EmailsService
├── events.go                  # EventsService
├── contacts.go                # ContactsService
├── campaigns.go               # CampaignsService
├── segments.go                # SegmentsService
├── mailrify_test.go           # Client tests
├── emails_test.go
├── events_test.go
├── contacts_test.go
├── campaigns_test.go
├── segments_test.go
├── testutil_test.go           # Shared test helpers / mock server
├── integration_test.go        # Build-tag gated: //go:build integration
├── openapi.json
├── go.mod
├── go.sum
├── .github/
│   └── workflows/
│       ├── ci.yml
│       └── release-please.yml
├── release-please-config.json
├── .release-please-manifest.json
├── .golangci.yml
├── AGENTS.md
├── README.md
├── LICENSE
└── CHANGELOG.md
```

---

## Key Types (`types.go`)

```go
package mailrify

type ClientConfig struct {
    APIKey  string
    BaseURL string // default: https://api.mailrify.com
    Timeout time.Duration // default: 30s
}

// Emails
type SendEmailParams struct {
    To          interface{}            `json:"to"`          // string | Recipient | []Recipient
    From        interface{}            `json:"from"`        // string | Sender
    Subject     string                 `json:"subject,omitempty"`
    Body        string                 `json:"body,omitempty"`
    Template    string                 `json:"template,omitempty"`
    Data        map[string]interface{} `json:"data,omitempty"`
    Headers     map[string]string      `json:"headers,omitempty"`
    Reply       string                 `json:"reply,omitempty"`
    Name        string                 `json:"name,omitempty"`
    Subscribed  *bool                  `json:"subscribed,omitempty"`
    Attachments []Attachment           `json:"attachments,omitempty"`
}

type Recipient struct {
    Name  string `json:"name,omitempty"`
    Email string `json:"email"`
}

type Attachment struct {
    Filename    string `json:"filename"`
    Content     string `json:"content"`
    ContentType string `json:"contentType"`
}

type VerifyEmailResult struct {
    Email          string   `json:"email"`
    Valid          bool     `json:"valid"`
    IsDisposable   bool     `json:"isDisposable"`
    IsAlias        bool     `json:"isAlias"`
    IsTypo         bool     `json:"isTypo"`
    IsPlusAddressed bool    `json:"isPlusAddressed"`
    IsRandomInput  bool     `json:"isRandomInput"`
    IsPersonalEmail bool    `json:"isPersonalEmail"`
    DomainExists   bool     `json:"domainExists"`
    HasWebsite     bool     `json:"hasWebsite"`
    HasMxRecords   bool     `json:"hasMxRecords"`
    SuggestedEmail *string  `json:"suggestedEmail,omitempty"`
    Reasons        []string `json:"reasons"`
}

// Events
type TrackEventParams struct {
    Email      string                 `json:"email"`
    Event      string                 `json:"event"`
    Subscribed *bool                  `json:"subscribed,omitempty"`
    Data       map[string]interface{} `json:"data,omitempty"`
}

// Contacts
type Contact struct {
    ID         string                 `json:"id"`
    Email      string                 `json:"email"`
    Subscribed bool                   `json:"subscribed"`
    Data       map[string]interface{} `json:"data"`
    CreatedAt  string                 `json:"createdAt"`
    UpdatedAt  string                 `json:"updatedAt"`
}

type ListContactsParams struct {
    Limit      int    `url:"limit,omitempty"`
    Cursor     string `url:"cursor,omitempty"`
    Subscribed *bool  `url:"subscribed,omitempty"`
    Search     string `url:"search,omitempty"`
}

// Segments
type Segment struct {
    ID              string                 `json:"id"`
    Name            string                 `json:"name"`
    Description     *string                `json:"description,omitempty"`
    Condition       map[string]interface{} `json:"condition"`
    TrackMembership bool                   `json:"trackMembership"`
    MemberCount     int                    `json:"memberCount"`
}

// Campaigns
type Campaign struct {
    ID          string  `json:"id"`
    Name        string  `json:"name"`
    Subject     string  `json:"subject"`
    Type        string  `json:"type"`        // ALL | SEGMENT | FILTERED
    Status      string  `json:"status"`      // DRAFT | SCHEDULED | SENDING | SENT
    ScheduledAt *string `json:"scheduledAt,omitempty"`
}

type ListCampaignsParams struct {
    Limit  int    `url:"limit,omitempty"`
    Cursor string `url:"cursor,omitempty"`
    Status string `url:"status,omitempty"`
}

type CreateCampaignParams struct {
    Name           string                 `json:"name"`
    Subject        string                 `json:"subject"`
    Body           string                 `json:"body"`
    From           string                 `json:"from"`
    AudienceType   string                 `json:"audienceType"`
    Description    string                 `json:"description,omitempty"`
    FromName       string                 `json:"fromName,omitempty"`
    ReplyTo        string                 `json:"replyTo,omitempty"`
    SegmentID      string                 `json:"segmentId,omitempty"`
    AudienceFilter map[string]interface{} `json:"audienceFilter,omitempty"`
}

type UpdateCampaignParams struct {
    Name              string                 `json:"name,omitempty"`
    Description       string                 `json:"description,omitempty"`
    Subject           string                 `json:"subject,omitempty"`
    Body              string                 `json:"body,omitempty"`
    From              string                 `json:"from,omitempty"`
    FromName          string                 `json:"fromName,omitempty"`
    ReplyTo           string                 `json:"replyTo,omitempty"`
    AudienceType      string                 `json:"audienceType,omitempty"`
    SegmentID         string                 `json:"segmentId,omitempty"`
    AudienceCondition map[string]interface{} `json:"audienceCondition,omitempty"`
}

type SendCampaignParams struct {
    ScheduledFor *string `json:"scheduledFor,omitempty"` // ISO 8601
}
```

---

## API Design Pattern

```go
// Create client
client := mailrify.NewClient("sk_your_key")

// Resource methods return (result, error)
result, err := client.Emails.Send(ctx, params)
if err != nil {
    var apiErr *mailrify.ValidationError
    if errors.As(err, &apiErr) { ... }
}
```

All methods accept `context.Context` as first parameter for cancellation/timeout.

---

## Test Commands

| Scope | Command |
|-------|---------|
| Unit | `go test ./...` |
| Integration | `MAILRIFY_API_KEY=sk_... go test -tags=integration ./...` |
| Lint | `golangci-lint run` |
| Vet | `go vet ./...` |

---

## `go.mod`

```
module github.com/Mailrify/mailrify-go

go 1.21

require (
    github.com/stretchr/testify v1.9.0
)
```

---

## Usage Examples (for README)

```go
package main

import (
    "context"
    "fmt"
    "log"

    mailrify "github.com/Mailrify/mailrify-go"
)

func main() {
    client := mailrify.NewClient("sk_your_api_key")
    ctx := context.Background()

    // Send email
    result, err := client.Emails.Send(ctx, &mailrify.SendEmailParams{
        To:      "user@example.com",
        From:    &mailrify.Recipient{Name: "My App", Email: "hello@myapp.com"},
        Subject: "Welcome!",
        Body:    "<h1>Hello {{name}}</h1>",
        Data:    map[string]interface{}{"name": "John"},
    })

    // Verify email
    verification, err := client.Emails.Verify(ctx, "user@example.com")
    fmt.Println(verification.Data.Valid, verification.Data.IsRandomInput)

    // Track event (public key)
    tracker := mailrify.NewClient("pk_your_public_key")
    _, err = tracker.Events.Track(ctx, &mailrify.TrackEventParams{
        Email: "user@example.com",
        Event: "purchase",
        Data:  map[string]interface{}{"product": "Premium"},
    })

    // Contacts
    contacts, err := client.Contacts.List(ctx, &mailrify.ListContactsParams{Limit: 50})
    contact, err := client.Contacts.Create(ctx, &mailrify.CreateContactParams{
        Email: "new@example.com",
        Data:  map[string]interface{}{"plan": "premium"},
    })
    _, err = client.Contacts.Update(ctx, contact.ID, &mailrify.UpdateContactParams{
        Subscribed: boolPtr(false),
    })
    err = client.Contacts.Delete(ctx, contact.ID)

    // Segments
    segment, err := client.Segments.Create(ctx, &mailrify.CreateSegmentParams{
        Name: "Premium Users",
        Condition: map[string]interface{}{
            "operator":   "AND",
            "conditions": []map[string]interface{}{
                {"field": "data.plan", "operator": "equals", "value": "premium"},
            },
        },
        TrackMembership: true,
    })
    members, err := client.Segments.ListContacts(ctx, segment.ID, &mailrify.ListSegmentContactsParams{Page: 1})

    // Campaigns
    campaign, err := client.Campaigns.Create(ctx, &mailrify.CreateCampaignParams{
        Name:         "Product Launch",
        Subject:      "Introducing our new feature!",
        Body:         "<h1>Big news!</h1><p>Check out our latest feature.</p>",
        From:         "hello@myapp.com",
        AudienceType: "ALL",
    })

    // Schedule
    scheduledFor := "2026-03-01T10:00:00Z"
    _, err = client.Campaigns.Send(ctx, campaign.ID, &mailrify.SendCampaignParams{
        ScheduledFor: &scheduledFor,
    })

    // Test email
    _, err = client.Campaigns.Test(ctx, campaign.ID, "preview@myapp.com")

    // Stats
    stats, err := client.Campaigns.Stats(ctx, campaign.ID)

    // Cancel
    _, err = client.Campaigns.Cancel(ctx, campaign.ID)
}

func boolPtr(b bool) *bool { return &b }
```

---

## Release Automation

Go modules are versioned by **git tags only** — there's no registry publish step. release-please handles creating the tag and GitHub Release automatically.

### `release-please-config.json`

```json
{
  "$schema": "https://raw.githubusercontent.com/googleapis/release-please/main/schemas/config.json",
  "packages": {
    ".": {
      "release-type": "go",
      "bump-minor-pre-major": true,
      "bump-patch-for-minor-pre-major": true
    }
  }
}
```

### `.release-please-manifest.json`

```json
{
  ".": "0.1.0"
}
```

### `.github/workflows/release-please.yml`

```yaml
name: Release Please

on:
  push:
    branches: [main]

permissions:
  contents: write
  pull-requests: write

jobs:
  release-please:
    runs-on: ubuntu-latest
    steps:
      - uses: googleapis/release-please-action@v4
        with:
          release-type: go
```

> **Note:** Go consumers install via `go get github.com/Mailrify/mailrify-go@v0.2.0` — the GitHub Release tag is the version source of truth. No separate publish workflow is needed.
