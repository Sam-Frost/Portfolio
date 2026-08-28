# CMS API

Backing API for the `domain-client` CMS page (`src/features/cms/`). Feature-sliced package at
`internal/cms/` (`model.go`, `repository.go` + `repository_memory.go` + `repository_postgres.go`,
`service.go`, `handler.go`, `publisher.go` + `publisher_s3.go`).

The CMS edits the content of the **public** portfolio site (`portfolio-client`): projects,
experience entries, blog posts ("writings"), and the home-page summary. Every edit is saved to
this package's Postgres tables as a **draft**. The public site does not read those tables — on
**Publish**, the draft is serialized to a single `content.json` and shipped to the public site's
S3 origin (+ a CloudFront invalidation). `portfolio-client` fetches `/content.json` at runtime,
with the values in `portfolio-client/src/content/fallback.ts` compiled in as an offline default.

"Unpublished changes" = the diff between the assembled draft and the last **successful**
publication's snapshot (per-item `updatedAt` and the envelope's `version`/`publishedAt` are
ignored, so re-saving identical values is not a change).

## Auth

Same gate as the rest of the domain area: every `/api/cms/*` route requires the bearer JWT
(`internal/auth`, enforced by `withAuth` in `cmd/main.go`).

## Publishing config (`cmd/main.go` → `newCMSPublisher`)

| env | meaning |
|-----|---------|
| `CMS_S3_BUCKET` | bucket serving the public site. **Unset ⇒ publishing disabled**, `POST /api/cms/publish` → `409`. |
| `CMS_S3_PREFIX` | key prefix (default `sat0ru`); must match `portfolio-client`'s `S3_PREFIX` / the CloudFront origin path. |
| `CMS_CONTENT_KEY` | live document basename (default `content.json`). |
| `CMS_CLOUDFRONT_DISTRIBUTION_ID` | distribution to invalidate (optional; unset skips invalidation). |
| `AWS_REGION`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` | standard AWS creds (prefer an EC2 instance role). |

IAM policy the server needs:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    { "Effect": "Allow", "Action": ["s3:PutObject"], "Resource": "arn:aws:s3:::YOUR_BUCKET/sat0ru/*" },
    { "Effect": "Allow", "Action": ["cloudfront:CreateInvalidation"], "Resource": "arn:aws:cloudfront::ACCOUNT_ID:distribution/YOUR_DIST_ID" }
  ]
}
```

On publish the server writes two objects: `sat0ru/content.json` (the live doc, `Cache-Control:
no-cache`) and `sat0ru/content/history/v<N>-<timestamp>.json` (immutable archive / rollback
trail).

## Data model

```
Project      { id, title, slug, description, stack: string[], github, liveLink, visible, order, updatedAt }
Experience   { id, logo, position, company, description, details: string[], techStack: string[],
               startDate, endDate, visible, order, updatedAt }
Blog         { id, title, slug, readTime, genre, date, body (markdown), visible, order, updatedAt }
Summary      { domain, imageSubText, heroHighlightText, heroName, heroSubText, heroDetails, updatedAt }
Publication  { id, version, publishedAt, status: "success"|"failed", error: string|null }
```

`slug` is unique across projects, and across blogs; auto-derived from the title when omitted;
must match `^[a-z0-9]+(-[a-z0-9]+)*$`. `github`/`liveLink` must be http(s) URLs when non-empty.
List endpoints return items ordered by `order`. New items sort last; reorder by `PATCH`ing
`order` (the CMS swaps two adjacent items).

## Endpoints

```
GET    /api/cms/content        assembled draft: { version, publishedAt, summary, projects, experiences, blogs }
GET    /api/cms/status         ChangeSummary: { hasUnpublishedChanges, changedSections[], neverPublished,
                                                lastPublishedAt, lastPublishVersion, lastPublishStatus, lastPublishError }
POST   /api/cms/publish        serialize draft → S3 + CloudFront; returns the Publication. 409 if not configured.
GET    /api/cms/publications   recent publication history (?limit=, default 20, max 50)

GET    /api/cms/projects
POST   /api/cms/projects                 body: CreateProjectInput
PATCH  /api/cms/projects/{id}             body: partial UpdateProjectInput (nil field = unchanged)
DELETE /api/cms/projects/{id}

GET    /api/cms/experiences
POST   /api/cms/experiences
PATCH  /api/cms/experiences/{id}
DELETE /api/cms/experiences/{id}

GET    /api/cms/blogs
POST   /api/cms/blogs
PATCH  /api/cms/blogs/{id}
DELETE /api/cms/blogs/{id}

GET    /api/cms/summary
PATCH  /api/cms/summary
```

Errors go through `internal/httpx` / `internal/apperr` (validation → 400, missing → 404,
publishing-not-configured → 409, publisher failure → 500 with the recorded `Publication` still
returned in the body).
