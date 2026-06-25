// Package theory implements the music-theory engine for jazz lead-sheet
// changes: pitch-class math, chord-symbol parsing, transposition, and
// Roman-numeral analysis. It has zero dependencies outside the standard
// library and knows nothing about HTTP or storage.
package theory

import (
	"fmt"
	"strings"
)

// PitchClass is a chromatic pitch class in the range [0,12): C=0, C#=1, … B=11.
type PitchClass int

const octave = 12

// sharpNames spells each pitch class using sharps (used in sharp keys).
var sharpNames = [octave]string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}

// flatNames spells each pitch class using flats (used in flat keys).
var flatNames = [octave]string{"C", "Db", "D", "Eb", "E", "F", "Gb", "G", "Ab", "A", "Bb", "B"}

// letterPC maps a natural note letter to its pitch class.
var letterPC = map[byte]PitchClass{'C': 0, 'D': 2, 'E': 4, 'F': 5, 'G': 7, 'A': 9, 'B': 11}

// Normalize wraps a pitch class into [0,12).
func (p PitchClass) Normalize() PitchClass {
	return PitchClass(((int(p) % octave) + octave) % octave)
}

// Transpose returns the pitch class shifted by the given number of semitones.
func (p PitchClass) Transpose(semitones int) PitchClass {
	return PitchClass(int(p) + semitones).Normalize()
}

// Name spells the pitch class, preferring flats when prefersFlats is true.
func (p PitchClass) Name(prefersFlats bool) string {
	p = p.Normalize()
	if prefersFlats {
		return flatNames[p]
	}
	return sharpNames[p]
}

// ParseNote parses a note name (letter + optional run of #/b accidentals) into
// a pitch class, returning the number of bytes consumed.
func ParseNote(s string) (pc PitchClass, n int, err error) {
	if s == "" {
		return 0, 0, fmt.Errorf("theory: empty note")
	}
	letter := s[0]
	base, ok := letterPC[letter]
	if !ok {
		// Accept lowercase letters too (minor-key shorthand like "bb").
		base, ok = letterPC[upper(letter)]
		if !ok {
			return 0, 0, fmt.Errorf("theory: %q is not a note letter", string(letter))
		}
	}
	n = 1
	for n < len(s) {
		switch s[n] {
		case '#':
			base++
		case 'b':
			base--
		default:
			return base.Normalize(), n, nil
		}
		n++
	}
	return base.Normalize(), n, nil
}

func upper(b byte) byte {
	if b >= 'a' && b <= 'z' {
		return b - ('a' - 'A')
	}
	return b
}

// Key is a tonal center used for transposition spelling and Roman-numeral
// analysis. Minor only affects display conventions, not the pitch math here.
type Key struct {
	Tonic PitchClass
	Name  string
	Minor bool
	flats bool // computed at parse time: spell notes with flats?
}

// Natural-letter keys whose signature takes flats. Among major keys only F
// (one flat) qualifies; among minor keys D, G, C, F do (1–4 flats). C/A and
// the rest default to sharps.
var naturalFlatMajor = map[string]bool{"F": true}
var naturalFlatMinor = map[string]bool{"D": true, "G": true, "C": true, "F": true}

// ParseKey parses a key name such as "Eb", "F#", or "Bbm" (trailing m/min = minor).
func ParseKey(s string) (Key, error) {
	raw := strings.TrimSpace(s)
	if raw == "" {
		return Key{}, fmt.Errorf("theory: empty key")
	}
	minor := false
	body := raw
	switch {
	case strings.HasSuffix(body, "min"):
		minor, body = true, strings.TrimSuffix(body, "min")
	case strings.HasSuffix(body, "m"):
		minor, body = true, strings.TrimSuffix(body, "m")
	}
	pc, n, err := ParseNote(body)
	if err != nil {
		return Key{}, err
	}
	if n != len(body) {
		return Key{}, fmt.Errorf("theory: trailing %q in key %q", body[n:], raw)
	}
	// Decide flat vs sharp spelling. body[0] is the letter; any accidental
	// lives in body[1:]. An explicit accidental is decisive; otherwise fall
	// back to the natural-key signature convention. Minor flips a couple of
	// natural letters (e.g. F minor takes flats just like F major), which the
	// naturalFlatLetters table already covers for the common cases.
	accidentals := body[1:]
	flats := false
	switch {
	case strings.Contains(accidentals, "b"):
		flats = true
	case strings.Contains(accidentals, "#"):
		flats = false
	default:
		letter := strings.ToUpper(body[:1])
		if minor {
			flats = naturalFlatMinor[letter]
		} else {
			flats = naturalFlatMajor[letter]
		}
	}
	return Key{Tonic: pc, Name: raw, Minor: minor, flats: flats}, nil
}

// PrefersFlats reports whether notes in this key should be spelled with flats.
func (k Key) PrefersFlats() bool {
	return k.flats
}
