# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository layout

This is a personal portfolio site (`sat0ru.dev`) split into two independent projects that are not currently wired together:

- `client/` — React 19 + TypeScript + Vite frontend. This is the actively developed part of the repo.
- `server/` — Go backend, very early scaffold (`go.mod` + a `Hello`-printing `main.go`). Package path `github.com/Sam-Frost/portfolio`. Structured for a standard layered API (`internal/handler`, `internal/service`, `internal/repository`, `internal/model`) but those directories are currently empty.

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
- `go test ./...` — run tests (once any exist)

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

## Deployment

`client/scripts/deploy.sh` builds the app and syncs it to S3 + invalidates CloudFront. It reads config from `client/.env` (see `client/.env.example` for required/optional vars: `BUCKET_NAME`, `DISTRIBUTION_ID`, `S3_PREFIX`, plus optional `BUILD_DIR`, `AWS_PROFILE`, `BUILD_CMD`). Requires AWS CLI v2 configured locally.
