// Package cache provides a tiny file-based output cache for statusline
// invocations. The same statusline gets called every few seconds by shell
// prompts and statuslines, so we cache the rendered output to keep latency
// off the GitHub API hot path.
package cache

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type entry struct {
	Output    string    `json:"output"`
	FetchedAt time.Time `json:"fetched_at"`
}

// Cache reads and writes a single string value keyed by a directory path.
type Cache struct {
	dir string
}

// New returns a cache rooted at $TMPDIR/gh-statusline.
func New() *Cache {
	return &Cache{dir: filepath.Join(os.TempDir(), "gh-statusline")}
}

// Path returns the cache file path for a given key (typically the cwd).
func (c *Cache) Path(key string) string {
	h := sha256.Sum256([]byte(key))
	return filepath.Join(c.dir, fmt.Sprintf("pr-%x.json", h[:8]))
}

// Read returns the cached output and whether it was fresh (within ttl).
// A second return value reports whether any cached value exists at all
// (used for stale-on-error fallback).
func (c *Cache) Read(key string, ttl time.Duration) (output string, fresh, present bool) {
	data, err := os.ReadFile(c.Path(key))
	if err != nil {
		return "", false, false
	}
	var e entry
	if json.Unmarshal(data, &e) != nil {
		return "", false, false
	}
	return e.Output, time.Since(e.FetchedAt) < ttl, true
}

// Write stores output in the cache, creating the directory if needed.
// Write errors are ignored — the cache is best-effort.
func (c *Cache) Write(key, output string) {
	if err := os.MkdirAll(c.dir, 0o755); err != nil {
		return
	}
	data, _ := json.Marshal(entry{Output: output, FetchedAt: time.Now()})
	_ = os.WriteFile(c.Path(key), data, 0o644)
}
