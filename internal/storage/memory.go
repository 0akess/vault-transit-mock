// Package storage provides an in-memory, concurrent-safe KV v2 store.
// State is process-local and lost on restart — this matches the
// "dev-only" contract of the surrounding mock.
package storage

import (
	"sync"
	"time"
)

// Version is a single revision of a secret as stored under a path.
type Version struct {
	Data        map[string]any
	Version     int
	CreatedTime time.Time
	Destroyed   bool
}

// Store is a thread-safe KV v2 backend. Each path holds an ordered
// slice of Version values; the latest version is at index len-1.
type Store struct {
	mu   sync.RWMutex
	data map[string][]Version
}

// New returns an empty Store.
func New() *Store {
	return &Store{data: make(map[string][]Version)}
}

// Put appends a new version for the given path and returns it.
func (s *Store) Put(path string, secret map[string]any) Version {
	s.mu.Lock()
	defer s.mu.Unlock()

	versions := s.data[path]
	v := Version{
		Data:        secret,
		Version:     len(versions) + 1,
		CreatedTime: time.Now().UTC(),
	}
	s.data[path] = append(versions, v)

	return v
}

// Get returns the latest non-destroyed version and true if found.
func (s *Store) Get(path string) (Version, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	versions := s.data[path]
	if len(versions) == 0 {
		return Version{}, false
	}

	return versions[len(versions)-1], true
}

// Delete removes all versions for a path. Returns true if anything
// was removed.
func (s *Store) Delete(path string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.data[path]; !ok {
		return false
	}

	delete(s.data, path)

	return true
}

// List returns the immediate child names under the given prefix.
// Vault represents directories with a trailing slash; we mirror that.
// A prefix of "" enumerates everything at the root.
func (s *Store) List(prefix string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if prefix != "" && prefix[len(prefix)-1] != '/' {
		prefix += "/"
	}

	seen := make(map[string]struct{})

	for k := range s.data {
		if prefix != "" && (len(k) < len(prefix) || k[:len(prefix)] != prefix) {
			continue
		}

		rest := k[len(prefix):]
		if rest == "" {
			continue
		}
		// emit first segment; suffix slash for nested directories
		idx := -1

		for i := range len(rest) {
			if rest[i] == '/' {
				idx = i

				break
			}
		}

		if idx == -1 {
			seen[rest] = struct{}{}
		} else {
			seen[rest[:idx+1]] = struct{}{}
		}
	}

	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}

	return out
}
