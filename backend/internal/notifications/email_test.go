package notifications

import "testing"

func TestStripCRLFRemovesCarriageReturnsAndNewlines(t *testing.T) {
	in := "Subject\r\nBcc: attacker@evil.com"
	want := "SubjectBcc: attacker@evil.com"
	if got := stripCRLF(in); got != want {
		t.Errorf("stripCRLF(%q) = %q, want %q", in, got, want)
	}
}

func TestStripCRLFNoChangeForCleanInput(t *testing.T) {
	in := "Monitor recovered: API"
	if got := stripCRLF(in); got != in {
		t.Errorf("stripCRLF(%q) = %q, want unchanged", in, got)
	}
}

func TestStripCRLFHandlesOnlyLF(t *testing.T) {
	in := "line1\nline2"
	want := "line1line2"
	if got := stripCRLF(in); got != want {
		t.Errorf("stripCRLF(%q) = %q, want %q", in, got, want)
	}
}
