# changes

A searchable database of jazz-standard chord progressions, with **transposition
to any key** and **Roman-numeral analysis**. Pick a tune, read its changes as a
lead-sheet grid, transpose, and toggle functional analysis.

Go server + embedded JSON corpus + a dependency-free vanilla-JS frontend.

## Run

```bash
make run          # serves on http://localhost:8080
# or
go run ./cmd/changes
```

Environment: `ADDR` (default `:8080`), `WEB_DIR` (default `web`).

## How it works

- `internal/theory/` — the engine: chord-symbol parsing, transposition, and
  Roman-numeral analysis. Pure, stdlib-only, heavily unit-tested.
- `internal/library/` — loads + validates the embedded corpus and renders a
  standard transposed to a target key.
- `internal/server/` — JSON API + static `web/` SPA.
- `internal/library/data/standards/*.json` — the corpus (embedded in the binary).

API:

```
GET /api/standards                              # list
GET /api/standards/{id}?key=Eb&roman=1          # transposed + analyzed
```

## The corpus

A hand-curated seed of common lead-sheet changes (currently: Blue Bossa, Bb Jazz
Blues, So What, Autumn Leaves, Take the A Train), designed to scale toward the
top ~100 standards. Changes are functional harmony (the kind iReal Pro shares);
each file notes its `source`.

**Adding a standard** — drop a validated JSON file in
`internal/library/data/standards/` (the embed glob picks it up, no code change):

```json
{
  "id": "tune-id",
  "title": "Tune Title",
  "composer": "Composer",
  "key": "C",
  "form": "AABA (32-bar)",
  "meter": "4/4",
  "source": "where these changes came from",
  "sections": [
    { "label": "A", "bars": [["Cmaj7"], ["A7"], ["Dm7"], ["G7"]] }
  ]
}
```

Each bar is a list of chord symbols (so split bars like `["Bb7","G7"]` work).
`make test` validates every chord parses. **Verify changes against a trusted
source before committing** — wrong changes are worse than missing ones.

## Roadmap

- Backfill the corpus toward 100 standards.
- Optional: full functional analysis (secondary-dominant labeling, e.g. `V7/ii`).
- Optional: deploy to the homelab cluster (`changes.burntbytes.com`).
