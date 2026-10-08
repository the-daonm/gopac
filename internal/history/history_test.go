package history

import (
	"fmt"
	"testing"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	if got := Load(); len(got) != 0 {
		t.Fatalf("expected empty history, got %v", got)
	}

	var entries []string
	for i := range MaxEntries + 5 {
		entries = append(entries, fmt.Sprintf("query%d", i))
	}
	if err := Save(entries); err != nil {
		t.Fatal(err)
	}

	got := Load()
	if len(got) != MaxEntries {
		t.Fatalf("expected %d entries, got %d", MaxEntries, len(got))
	}
	if got[0] != "query5" || got[len(got)-1] != fmt.Sprintf("query%d", MaxEntries+4) {
		t.Errorf("unexpected entries kept: first=%q last=%q", got[0], got[len(got)-1])
	}
}
