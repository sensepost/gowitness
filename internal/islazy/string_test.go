package islazy

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestTruncateLeavesShortStringsAlone(t *testing.T) {
	if got := Truncate("short", 200); got != "short" {
		t.Errorf("expected the input unchanged, got %q", got)
	}
}

func TestTruncateBoundsTheLength(t *testing.T) {
	if got := Truncate(strings.Repeat("a", 5000), 200); len(got) > 200 {
		t.Errorf("expected at most 200 bytes, got %d", len(got))
	}
}

func TestTruncateKeepsTheStartOfTheString(t *testing.T) {
	got := Truncate("https---example.com-443"+strings.Repeat("a", 500), 200)
	if !strings.HasPrefix(got, "https---example.com-443") {
		t.Errorf("expected the identifying prefix to survive, got %q", got)
	}
}

func TestTruncateSeparatesValuesSharingAPrefix(t *testing.T) {
	prefix := strings.Repeat("a", 300)

	if Truncate(prefix+"one", 200) == Truncate(prefix+"two", 200) {
		t.Error("expected values differing past the cut to truncate differently")
	}
}

func TestTruncateDoesNotSplitRunes(t *testing.T) {
	if got := Truncate(strings.Repeat("ä", 500), 200); !utf8.ValidString(got) {
		t.Errorf("expected valid utf8, got %q", got)
	}
}
