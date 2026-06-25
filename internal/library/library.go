// Package library loads the corpus of jazz standards and renders their
// changes transposed to any key, optionally with Roman-numeral analysis.
// It depends only on the theory engine and the standard library.
package library

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"changes/internal/theory"
)

//go:embed data/standards/*.json
var embedded embed.FS

// Standard is one tune as stored on disk: metadata plus an ordered list of
// sections, each a list of bars, each bar a list of chord symbols.
type Standard struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	Composer string    `json:"composer"`
	Key      string    `json:"key"`
	Form     string    `json:"form"`
	Meter    string    `json:"meter"`
	Sections []Section `json:"sections"`
	Source   string    `json:"source,omitempty"`
}

// Section is a labelled run of bars (e.g. "A", "Bridge").
type Section struct {
	Label string     `json:"label"`
	Bars  [][]string `json:"bars"`
}

// Summary is the lightweight metadata returned by List.
type Summary struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Composer string `json:"composer"`
	Key      string `json:"key"`
}

// Library is an immutable, validated in-memory corpus.
type Library struct {
	byID  map[string]Standard
	order []string // ids sorted by title
}

// Default loads the corpus embedded in the binary.
func Default() (*Library, error) {
	return Load(embedded, "data/standards")
}

// Load reads and validates every *.json standard under dir in fsys.
func Load(fsys fs.FS, dir string) (*Library, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, fmt.Errorf("library: read %s: %w", dir, err)
	}
	lib := &Library{byID: make(map[string]Standard)}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		raw, rerr := fs.ReadFile(fsys, dir+"/"+e.Name())
		if rerr != nil {
			return nil, fmt.Errorf("library: read %s: %w", e.Name(), rerr)
		}
		var std Standard
		if jerr := json.Unmarshal(raw, &std); jerr != nil {
			return nil, fmt.Errorf("library: parse %s: %w", e.Name(), jerr)
		}
		if verr := validate(std); verr != nil {
			return nil, fmt.Errorf("library: invalid %s: %w", e.Name(), verr)
		}
		if _, dup := lib.byID[std.ID]; dup {
			return nil, fmt.Errorf("library: duplicate id %q (%s)", std.ID, e.Name())
		}
		lib.byID[std.ID] = std
		lib.order = append(lib.order, std.ID)
	}
	sort.Slice(lib.order, func(i, j int) bool {
		return lib.byID[lib.order[i]].Title < lib.byID[lib.order[j]].Title
	})
	return lib, nil
}

// validate ensures the key parses and every chord symbol is well-formed, so a
// malformed corpus fails at load rather than at request time.
func validate(s Standard) error {
	if s.ID == "" || s.Title == "" {
		return fmt.Errorf("missing id or title")
	}
	if _, err := theory.ParseKey(s.Key); err != nil {
		return fmt.Errorf("key %q: %w", s.Key, err)
	}
	for si, sec := range s.Sections {
		for bi, bar := range sec.Bars {
			for _, sym := range bar {
				if _, err := theory.ParseChord(sym); err != nil {
					return fmt.Errorf("section %d bar %d: %w", si, bi, err)
				}
			}
		}
	}
	return nil
}

// List returns all standards' metadata, sorted by title.
func (l *Library) List() []Summary {
	out := make([]Summary, 0, len(l.order))
	for _, id := range l.order {
		s := l.byID[id]
		out = append(out, Summary{ID: s.ID, Title: s.Title, Composer: s.Composer, Key: s.Key})
	}
	return out
}

// Get returns the raw standard by id.
func (l *Library) Get(id string) (Standard, bool) {
	s, ok := l.byID[id]
	return s, ok
}

// RenderedChord is one chord after transposition, with optional analysis.
type RenderedChord struct {
	Symbol string `json:"symbol"`
	Roman  string `json:"roman,omitempty"`
}

// RenderedSection mirrors Section with rendered chords.
type RenderedSection struct {
	Label string            `json:"label"`
	Bars  [][]RenderedChord `json:"bars"`
}

// Rendered is a standard transposed to a target key, ready for the frontend.
type Rendered struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Composer    string            `json:"composer"`
	OriginalKey string            `json:"originalKey"`
	Key         string            `json:"key"`
	Form        string            `json:"form"`
	Meter       string            `json:"meter"`
	Sections    []RenderedSection `json:"sections"`
}

// Render transposes a standard to targetKey (empty = its original key) and,
// when withRoman is set, annotates each chord with its Roman numeral.
func (l *Library) Render(id, targetKey string, withRoman bool) (Rendered, error) {
	std, ok := l.byID[id]
	if !ok {
		return Rendered{}, fmt.Errorf("library: no standard %q", id)
	}
	orig, err := theory.ParseKey(std.Key)
	if err != nil {
		return Rendered{}, err
	}
	target := orig
	if targetKey != "" {
		if target, err = theory.ParseKey(targetKey); err != nil {
			return Rendered{}, fmt.Errorf("library: target key %q: %w", targetKey, err)
		}
	}
	semitones := int(target.Tonic) - int(orig.Tonic)

	out := Rendered{
		ID: std.ID, Title: std.Title, Composer: std.Composer,
		OriginalKey: std.Key, Key: target.Name, Form: std.Form, Meter: std.Meter,
	}
	for _, sec := range std.Sections {
		rsec := RenderedSection{Label: sec.Label}
		for _, bar := range sec.Bars {
			rbar := make([]RenderedChord, 0, len(bar))
			for _, sym := range bar {
				ch, perr := theory.ParseChord(sym)
				if perr != nil {
					return Rendered{}, perr // validated at load, but be safe
				}
				t := ch.Transpose(semitones, target)
				rc := RenderedChord{Symbol: t.StringIn(target)}
				if withRoman {
					rc.Roman = t.Roman(target)
				}
				rbar = append(rbar, rc)
			}
			rsec.Bars = append(rsec.Bars, rbar)
		}
		out.Sections = append(out.Sections, rsec)
	}
	return out, nil
}
