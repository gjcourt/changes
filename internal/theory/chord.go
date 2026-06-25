package theory

import (
	"fmt"
	"strings"
)

// Quality is the harmonic family of a chord, used to decide Roman-numeral
// casing and a normalized suffix. The full original suffix is preserved
// separately for faithful display.
type Quality int

const (
	QualityMajor      Quality = iota // C, Cmaj7, C6, Cadd9, Csus4
	QualityMinor                     // Cm, Cm7, Cm6, Cm9
	QualityDominant                  // C7, C9, C13, C7alt
	QualityHalfDim                   // Cm7b5 / Cø
	QualityDiminished                // Cdim, Co, Cdim7, Co7
	QualityAugmented                 // Caug, C+
)

// Chord is a parsed chord symbol: a root, the verbatim quality suffix, and an
// optional slash-bass note. Transposition shifts the root (and bass); the
// suffix is carried through unchanged, which is correct for every quality and
// extension.
type Chord struct {
	Root   PitchClass
	Suffix string // e.g. "maj7", "m7", "7#11", "" for a bare major triad
	Bass   *PitchClass

	rootText string // original root spelling, for round-tripping unparsed
}

// ParseChord parses a chord symbol such as "Cmaj7", "F#m7b5", or "Bb7/D".
func ParseChord(sym string) (Chord, error) {
	s := strings.TrimSpace(sym)
	if s == "" {
		return Chord{}, fmt.Errorf("theory: empty chord")
	}

	root, n, err := ParseNote(s)
	if err != nil {
		return Chord{}, fmt.Errorf("theory: chord %q: %w", sym, err)
	}
	rootText := s[:n]
	rest := s[n:]

	var bass *PitchClass
	if i := strings.IndexByte(rest, '/'); i >= 0 {
		bassText := rest[i+1:]
		bpc, bn, berr := ParseNote(bassText)
		if berr != nil {
			return Chord{}, fmt.Errorf("theory: chord %q bass: %w", sym, berr)
		}
		if bn != len(bassText) {
			return Chord{}, fmt.Errorf("theory: chord %q has trailing %q after bass", sym, bassText[bn:])
		}
		bass = &bpc
		rest = rest[:i]
	}

	return Chord{Root: root, Suffix: rest, Bass: bass, rootText: rootText}, nil
}

// Quality classifies the chord's suffix into a harmonic family.
func (c Chord) Quality() Quality {
	s := c.Suffix
	switch {
	case s == "":
		return QualityMajor
	case hasAnyPrefix(s, "m7b5", "min7b5", "ø", "halfdim"):
		return QualityHalfDim
	case hasAnyPrefix(s, "dim", "o", "°"):
		return QualityDiminished
	case hasAnyPrefix(s, "aug", "+"):
		return QualityAugmented
	case hasAnyPrefix(s, "maj", "Maj", "M", "Δ", "6/9", "6", "add", "sus", "5"):
		// Major family: maj7, 6, add9, sus4, power chord, etc.
		return QualityMajor
	case hasAnyPrefix(s, "m", "min", "-"):
		// Minor family — checked AFTER maj/M so "M7" isn't read as minor.
		return QualityMinor
	default:
		// Bare numbers/alterations (7, 9, 11, 13, 7#9, …) are dominant.
		return QualityDominant
	}
}

// Transpose returns the chord moved by semitones, spelled for the target key.
func (c Chord) Transpose(semitones int, target Key) Chord {
	out := Chord{
		Root:   c.Root.Transpose(semitones),
		Suffix: c.Suffix,
	}
	if c.Bass != nil {
		b := c.Bass.Transpose(semitones)
		out.Bass = &b
	}
	out.rootText = out.Root.Name(target.PrefersFlats())
	return out
}

// String renders the chord symbol (root + suffix + optional /bass), spelling
// any note for which we have no explicit text using sharps by default.
func (c Chord) String() string {
	root := c.rootText
	if root == "" {
		root = c.Root.Name(false)
	}
	out := root + c.Suffix
	if c.Bass != nil {
		out += "/" + c.Bass.Name(false)
	}
	return out
}

// StringIn renders the chord spelled for the given key (root and bass).
func (c Chord) StringIn(k Key) string {
	out := c.Root.Name(k.PrefersFlats()) + c.Suffix
	if c.Bass != nil {
		out += "/" + c.Bass.Name(k.PrefersFlats())
	}
	return out
}

// chromaticRoman labels each semitone above the tonic with its conventional
// (uppercase) Roman degree. Casing/suffix are applied per chord quality.
var chromaticRoman = [octave]string{
	"I", "bII", "II", "bIII", "III", "IV", "#IV", "V", "bVI", "VI", "bVII", "VII",
}

// Roman returns the Roman-numeral analysis of the chord in the given key,
// e.g. "ii7", "V7", "Imaj7", "vim7", "iiø7". Analysis is degree-and-quality
// (literal), not full functional analysis (no secondary-dominant labeling yet).
func (c Chord) Roman(k Key) string {
	degree := int(c.Root.Transpose(-int(k.Tonic)).Normalize())
	numeral := chromaticRoman[degree]
	q := c.Quality()

	if q == QualityMinor || q == QualityHalfDim || q == QualityDiminished {
		numeral = strings.ToLower(numeral)
	}

	switch q {
	case QualityMajor:
		return numeral + majorRomanSuffix(c.Suffix)
	case QualityMinor:
		return numeral + minorRomanSuffix(c.Suffix)
	case QualityDominant:
		return numeral + c.Suffix // "V7", "I7", "II7#11"
	case QualityHalfDim:
		return numeral + "ø7"
	case QualityDiminished:
		return numeral + "°7"
	case QualityAugmented:
		return numeral + "+"
	default:
		return numeral + c.Suffix
	}
}

// majorRomanSuffix keeps the recognizable part of a major-family suffix.
func majorRomanSuffix(s string) string {
	switch {
	case s == "":
		return ""
	case hasAnyPrefix(s, "maj", "Maj", "Δ"):
		return "maj7"
	case hasAnyPrefix(s, "M7"):
		return "maj7"
	default:
		return s // "6", "6/9", "add9", "sus4", …
	}
}

// minorRomanSuffix drops the redundant leading minor marker (the lowercase
// numeral already conveys "minor"), keeping the meaningful tail.
func minorRomanSuffix(s string) string {
	for _, p := range []string{"min", "m", "-"} {
		if strings.HasPrefix(s, p) {
			return strings.TrimPrefix(s, p) // "m7" -> "7", "m" -> ""
		}
	}
	return s
}

func hasAnyPrefix(s string, prefixes ...string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}
