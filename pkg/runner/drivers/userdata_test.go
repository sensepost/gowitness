package driver

import (
	"os"
	"testing"
)

func TestResolveUserDataDirCreatesTempWhenEmpty(t *testing.T) {
	dir, ephemeral, err := resolveUserDataDir("", "gowitness-test-*")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ephemeral {
		t.Fatalf("expected an ephemeral directory when none was configured")
	}
	if dir == "" {
		t.Fatal("expected a temporary directory path, got an empty string")
	}
	defer os.RemoveAll(dir)

	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("expected the temporary directory to exist: %v", err)
	}
}

func TestResolveUserDataDirKeepsConfigured(t *testing.T) {
	// a directory the user supplied should be returned untouched and never
	// flagged for cleanup, so we don't wipe someone's real Chrome profile.
	configured := t.TempDir()

	dir, ephemeral, err := resolveUserDataDir(configured, "gowitness-test-*")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ephemeral {
		t.Fatal("a configured directory should not be marked ephemeral")
	}
	if dir != configured {
		t.Fatalf("expected %q, got %q", configured, dir)
	}
}
