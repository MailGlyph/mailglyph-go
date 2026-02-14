# Mailrify SDK Generation Plan — Overview

This is the shared specification for all Mailrify API SDKs. Each SDK has its own detailed plan:

| SDK | Plan | Repo | Registry |
|-----|------|------|----------|
| Node.js / TypeScript | [sdk-plan-node.md](./sdk-plan-node.md) | [mailrify-node](https://github.com/Mailrify/mailrify-node) | [npm `mailrify`](https://www.npmjs.com/package/mailrify) |
| Python | [sdk-plan-python.md](./sdk-plan-python.md) | [mailrify-python](https://github.com/Mailrify/mailrify-python) | [PyPI `mailrify`](https://pypi.org/project/mailrify/) |
| Go | [sdk-plan-go.md](./sdk-plan-go.md) | [mailrify-go](https://github.com/Mailrify/mailrify-go) | `go get github.com/Mailrify/mailrify-go` |
| PHP | [sdk-plan-php.md](./sdk-plan-php.md) | [mailrify-php](https://github.com/Mailrify/mailrify-php) | [Packagist `mailrify/mailrify-php`](https://packagist.org/packages/mailrify/mailrify-php) |

> Each SDK lives in its **own** GitHub repository — not in this monorepo.

Source of truth: [openapi.json](../apps/docs/openapi.json) (OpenAPI 3.0.3)

---

## API Endpoint Inventory (In-Scope)

### Emails (requires `sk_*` secret key)

| Method | Path | OperationId | SDK Method |
|--------|------|-------------|------------|
| `POST` | `/v1/send` | `sendEmail` | `emails.send(params)` |
| `POST` | `/v1/verify` | `verifyEmail` | `emails.verify(email)` |

**Send** request: `to` (string | object | array), `from` (string | object), `subject?`, `body?`, `template?`, `data?`, `headers?`, `reply?`, `attachments?[]`, `subscribed?`, `name?`

**Verify** request: `email` (required). Response includes: `valid`, `isDisposable`, `isAlias`, `isTypo`, `isPlusAddressed`, `isRandomInput`, `isPersonalEmail`, `domainExists`, `hasWebsite`, `hasMxRecords`, `suggestedEmail?`, `reasons[]`.

### Events

| Method | Path | OperationId | Auth | SDK Method |
|--------|------|-------------|------|------------|
| `POST` | `/v1/track` | `trackEvent` | `pk_*` only | `events.track(params)` |
| `GET` | `/events/names` | `listEventNames` | `sk_*` | `events.listNames()` |

- `/v1/track` is the **only** endpoint that accepts public keys. It does **not** accept secret keys.

### Contacts (requires `sk_*` secret key)

| Method | Path | OperationId | SDK Method |
|--------|------|-------------|------------|
| `GET` | `/contacts` | `listContacts` | `contacts.list(params?)` |
| `GET` | `/contacts/{id}` | `getContact` | `contacts.get(id)` |
| `POST` | `/contacts` | `createContact` | `contacts.create(params)` |
| `PATCH` | `/contacts/{id}` | `updateContact` | `contacts.update(id, params)` |
| `DELETE` | `/contacts/{id}` | `deleteContact` | `contacts.delete(id)` |

Pagination: Cursor-based (`cursor`, `limit`, `hasMore`, `total`). Filters: `subscribed`, `search`.

### Campaigns (requires `sk_*` secret key)

| Method | Path | OperationId | SDK Method |
|--------|------|-------------|------------|
| `GET` | `/campaigns` | `listCampaigns` | `campaigns.list(params?)` |
| `POST` | `/campaigns` | `createCampaign` | `campaigns.create(params)` |
| `GET` | `/campaigns/{id}` | `getCampaign` | `campaigns.get(id)` |
| `PUT` | `/campaigns/{id}` | `updateCampaign` | `campaigns.update(id, params)` |
| `POST` | `/campaigns/{id}/send` | `sendCampaign` | `campaigns.send(id, params?)` |
| `POST` | `/campaigns/{id}/cancel` | `cancelCampaign` | `campaigns.cancel(id)` |
| `POST` | `/campaigns/{id}/test` | `testCampaign` | `campaigns.test(id, email)` |
| `GET` | `/campaigns/{id}/stats` | `campaignStats` | `campaigns.stats(id)` |

Campaign statuses: `DRAFT`, `SCHEDULED`, `SENDING`, `SENT`. Audience types: `ALL`, `SEGMENT`, `FILTERED`.

Create requires: `name`, `subject`, `body`, `from`, `audienceType`. Optional: `description`, `fromName`, `replyTo`, `segmentId`, `audienceFilter`.

Send accepts optional `scheduledFor` (ISO 8601) — omit for immediate send.

### Segments (requires `sk_*` secret key)

| Method | Path | OperationId | SDK Method |
|--------|------|-------------|------------|
| `GET` | `/segments` | `listSegments` | `segments.list()` |
| `POST` | `/segments` | `createSegment` | `segments.create(params)` |
| `GET` | `/segments/{id}` | `getSegment` | `segments.get(id)` |
| `PATCH` | `/segments/{id}` | `updateSegment` | `segments.update(id, params)` |
| `DELETE` | `/segments/{id}` | `deleteSegment` | `segments.delete(id)` |
| `GET` | `/segments/{id}/contacts` | `listSegmentContacts` | `segments.listContacts(id, params?)` |

Segment contacts pagination: Page-based (`page`, `pageSize`, `total`, `totalPages`).

---

## Authentication

```
Authorization: Bearer sk_abc123   (secret key — all endpoints except /v1/track)
Authorization: Bearer pk_abc123   (public key — /v1/track ONLY)
```

Client constructor parameters:

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `apiKey` | **Yes** | — | `sk_*` or `pk_*` key |
| `baseUrl` | No | `https://api.mailrify.com` | Custom API base URL |
| `timeout` | No | `30000` (ms) | Request timeout |

Rules:
1. Auto-detect key type from prefix (`sk_` vs `pk_`)
2. Throw if `pk_*` used with non-track endpoint
3. Throw if `sk_*` used with `/v1/track`
4. Send `Content-Type: application/json`
5. Send `User-Agent: mailrify-{lang}/{version}`

---

## Architecture (All SDKs)

```
Client (config, auth) ──► emails.*
                       ──► events.*
                       ──► contacts.*
                       ──► campaigns.*
                       ──► segments.*
                       ──► HttpClient (request, retry, errors) ──► Mailrify API
```

| Component | Responsibility |
|-----------|---------------|
| **Client** | Entry point. Config. Resource namespaces. |
| **HttpClient** | Transport, auth injection, error parsing, retry. |
| **Resources** | `Emails`, `Events`, `Contacts`, `Campaigns`, `Segments` |
| **Types/Models** | Typed request/response objects from OpenAPI. |
| **Errors** | `MailrifyError` → `AuthenticationError`, `ValidationError`, `NotFoundError`, `RateLimitError`, `ApiError` |

### Error mapping

| HTTP | Error | Retry? |
|------|-------|--------|
| 400 | `ValidationError` | No |
| 401 | `AuthenticationError` | No |
| 404 | `NotFoundError` | No |
| 429 | `RateLimitError` | Yes |
| 5xx | `ApiError` | Yes |

Retry: max 3, exponential backoff `1s → 2s → 4s` with jitter, respect `Retry-After`.

---

## Testing Strategy (All SDKs)

### Unit test cases per resource

**Emails:**
1. `send()` — simple string `to`/`from`
2. `send()` — object `to`/`from` with names
3. `send()` — array of recipients
4. `send()` — template with data
5. `send()` — with attachments
6. `send()` — validation error (missing `to`)
7. `verify()` — valid email, full result including `isRandomInput`
8. `verify()` — typo with suggestion
9. `verify()` — validation error

**Events:**
1. `track()` — simple event
2. `track()` — with custom data
3. `track()` — works with `pk_*`
4. `track()` — rejects `sk_*`
5. `listNames()` — returns names array
6. `listNames()` — rejects `pk_*`

**Contacts:**
1. `list()` — paginated with cursor
2. `list()` — with filters
3. `get()` — single contact
4. `get()` — 404 → `NotFoundError`
5. `create()` — new (`_meta.isNew`)
6. `create()` — upsert (`_meta.isUpdate`)
7. `update()` — subscribed only
8. `update()` — custom data
9. `delete()` — 204 success
10. `delete()` — 404 → `NotFoundError`

**Campaigns:**
1. `list()` — returns paginated campaigns
2. `list()` — with status filter
3. `create()` — with required fields
4. `create()` — validation error (missing `subject`)
5. `get()` — by ID
6. `get()` — 404 → `NotFoundError`
7. `update()` — partial update
8. `send()` — immediate send (no `scheduledFor`)
9. `send()` — scheduled send with ISO date
10. `cancel()` — cancel scheduled campaign
11. `cancel()` — 404
12. `test()` — send test email
13. `test()` — validation error (missing email)
14. `stats()` — returns analytics object
15. `stats()` — 404

**Segments:**
1. `list()` — returns segments
2. `create()` — with conditions
3. `get()` — by ID
4. `get()` — 404
5. `update()` — conditions
6. `delete()` — 204
7. `listContacts()` — paginated
8. `listContacts()` — with params

**Cross-cutting:**
1. Bearer auth header
2. User-Agent header
3. 401/400/404/429/500 error mapping
4. Timeout handling
5. Custom base URL

### Integration tests (env-gated)

Require `MAILRIFY_API_KEY` and optionally `MAILRIFY_BASE_URL`:

1. Contact CRUD lifecycle
2. Campaign lifecycle: create → update → send test → schedule → cancel
3. Segment CRUD + list contacts
4. Track event → list names → verify presence
5. Verify a valid email
6. Send test email (optional, needs verified domain)

---

## Agent Build Order

```
1. Scaffold project
2. Implement error classes
3. Implement HttpClient
4. Implement Client
5. Resources (one at a time, with tests):
   a. Emails   → tests → verify
   b. Events   → tests → verify
   c. Contacts  → tests → verify
   d. Campaigns → tests → verify
   e. Segments  → tests → verify
6. README with usage examples
7. CI workflow
8. Run all tests, fix failures
```

### Agent validation checklist

- [ ] All 22 endpoints implemented
- [ ] All unit tests pass
- [ ] Type checking passes
- [ ] Linting passes
- [ ] README with install + usage per resource
- [ ] Package builds successfully

---

## CI/CD & Release Strategy (All SDKs)

### CI Workflow (`.github/workflows/ci.yml`)

Triggers on every push and PR to `main`:

1. **Lint** — language-specific linter
2. **Type Check** — static analysis
3. **Unit Tests** — with coverage
4. **Build** — verify package compiles

### Release Automation — [release-please](https://github.com/googleapis/release-please)

All SDKs use **release-please** (via GitHub Action) for fully automated versioning and changelog generation. No manual version bumps or changelog edits.

**How it works:**

1. Developers merge PRs with [Conventional Commits](https://www.conventionalcommits.org/) into `main`
2. release-please automatically opens/updates a **Release PR** that bumps the version and updates `CHANGELOG.md`
3. When the Release PR is merged, release-please creates a **GitHub Release** with auto-generated notes
4. The publish workflow triggers on the new release and pushes to the package registry

**Per-language strategy types:**

| SDK | release-please strategy | Version file(s) updated |
|-----|------------------------|------------------------|
| Node.js | `node` | `package.json` |
| Python | `python` | `pyproject.toml`, `src/mailrify/__init__.py` |
| Go | `go` | tag only (Go convention) |
| PHP | `php` | `composer.json` |

### GitHub Repositories

| SDK | Repository | Package Registry |
|-----|-----------|------------------|
| Node.js | https://github.com/Mailrify/mailrify-node | https://www.npmjs.com/package/mailrify |
| Python | https://github.com/Mailrify/mailrify-python | https://pypi.org/project/mailrify/ |
| Go | https://github.com/Mailrify/mailrify-go | `go get github.com/Mailrify/mailrify-go` |
| PHP | https://github.com/Mailrify/mailrify-php | https://packagist.org/packages/mailrify/mailrify-php |

> **Note:** These repos contain legacy code that will be replaced. Start with a clean `main` branch.

### Release-Please GitHub Action (`.github/workflows/release-please.yml`)

Every SDK repo must include this workflow:

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
    outputs:
      release_created: ${{ steps.release.outputs.release_created }}
      tag_name: ${{ steps.release.outputs.tag_name }}
    steps:
      - uses: googleapis/release-please-action@v4
        id: release
        with:
          release-type: node  # Change per SDK: node | python | go | php
```

### Publish Workflow (`.github/workflows/publish.yml`)

Triggers **only** when release-please creates a GitHub Release:

```yaml
name: Publish

on:
  release:
    types: [published]

jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      # ... language-specific build + publish steps
```

See each per-language plan for the specific publish steps.

### Conventional Commits Guide

All contributors must follow [Conventional Commits](https://www.conventionalcommits.org/):

| Prefix | Version Bump | Example |
|--------|-------------|---------|
| `fix:` | Patch (0.0.x) | `fix: handle null response in contacts.get` |
| `feat:` | Minor (0.x.0) | `feat: add campaigns resource` |
| `feat!:` or `BREAKING CHANGE:` | Major (x.0.0) | `feat!: rename send() to sendEmail()` |
| `chore:` | No release | `chore: update dev dependencies` |
| `docs:` | No release | `docs: add campaign examples to README` |
| `test:` | No release | `test: add integration tests for segments` |

### Per-SDK `release-please-config.json`

Each repo root should include:

```json
{
  "$schema": "https://raw.githubusercontent.com/googleapis/release-please/main/schemas/config.json",
  "packages": {
    ".": {
      "release-type": "node",
      "bump-minor-pre-major": true,
      "bump-patch-for-minor-pre-major": true
    }
  }
}
```

> `bump-minor-pre-major: true` ensures that during `0.x` development, `feat:` bumps patch instead of minor, preventing premature `1.0.0`.
