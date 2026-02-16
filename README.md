# Mailrify Go SDK

[![CI](https://github.com/Mailrify/mailrify-go/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/Mailrify/mailrify-go/actions/workflows/ci.yml)
[![Release Please](https://github.com/Mailrify/mailrify-go/actions/workflows/release-please.yml/badge.svg?branch=main)](https://github.com/Mailrify/mailrify-go/actions/workflows/release-please.yml)

Official Go SDK for the Mailrify API.

## Install

```bash
go get github.com/Mailrify/mailrify-go
```

## Initialize

```go
package main

import (
    "time"

    mailrify "github.com/Mailrify/mailrify-go"
)

func main() {
    _ = mailrify.New("sk_your_api_key", mailrify.WithTimeout(10*time.Second))
}
```

## Usage

### Emails

```go
package main

import (
    "context"
    "fmt"

    mailrify "github.com/Mailrify/mailrify-go"
)

func main() {
    client := mailrify.New("sk_your_api_key")
    ctx := context.Background()

    subject := "Welcome!"
    body := "<h1>Hello {{name}}</h1>"

    sendResult, err := client.Emails.Send(ctx, &mailrify.SendEmailParams{
        To:      "user@example.com",
        From:    "hello@myapp.com",
        Subject: &subject,
        Body:    &body,
        Data: map[string]interface{}{
            "name": "John",
        },
    })
    if err != nil {
        panic(err)
    }
    fmt.Println(sendResult.Success)

    verifyResult, err := client.Emails.Verify(ctx, "user@example.com")
    if err != nil {
        panic(err)
    }
    fmt.Println(verifyResult.Data.Valid, verifyResult.Data.IsRandomInput)
}
```

### Events

```go
package main

import (
    "context"

    mailrify "github.com/Mailrify/mailrify-go"
)

func main() {
    ctx := context.Background()

    tracker := mailrify.New("pk_your_public_key")
    _, err := tracker.Events.Track(ctx, &mailrify.TrackEventParams{
        Email: "user@example.com",
        Event: "purchase",
        Data: map[string]interface{}{
            "product": "Premium",
        },
    })
    if err != nil {
        panic(err)
    }

    admin := mailrify.New("sk_your_secret_key")
    names, err := admin.Events.GetNames(ctx)
    if err != nil {
        panic(err)
    }
    _ = names
}
```

### Contacts

```go
package main

import (
    "context"

    mailrify "github.com/Mailrify/mailrify-go"
)

func main() {
    client := mailrify.New("sk_your_api_key")
    ctx := context.Background()

    contacts, err := client.Contacts.List(ctx, &mailrify.ListContactsParams{Limit: intPtr(50)})
    if err != nil {
        panic(err)
    }
    _ = contacts

    contact, err := client.Contacts.Create(ctx, &mailrify.CreateContactParams{
        Email: "new@example.com",
        Data: map[string]interface{}{"plan": "premium"},
    })
    if err != nil {
        panic(err)
    }

    fetched, err := client.Contacts.Get(ctx, contact.ID)
    if err != nil {
        panic(err)
    }
    _ = fetched

    _, err = client.Contacts.Update(ctx, contact.ID, &mailrify.UpdateContactParams{Subscribed: boolPtr(false)})
    if err != nil {
        panic(err)
    }

    total, err := client.Contacts.Count(ctx, nil)
    if err != nil {
        panic(err)
    }
    _ = total

    if err := client.Contacts.Delete(ctx, contact.ID); err != nil {
        panic(err)
    }
}

func boolPtr(v bool) *bool { return &v }
func intPtr(v int) *int    { return &v }
```

### Campaigns

```go
package main

import (
    "context"

    mailrify "github.com/Mailrify/mailrify-go"
)

func main() {
    client := mailrify.New("sk_your_api_key")
    ctx := context.Background()

    campaigns, err := client.Campaigns.List(ctx, &mailrify.ListCampaignsParams{
        Page: intPtr(1),
        PageSize: intPtr(20),
        Status: strPtr("DRAFT"),
    })
    if err != nil {
        panic(err)
    }
    _ = campaigns

    campaign, err := client.Campaigns.Create(ctx, &mailrify.CreateCampaignParams{
        Name:         "Product Launch",
        Subject:      "Introducing our new feature!",
        Body:         "<h1>Big news!</h1><p>Check out our latest feature.</p>",
        From:         "hello@myapp.com",
        AudienceType: "ALL",
    })
    if err != nil {
        panic(err)
    }

    scheduledFor := "2026-03-01T10:00:00Z"
    if err := client.Campaigns.Send(ctx, campaign.ID, &mailrify.SendCampaignParams{ScheduledFor: &scheduledFor}); err != nil {
        panic(err)
    }

    _, err = client.Campaigns.Test(ctx, campaign.ID, "preview@myapp.com")
    if err != nil {
        panic(err)
    }

    stats, err := client.Campaigns.Stats(ctx, campaign.ID)
    if err != nil {
        panic(err)
    }
    _ = stats

    _, err = client.Campaigns.Cancel(ctx, campaign.ID)
    if err != nil {
        panic(err)
    }
}

func intPtr(v int) *int { return &v }
func strPtr(v string) *string { return &v }
```

### Segments

```go
package main

import (
    "context"

    mailrify "github.com/Mailrify/mailrify-go"
)

func main() {
    client := mailrify.New("sk_your_api_key")
    ctx := context.Background()

    segment, err := client.Segments.Create(ctx, &mailrify.CreateSegmentParams{
        Name: "Premium Users",
        Condition: &mailrify.FilterCondition{
            Logic: "AND",
            Groups: []mailrify.FilterGroup{
                {
                    Filters: []mailrify.SegmentFilter{
                        {Field: "data.plan", Operator: "equals", Value: "premium"},
                    },
                },
            },
        },
        TrackMembership: boolPtr(true),
    })
    if err != nil {
        panic(err)
    }

    _, err = client.Segments.List(ctx)
    if err != nil {
        panic(err)
    }

    _, err = client.Segments.Get(ctx, segment.ID)
    if err != nil {
        panic(err)
    }

    _, err = client.Segments.Update(ctx, segment.ID, &mailrify.UpdateSegmentParams{Name: strPtr("Premium Users v2")})
    if err != nil {
        panic(err)
    }

    members, err := client.Segments.ListContacts(ctx, segment.ID, &mailrify.ListSegmentContactsParams{Page: intPtr(1), PageSize: intPtr(20)})
    if err != nil {
        panic(err)
    }
    _ = members

    if err := client.Segments.Delete(ctx, segment.ID); err != nil {
        panic(err)
    }
}

func boolPtr(v bool) *bool { return &v }
func intPtr(v int) *int    { return &v }
func strPtr(v string) *string { return &v }
```

## Error Handling

```go
package main

import (
    "context"
    "errors"

    mailrify "github.com/Mailrify/mailrify-go"
)

func main() {
    client := mailrify.New("sk_your_api_key")

    _, err := client.Contacts.Get(context.Background(), "missing")
    if err != nil {
        var notFound *mailrify.NotFoundError
        if errors.As(err, &notFound) {
            // Handle 404.
        }
    }
}
```
