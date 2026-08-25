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
- **Content as data, not CMS**: page content (projects, blog metadata, experience, skills, socials) lives as typed arrays/objects in `src/data/*.ts`, imported directly by the section components. There is no backend fetch for this content today — adding/editing content means editing these files.
- **Blog content**: blog post bodies are markdown files under `src/assets/blogs/`, rendered via `react-markdown` + `remark-gfm` (see `BlogPage.tsx`).
- **Styling**: Tailwind CSS v4 via `@tailwindcss/vite`, using CSS custom properties for the design system (colors/spacing referenced as `bg-(--bg)`, `text-(--text-muted)`, `border-(--line)`, etc. rather than hardcoded Tailwind palette classes). Keep new UI consistent with this custom-property convention instead of introducing raw hex/Tailwind-default colors.
- **React Compiler**: enabled via `babel-plugin-react-compiler` in `vite.config.ts` — avoid manual `useMemo`/`useCallback` micro-optimizations that fight the compiler.

### Authenticated / "domain" area convention

The site is being split into a public marketing site and a gated "domain" area (see `/domain-expansion`, currently a password-entry stub with no real auth wired up yet). Going forward, **all authenticated/domain-restricted UI should live under a `domain` subdirectory inside each of the existing top-level dirs**, mirroring the public structure rather than being mixed into it:

- `src/pages/domain/...`
- `src/layouts/domain/...`
- `src/components/domain/...`

Keep this separation even for small pieces (a single domain-only component still goes in `components/domain/`), so gated code stays easy to audit separately from the public site.

## Backend architecture

- **Feature-sliced packages, not technical layers.** Each feature owns a self-contained package at `internal/<feature>/` holding its own `model.go`, `repository.go` (+ a `repository_<impl>.go` per backing store), `service.go`, and `handler.go` — e.g. `internal/todo/`. Do **not** create `internal/handler`, `internal/service`, `internal/repository`, or `internal/model` packages split by technical role — with several features, that shape lets unrelated domains reach into each other's internals and turns each layer package into an unbounded pile of unrelated files. `internal/todo/` is the reference example.
- **Repository interfaces are shaped for the eventual real store, not the current one.** A feature's `Repository` interface should look like something a SQL implementation would want (e.g. `Update(ctx, id, input UpdateInput)`, not a mutation closure) even while the only implementation is in-memory — don't let the in-memory implementation's convenience leak into the interface, since that's the seam that's expensive to change once a real backing store lands.
- **Shared cross-feature packages** (not tied to one feature): `internal/apperr` — domain error kinds (`InvalidInput`, `NotFound`, `Internal`), no HTTP knowledge, so services stay usable outside an HTTP handler. `internal/httpx` — `WriteJSON`, `WriteError` (maps an `*apperr.Error`'s `Kind` to an HTTP status once, for every feature), `DecodeJSON`. `internal/id` — resource ID generation. New shared concerns (auth, request logging, pagination helpers, etc.) belong here, not copy-pasted into each feature package.
- **Handlers should not hand-roll error → status mapping.** A service returns an `*apperr.Error` (via `apperr.InvalidInput(...)`, `apperr.NotFound(...)`, etc.); the handler just calls `httpx.WriteError(w, err)`. Don't add per-handler `switch { case errors.Is(...) }` blocks — that's the pattern that gets copy-pasted and drifts across features.
- **Composition happens in `cmd/main.go`** (`newRouter()`): construct each feature's repository → service → handler and call `.Register(mux)`. One block per feature; no DI framework — not warranted at this size.
- **Write tests for real logic** (non-trivial sort/comparison logic, validation rules) alongside the feature package, e.g. `internal/todo/repository_memory_test.go`, `internal/todo/service_test.go`. Trivial wiring/handler plumbing doesn't need it.

## Deployment

`client/scripts/deploy.sh` builds the app and syncs it to S3 + invalidates CloudFront. It reads config from `client/.env` (see `client/.env.example` for required/optional vars: `BUCKET_NAME`, `DISTRIBUTION_ID`, `S3_PREFIX`, plus optional `BUILD_DIR`, `AWS_PROFILE`, `BUILD_CMD`). Requires AWS CLI v2 configured locally.
