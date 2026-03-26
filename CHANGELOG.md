# Changelog

## [Unreleased]

### ⚠ BREAKING CHANGES

* align schema-backed response/types with current OpenAPI:
  * `ListContactsResponse` now uses `data` (was `contacts`)
  * `Campaigns.Send` now returns `(*SendCampaignResponse, error)` instead of `error`

### Bug Fixes

* parse `GET /contacts` using `{ data, cursor, hasMore, total }`
* add `TemplatesService.List` with `GET /templates` pagination shape `{ data, total, page, pageSize, totalPages }`
* parse `POST /campaigns/{id}/send` as `{ success, data, message }`
* update schema models: `Contact` (`status`, `expiresAt`, `projectId`, nullable `data`), `Template` (expanded fields), `Segment` (`type`, nullable `condition`, `projectId`, timestamps)

## [1.1.0](https://github.com/MailGlyph/mailglyph-go/compare/v1.0.0...v1.1.0) (2026-03-10)


### Features

* Add `text` field to email sending parameters. ([0abb034](https://github.com/MailGlyph/mailglyph-go/commit/0abb034084556321e254669b48a359ad17160d39))
* Implement methods to add and remove contacts from static segments, along with supporting types and documentation. ([dfaba72](https://github.com/MailGlyph/mailglyph-go/commit/dfaba72558e00fbad9df277ac014a2f39f7ff757))
* Load environment variables from a `.env` file for integration tests. ([7710d4b](https://github.com/MailGlyph/mailglyph-go/commit/7710d4b00d5d3e921a18dafa04b03e955398703e))
* Rebranding ([4f17d2b](https://github.com/MailGlyph/mailglyph-go/commit/4f17d2beb02e7924c3908b7952e6b6095c4e7ae8))

## 1.0.0 (2026-02-16)


### ⚠ BREAKING CHANGES

* complete API redesign, not compatible with 0.1.x

### Features

* initial SDK implementation ([6ad353d](https://github.com/MailGlyph/mailglyph-go/commit/6ad353de099c190a81f8bd2545c666eb114acf6b))


### Bug Fixes

* align golangci config with CI runner ([c6d7888](https://github.com/MailGlyph/mailglyph-go/commit/c6d788880d700899e673e4248e1a0d2f772667b1))
