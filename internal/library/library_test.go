package library

import "testing"

func load(t *testing.T) *Library {
	t.Helper()
	l, err := Default()
	if err != nil {
		t.Fatalf("Default(): %v", err)
	}
	return l
}

func TestDefaultLoadsAndValidates(t *testing.T) {
	l := load(t)
	if len(l.List()) < 5 {
		t.Fatalf("expected >=5 standards, got %d", len(l.List()))
	}
	if _, ok := l.Get("blue-bossa"); !ok {
		t.Errorf("blue-bossa not loaded")
	}
}

func TestListSortedByTitle(t *testing.T) {
	got := load(t).List()
	for i := 1; i < len(got); i++ {
		if got[i-1].Title > got[i].Title {
			t.Errorf("List not sorted: %q before %q", got[i-1].Title, got[i].Title)
		}
	}
}

func TestRenderOriginalKeyRoman(t *testing.T) {
	l := load(t)
	r, err := l.Render("blue-bossa", "", true)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	first := r.Sections[0].Bars[0][0]
	if first.Symbol != "Cm7" {
		t.Errorf("first chord = %q, want Cm7", first.Symbol)
	}
	if first.Roman != "i7" {
		t.Errorf("first roman = %q, want i7", first.Roman)
	}
	// The ii-V into the tonic: bar 5 Dm7b5 -> iiø7, bar 6 G7 -> V7.
	if got := r.Sections[0].Bars[4][0].Roman; got != "iiø7" {
		t.Errorf("bar5 roman = %q, want iiø7", got)
	}
	if got := r.Sections[0].Bars[5][0].Roman; got != "V7" {
		t.Errorf("bar6 roman = %q, want V7", got)
	}
}

func TestRenderTransposed(t *testing.T) {
	l := load(t)
	// Blue Bossa (Cm) up to Dm: every root shifts up 2 semitones.
	r, err := l.Render("blue-bossa", "Dm", false)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if r.Key != "Dm" || r.OriginalKey != "Cm" {
		t.Errorf("keys = %s/%s, want Dm/Cm", r.Key, r.OriginalKey)
	}
	if got := r.Sections[0].Bars[0][0].Symbol; got != "Dm7" {
		t.Errorf("transposed first chord = %q, want Dm7", got)
	}
	// Roman numerals are invariant under transposition: re-render with roman on.
	rr, _ := l.Render("blue-bossa", "Dm", true)
	if got := rr.Sections[0].Bars[0][0].Roman; got != "i7" {
		t.Errorf("transposed roman = %q, want i7 (invariant)", got)
	}
}

func TestRenderUnknownStandard(t *testing.T) {
	if _, err := load(t).Render("nope", "", false); err == nil {
		t.Errorf("expected error for unknown id")
	}
}
