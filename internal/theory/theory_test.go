package theory

import "testing"

func mustKey(t *testing.T, s string) Key {
	t.Helper()
	k, err := ParseKey(s)
	if err != nil {
		t.Fatalf("ParseKey(%q): %v", s, err)
	}
	return k
}

func TestParseNote(t *testing.T) {
	cases := map[string]PitchClass{
		"C": 0, "C#": 1, "Db": 1, "D": 2, "Eb": 3, "E": 4, "F": 5,
		"F#": 6, "Gb": 6, "G": 7, "Ab": 8, "A": 9, "Bb": 10, "B": 11,
		"Cb": 11, "B#": 0,
	}
	for in, want := range cases {
		got, n, err := ParseNote(in)
		if err != nil {
			t.Errorf("ParseNote(%q): %v", in, err)
			continue
		}
		if got != want || n != len(in) {
			t.Errorf("ParseNote(%q) = %d (n=%d), want %d (n=%d)", in, got, n, want, len(in))
		}
	}
}

func TestParseChord(t *testing.T) {
	cases := []struct {
		sym        string
		root       PitchClass
		suffix     string
		bass       PitchClass
		hasBass    bool
		wantParses bool
	}{
		{sym: "C", root: 0, suffix: "", wantParses: true},
		{sym: "Cmaj7", root: 0, suffix: "maj7", wantParses: true},
		{sym: "Dm7", root: 2, suffix: "m7", wantParses: true},
		{sym: "F#m7b5", root: 6, suffix: "m7b5", wantParses: true},
		{sym: "Bb13", root: 10, suffix: "13", wantParses: true},
		{sym: "G7#11", root: 7, suffix: "7#11", wantParses: true},
		{sym: "Bb7/D", root: 10, suffix: "7", bass: 2, hasBass: true, wantParses: true},
		{sym: "", wantParses: false},
		{sym: "H7", wantParses: false},
	}
	for _, c := range cases {
		got, err := ParseChord(c.sym)
		if !c.wantParses {
			if err == nil {
				t.Errorf("ParseChord(%q): expected error, got %+v", c.sym, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseChord(%q): %v", c.sym, err)
			continue
		}
		if got.Root != c.root || got.Suffix != c.suffix {
			t.Errorf("ParseChord(%q) = root %d suffix %q, want root %d suffix %q",
				c.sym, got.Root, got.Suffix, c.root, c.suffix)
		}
		if c.hasBass && (got.Bass == nil || *got.Bass != c.bass) {
			t.Errorf("ParseChord(%q) bass = %v, want %d", c.sym, got.Bass, c.bass)
		}
		if !c.hasBass && got.Bass != nil {
			t.Errorf("ParseChord(%q) unexpected bass %d", c.sym, *got.Bass)
		}
	}
}

func TestQuality(t *testing.T) {
	cases := map[string]Quality{
		"C": QualityMajor, "Cmaj7": QualityMajor, "C6": QualityMajor, "Cadd9": QualityMajor,
		"Csus4": QualityMajor, "CM7": QualityMajor,
		"Cm": QualityMinor, "Cm7": QualityMinor, "Cmin7": QualityMinor, "C-7": QualityMinor,
		"C7": QualityDominant, "C9": QualityDominant, "C13": QualityDominant, "C7#9": QualityDominant,
		"Cm7b5": QualityHalfDim, "Cø7": QualityHalfDim,
		"Cdim": QualityDiminished, "Cdim7": QualityDiminished, "Co7": QualityDiminished,
		"Caug": QualityAugmented, "C+": QualityAugmented,
	}
	for sym, want := range cases {
		c, err := ParseChord(sym)
		if err != nil {
			t.Fatalf("ParseChord(%q): %v", sym, err)
		}
		if got := c.Quality(); got != want {
			t.Errorf("Quality(%q) = %d, want %d", sym, got, want)
		}
	}
}

func TestTransposeRootSpelling(t *testing.T) {
	cases := []struct {
		chord     string
		semitones int
		key       string
		want      string
	}{
		{"C", 3, "Eb", "Eb"},       // flat key spells with flats
		{"C", 3, "D", "D#"},        // sharp key spells with sharps
		{"Cmaj7", 2, "D", "Dmaj7"}, // suffix carried through (C+2 = D)
		{"Dm7", 5, "G", "Gm7"},     // D up a 4th = G, in G stays sharp side
		{"G7", 5, "C", "C7"},       // V7 of C transposed
		{"Bb7/D", 2, "C", "C7/E"},  // bass transposes too
	}
	for _, c := range cases {
		ch, err := ParseChord(c.chord)
		if err != nil {
			t.Fatalf("ParseChord(%q): %v", c.chord, err)
		}
		got := ch.Transpose(c.semitones, mustKey(t, c.key)).StringIn(mustKey(t, c.key))
		if got != c.want {
			t.Errorf("%q +%d in %s = %q, want %q", c.chord, c.semitones, c.key, got, c.want)
		}
	}
}

func TestTransposeRoundTripPreservesIntervals(t *testing.T) {
	// Transposing a progression up 12 semitones must return the same chords.
	prog := []string{"Cmaj7", "Am7", "Dm7", "G7", "F#m7b5", "Bb7/D"}
	key := mustKey(t, "C")
	for _, sym := range prog {
		ch, err := ParseChord(sym)
		if err != nil {
			t.Fatalf("ParseChord(%q): %v", sym, err)
		}
		up := ch.Transpose(12, key)
		if up.Root != ch.Root {
			t.Errorf("%q +12 changed root: %d -> %d", sym, ch.Root, up.Root)
		}
		if (up.Bass == nil) != (ch.Bass == nil) {
			t.Errorf("%q +12 changed bass presence", sym)
		}
	}
}

func TestRomanDiatonicMajor(t *testing.T) {
	c := mustKey(t, "C")
	cases := map[string]string{
		"Cmaj7": "Imaj7",
		"Dm7":   "ii7",
		"Em7":   "iii7",
		"Fmaj7": "IVmaj7",
		"G7":    "V7",
		"Am7":   "vi7",
		"Bm7b5": "viiø7",
		"Dm7b5": "iiø7",
		"Bdim7": "vii°7",
		"Ab7":   "bVI7",
		"Db7":   "bII7",
	}
	for sym, want := range cases {
		ch, err := ParseChord(sym)
		if err != nil {
			t.Fatalf("ParseChord(%q): %v", sym, err)
		}
		if got := ch.Roman(c); got != want {
			t.Errorf("Roman(%q in C) = %q, want %q", sym, got, want)
		}
	}
}

func TestRomanIsKeyRelative(t *testing.T) {
	// The same ii-V-I shape must yield the same Roman numerals in any key.
	want := []string{"ii7", "V7", "Imaj7"}
	for _, key := range []string{"C", "Eb", "F#", "Bb", "A"} {
		k := mustKey(t, key)
		tonic := k.Tonic
		prog := []Chord{
			{Root: tonic.Transpose(2), Suffix: "m7"}, // ii
			{Root: tonic.Transpose(7), Suffix: "7"},  // V
			{Root: tonic, Suffix: "maj7"},            // I
		}
		for i, ch := range prog {
			if got := ch.Roman(k); got != want[i] {
				t.Errorf("key %s: Roman(degree %d) = %q, want %q", key, i, got, want[i])
			}
		}
	}
}

func TestKeyPrefersFlats(t *testing.T) {
	flat := []string{"F", "Bb", "Eb", "Ab", "Db", "Gb", "Bbm", "Cm"}
	sharp := []string{"C", "G", "D", "A", "E", "B", "F#", "Em"}
	for _, s := range flat {
		if !mustKey(t, s).PrefersFlats() {
			t.Errorf("key %s should prefer flats", s)
		}
	}
	for _, s := range sharp {
		if mustKey(t, s).PrefersFlats() {
			t.Errorf("key %s should prefer sharps", s)
		}
	}
}
