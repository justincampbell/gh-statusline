package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newAt(dir string) *Cache { return &Cache{dir: dir} }

func TestRoundtrip(t *testing.T) {
	c := newAt(t.TempDir())

	_, fresh, present := c.Read("/foo", 30*time.Second)
	if fresh || present {
		t.Fatal("expected no value on first read")
	}

	c.Write("/foo", "✓ #42")
	out, fresh, present := c.Read("/foo", 30*time.Second)
	if !fresh || !present {
		t.Fatal("expected fresh hit after Write")
	}
	if out != "✓ #42" {
		t.Errorf("got %q, want %q", out, "✓ #42")
	}
}

func TestStaleAfterTTL(t *testing.T) {
	dir := t.TempDir()
	c := newAt(dir)

	// Manually write an entry with an old timestamp.
	path := c.Path("/foo")
	data, _ := json.Marshal(entry{Output: "old", FetchedAt: time.Now().Add(-2 * time.Minute)})
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	out, fresh, present := c.Read("/foo", 30*time.Second)
	if fresh {
		t.Fatal("expected stale, not fresh")
	}
	if !present {
		t.Fatal("expected entry to be present even if stale")
	}
	if out != "old" {
		t.Errorf("expected stale value, got %q", out)
	}
}

func TestDifferentKeysProduceDifferentPaths(t *testing.T) {
	c := newAt(t.TempDir())
	if c.Path("/a") == c.Path("/b") {
		t.Fatal("different keys must produce different paths")
	}
	if c.Path("/a") != c.Path("/a") {
		t.Fatal("same key must produce same path")
	}
}

func TestPathInRoot(t *testing.T) {
	c := newAt("/tmp/test")
	p := c.Path("/foo")
	if filepath.Dir(p) != "/tmp/test" {
		t.Errorf("expected path under /tmp/test, got %q", p)
	}
}
