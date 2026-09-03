# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository layout

This is a personal portfolio site (`sat0ru.dev`) split into two independent projects that are not currently wired together:

- `client/` — React 19 + TypeScript + Vite frontend. This is the actively developed part of the repo.
- `server/` — Go backend, early scaffold. Package path `github.com/Sam-Frost/portfolio`. See "Backend architecture" below for the package convention.

## Commands

All frontend commands run from `client/`:

- `bun install` — install deps (repo uses `bun.lock`, prefer bun over npm/yarn)
- `bun run dev` — start Vite dev server
- `bun run build` — typecheck (`tsc -b`) then production build via Vite
- `bun run lint` — lint with oxlint
- `bun run preview` — preview the production build locally

There is no test runner configured in `client/` yet.

Backend (`server/`), from that directory:
- `go run ./cmd` — run the server
- `go build ./...` — build
- `go test ./...` — run tests

## Frontend architecture

- **Routing**: `client/src/router.tsx` defines all routes via `createBrowserRouter`. Most routes nest under `RootLayout` (`client/src/layouts/RootLayout.tsx`), which renders `Navbar` + `Footer` around an `<Outlet />`. `/domain-expansion` is deliberately mounted outside `RootLayout` (no navbar/footer) as a standalone password-gated entry page.
- **Pages vs. sections vs. components**: `src/pages/` holds route-level components; `src/components/` holds reusable pieces, several of which are themselves page sections (e.g. `HeroSection`, `ProjectSection`, `BlogSection`, `ExperienceSection`, `SkillSection`) paired with a singular per-item component (`Project`, `Blog`, `Experience`, `Skill`).
- **Content via CMS + runtime `content.json`**: projects, experience, writings (blogs), and the home summary are edited in the domain-area CMS (`domain-client/src/features/cms/`, backed by `server/internal/cms/`), saved server-side as a draft, and pushed live by a **Publish** that writes a single `content.json` to the public site's S3 origin. `portfolio-client` fetches `/content.json` at runtime via `src/content/ContentContext.tsx` (`useContent()`), with `src/content/fallback.ts` compiled in as the offline default. Skills and socials are still code-edited in `portfolio-client/src/data/*.ts`. See `server/docs/cms-api.md`.
- **Blog content**: blog post bodies are markdown, edited in the CMS and stored in Postgres (shipped inline in `content.json`), rendered via `react-markdown` + `remark-gfm` (see `BlogPage.tsx`). The seed post still exists as a file under `portfolio-client/src/assets/blogs/` only because `fallback.ts` imports it.
- **Styling**: Tailwind CSS v4 via `@tailwindcss/vite`, using CSS custom properties for the design system (colors/spacing referenced as `bg-(--bg)`, `text-(--text-muted)`, `border-(--line)`, etc. rather than hardcoded Tailwind palette classes). Keep new UI consistent with this custom-property convention instead of introducing raw hex/Tailwind-default colors.
- **React Compiler**: enabled via `babel-plugin-react-compiler` in `vite.config.ts` — avoid manual `useMemo`/`useCallback` micro-optimizations that fight the compiler.
- **`domain-client` is an installed PWA**: `public/manifest.webmanifest` + `public/sw.js` (registered in `src/lib/pwa.ts` from `main.tsx`). The service worker's only job is Web Push notifications (show every push, route a click back through `router.navigate`) — it caches nothing. Web Push on iOS only works from the Home-Screen app, which is also why `router.tsx` switches to `createMemoryRouter` in standalone mode.

### Authenticated / "domain" area convention

The site is being split into a public marketing site and a gated "domain" area (see `/domain-expansion`, currently a password-entry stub with no real auth wired up yet). Going forward, **all authenticated/domain-restricted UI should live under a `domain` subdirectory inside each of the existing top-level dirs**, mirroring the public structure rather than being mixed into it:

- `src/pages/domain/...`
- `src/layouts/domain/...`
- `src/components/domain/...`

Keep this separation even for small pieces (a single domain-only component still goes in `components/domain/`), so gated code stays easy to audit separately from the public site.

## Backend architecture

