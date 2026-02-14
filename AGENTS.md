# AGENTS.md

## Project
This is the official Mailrify Go SDK. See `./docs/` for the full specification.

Repository: https://github.com/Mailrify/mailrify-go
Module path: `github.com/Mailrify/mailrify-go`

## Context Files (read these FIRST)
1. [sdk-plan.md](./docs/sdk-plan.md) — Shared API spec, all 22 endpoints, auth rules, error hierarchy, testing strategy, release-please setup
2. [sdk-plan-go.md](./docs/sdk-plan-go.md) — Go-specific implementation plan (structure, struct types, workflows)
3. [openapi.json](./docs/openapi.json) — OpenAPI 3.0.3 specification (source of truth for schemas)

## Build Order
1. Scaffold: `go.mod` (module `github.com/Mailrify/mailrify-go`), `.gitignore`, `.golangci.yml`
2. Error types (`errors.go`) — `MailrifyError`, `AuthenticationError`, `ValidationError`, `NotFoundError`, `RateLimitError`, `ApiError`
3. Types (`types.go`) — all request/response structs: `Contact`, `Segment`, `Campaign`, `ListCampaignsParams`, `CreateCampaignParams`, etc.
4. HttpClient (`http.go`) — stdlib `net/http`, Bearer auth, JSON marshal/unmarshal, error mapping, retry with exponential backoff for 429/5xx
5. Client (`mailrify.go`) — `New(apiKey, ...Option)` constructor, exposes `.Emails`, `.Events`, `.Contacts`, `.Campaigns`, `.Segments` services. All methods take `context.Context` as first parameter
6. Resources one at a time with `testing` package tests:
   - Emails (Send, Verify) → tests
   - Events (Track with `pk_*` support, GetNames) → tests
   - Contacts (List, Create, Get, Update, Delete, Count) → tests
   - Campaigns (List, Create, Get, Update, Send, Cancel, Test, Stats) → tests
   - Segments (List, Create, Get, Update, Delete, ListContacts) → tests
7. CI: [.github/workflows/ci.yml](cci:7://file:///Users/sharo/Library/CloudStorage/Dropbox/Projects/Plunk/plunk/.github/workflows/ci.yml:0:0-0:0), `release-please.yml`
8. README with install + usage examples
9. Run `go test ./...` — all tests must pass

## Standards
- Go 1.21+, stdlib `net/http` only (no third-party HTTP client)
- `context.Context` as first parameter on every API method
- Functional options pattern for client configuration: `mailrify.New(key, WithTimeout(10*time.Second))`
- Pointer types for optional fields (`*string`, `*bool`)
- `httptest.NewServer` for unit test mocking (no real API calls)
- golangci-lint for linting
- Conventional Commits for all commit messages
- No separate publish workflow — Go modules use git tags via release-please
