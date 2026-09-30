<!-- readme-type: service -->
# changes

Jazz standard chord progression database — transpose to any key + Roman-numeral analysis

Jazz musicians need a quick, accurate reference for a standard's changes in a key
that isn't the original, and transposing a lead sheet by hand is slow and
error-prone. changes is a searchable database of curated lead-sheet changes for
common jazz standards: pick a tune, read it as a lead-sheet grid, transpose it to
any of the 12 keys, and toggle Roman-numeral analysis. It's for musicians,
students, and educators who want changes they can trust rather than a chord
chart copied from an inconsistent source.

**Status:** running on the homelab since 2026-06 — production at
`changes.burntbytes.com`, staging at `changes.stage.burntbytes.com`.

## Quick start

Needs: Go 1.25+.

```bash
git clone https://github.com/gjcourt/changes && cd changes
make run
```

Then open <http://localhost:8080>.

## Usage

Transpose a standard to a new key and get its Roman-numeral analysis:

```bash
curl 'http://localhost:8080/api/standards/blue-bossa?key=Eb&roman=1'
```

## Configuration

| Variable | Default | Meaning |
|---|---|---|
| `ADDR` | `:8080` | Listen address |
| `WEB_DIR` | `web` | Directory the static frontend is served from |

## How it works

A pure, stdlib-only theory engine (`internal/theory`) parses chord symbols,
transposes them, and derives Roman numerals; `internal/library` loads and
validates the embedded JSON corpus and renders a standard through the engine;
`internal/server` exposes both as a JSON API and serves the vanilla-JS `web/`
frontend. There is no database — the corpus is embedded in the binary and
validated once at startup. See [`docs/architecture.md`](docs/architecture.md)
for the component diagram and request flow.

## Development

```bash
make lint   # golangci-lint run ./... — CI lint job
make test   # go test -race ./... — CI test job
```

Conventions for contributors and agents: [AGENTS.md](AGENTS.md).

## Deployment

Runs on the homelab; the manifests live in
[`gjcourt/homelab`](https://github.com/gjcourt/homelab/tree/master/apps/base/changes)
(`apps/base/changes/`, with the production and staging routes in
`apps/production/changes/` and `apps/staging/changes/`). CI ([`.github/workflows/image.yml`](.github/workflows/image.yml))
builds and pushes `ghcr.io/gjcourt/changes` on every push to `master`; bump the
image tag in `apps/base/changes/deployment.yaml` in the homelab repo to roll
out a new build.

## License

No licence file yet.
