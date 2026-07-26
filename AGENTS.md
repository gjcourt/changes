# AGENTS.md

> changes is a Go web app: a searchable database of jazz-standard chord progressions with transposition to any key and Roman-numeral analysis. — https://github.com/gjcourt/changes

## Commands

| Command | Use |
|---------|-----|
| `make build` | Compile binary to `./changes` |
| `make run` | Build + run (serves on `:8080`) |
| `make test` | Run tests with race detector |
| `make lint` | golangci-lint |
| `make clean` | Remove build artifacts |
| `make all` | clean + lint + test + build |
| `make image` | Build + push multi-arch image to ghcr.io |

Single test: `go test ./internal/theory -run TestRoman -v`
Pre-push: `make all`

## Architecture

> Full write-up with a component diagram and the end-to-end request flow: [`docs/architecture.md`](docs/architecture.md).

Entry point: `cmd/changes/main.go`. Lean layering (the corpus is read-only static data, so no storage adapters):

- `internal/theory/` — the music-theory engine: pitch-class math, chord-symbol parsing, transposition, Roman-numeral analysis. **Zero dependencies outside stdlib**; knows nothing about HTTP, JSON, or the corpus.
- `internal/library/` — loads + validates the embedded standards corpus and renders a standard transposed to a target key (optionally with Roman numerals). Depends only on `theory`.
- `internal/server/` — HTTP driving adapter: serves the `web/` SPA and the JSON API.
- `web/` — frontend (vanilla HTML/CSS/JS, no build step). Served from disk via `WEB_DIR`.
- `internal/library/data/standards/*.json` — the corpus, **embedded** into the binary via `//go:embed`.

The bug-prone music logic lives in Go and is unit-tested; `web/` is a dumb renderer that calls the API.

## Conventions

- **`internal/theory/` imports only the standard library** — it's the pure core.
- **The corpus is validated at load** (`library.Load`): every key and chord symbol is parsed, so malformed data fails at startup, not at request time.
- **Chord transposition shifts the root (and slash bass); the suffix is carried through verbatim** — correct for every quality/extension and keeps display faithful.
- **Test files co-located** with implementation (`_test.go` in the same package).
- **Conventional Commits** (`feat:`, `fix:`, `chore:`, `refactor:`, `docs:`, `test:`, `ci:`).
- **Branch names** follow `<type>/<description>`; branch + PR for every change.

## Invariants

- `internal/theory/` must not import any third-party package.
- Roman-numeral analysis is **transposition-invariant**: the same progression yields the same numerals in any key (enforced by `TestRomanIsKeyRelative`).
- Adding a standard = adding a validated JSON file under `internal/library/data/standards/`; it is picked up by the embed glob with no code change.
- The compiled binary lives at `./changes`; never committed.

## What NOT to Do

- Do not put HTTP/JSON types in `internal/theory/`.
- Do not commit chord changes you haven't verified — see the data note below.
- Do not skip `make lint` / `make test` before committing.

## Domain

A reference tool for jazz musicians: pick a standard, see its changes as a lead-sheet grid, transpose to any of the 12 keys, and toggle Roman-numeral analysis. The corpus is a hand-curated seed of common lead-sheet changes (see `internal/library/data/standards/`), designed to scale toward the top ~100 standards. Changes are factual/functional harmony (the kind iReal Pro shares); each file records a `source` note. **Verify changes against a trusted source before adding** — wrong changes are worse than missing ones.

## API

| Route | Returns |
|---|---|
| `GET /api/standards` | `[{id,title,composer,key}]`, sorted by title |
| `GET /api/standards/{id}?key=<tonic>&roman=1` | the standard transposed to `key` (default original), Roman numerals when `roman` is truthy |
| `GET /api/health` | `200 {status, standards}` once the corpus is loaded (liveness/readiness probe) |

## Container image

`.github/workflows/image.yml` builds and pushes the image to GHCR on every push to `master` (plus manual `workflow_dispatch`). It authenticates with the built-in `GITHUB_TOKEN` — no PAT — and builds multi-arch (`linux/amd64,linux/arm64`). This is the CI-driven successor to the manual `make image` target.

**Image:** `ghcr.io/gjcourt/changes`

Each build publishes three tags:

| Tag | Mutability | Use |
|---|---|---|
| `YYYY-MM-DD` | mutable — a later same-day build overwrites it | build date (UTC) |
| `YYYY-MM-DD-<sha7>` | **immutable & unique** | **the tag to pin in deployments** |
| `latest` | mutable — always the newest build | convenience |

**Deploying:** after a push to `master`, read the exact published tag from the `image.yml` run (or `gh api user/packages/container/changes/versions`), then pin the `YYYY-MM-DD-<sha7>` tag in `homelab/apps/base/changes/deployment.yaml`.

**First-build gotcha:** if a `GITHUB_TOKEN` push ever 403s, the GHCR package exists but is unlinked (created by an old manual PAT push) — delete it (`gh api --method DELETE user/packages/container/changes`, needs the `delete:packages` scope) so the next run recreates it auto-linked, then re-run.