- **Feature-sliced packages, not technical layers.** Each feature owns a self-contained package at `internal/<feature>/` holding its own `model.go`, `repository.go` (+ a `repository_<impl>.go` per backing store), `service.go`, and `handler.go` — e.g. `internal/todo/`. Do **not** create `internal/handler`, `internal/service`, `internal/repository`, or `internal/model` packages split by technical role — with several features, that shape lets unrelated domains reach into each other's internals and turns each layer package into an unbounded pile of unrelated files. `internal/todo/` is the reference example.
- **Repository interfaces are shaped for the eventual real store, not the current one.** A feature's `Repository` interface should look like something a SQL implementation would want (e.g. `Update(ctx, id, input UpdateInput)`, not a mutation closure) even while the only implementation is in-memory — don't let the in-memory implementation's convenience leak into the interface, since that's the seam that's expensive to change once a real backing store lands.
- **Shared cross-feature packages** (not tied to one feature): `internal/apperr` — domain error kinds (`InvalidInput`, `NotFound`, `Internal`), no HTTP knowledge, so services stay usable outside an HTTP handler. `internal/httpx` — `WriteJSON`, `WriteError` (maps an `*apperr.Error`'s `Kind` to an HTTP status once, for every feature), `DecodeJSON`. `internal/id` — resource ID generation. `internal/mailer` — outbound SMTP (`Mailer` interface + `SMTPMailer`/`NoopMailer`, built from `SMTP_*` env). New shared concerns (auth, request logging, pagination helpers, etc.) belong here, not copy-pasted into each feature package.
- **Background work: `internal/scheduler`.** Everything else is evaluated lazily on an HTTP request — the one exception is `internal/scheduler`, a single in-process ticker started from `main()` that fires the morning todo digest and per-todo reminders. It depends on feature services through narrow local interfaces (no new package cycles). Anything that must happen on a clock, not a request, goes here — don't spawn a second goroutine/process. Notification delivery (email + Web Push subscriptions + fan-out) lives in `internal/notification`; per-todo alarms in `internal/reminder` (a repeating reminder self-cancels via `JOIN todos WHERE done = false`). Work Profile — tabbed, Jira-gated work tasks — is `internal/workprofile`. See `server/docs/notifications-api.md` and `work-profile-api.md`.
- **Handlers should not hand-roll error → status mapping.** A service returns an `*apperr.Error` (via `apperr.InvalidInput(...)`, `apperr.NotFound(...)`, etc.); the handler just calls `httpx.WriteError(w, err)`. Don't add per-handler `switch { case errors.Is(...) }` blocks — that's the pattern that gets copy-pasted and drifts across features.
- **Composition happens in `cmd/main.go`** (`newRouter()`): construct each feature's repository → service → handler and call `.Register(mux)`. One block per feature; no DI framework — not warranted at this size.
- **Write tests for real logic** (non-trivial sort/comparison logic, validation rules) alongside the feature package, e.g. `internal/todo/repository_memory_test.go`, `internal/todo/service_test.go`. Trivial wiring/handler plumbing doesn't need it.

## Deployment

`portfolio-client/scripts/deploy.sh` builds the app and syncs it to S3 + invalidates CloudFront. It reads config from `portfolio-client/.env` (see `.env.example` for required/optional vars: `BUCKET_NAME`, `DISTRIBUTION_ID`, `S3_PREFIX`, plus optional `BUILD_DIR`, `AWS_PROFILE`, `BUILD_CMD`). Requires AWS CLI v2 configured locally. The sync excludes `content.json` / `content/*` from `--delete` — those are owned by the CMS Publish flow, not the frontend build.

**CMS publishing**: `POST /api/cms/publish` (from the domain-area CMS) has the Go server write `content.json` to the same bucket/prefix and invalidate CloudFront, using the AWS SDK (no `aws` CLI on the box). Configured via `CMS_S3_BUCKET` / `CMS_S3_PREFIX` / `CMS_CLOUDFRONT_DISTRIBUTION_ID` / `AWS_REGION` in the server's env; unset ⇒ publishing disabled (dev). IAM: `s3:PutObject` on the prefix + `cloudfront:CreateInvalidation`. Details in `server/docs/cms-api.md`.
