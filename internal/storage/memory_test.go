package storage

import (
	"fmt"
	"sort"
	"sync"
	"testing"
)

func TestPutGet(t *testing.T) {
	s := New()
	s.Put("a", map[string]any{"k": "v1"})

	v, ok := s.Get("a")
	if !ok {
		t.Fatal("expected to find a")
	}

	if v.Version != 1 || v.Data["k"] != "v1" {
		t.Fatalf("unexpected version: %+v", v)
	}
}

func TestPut_MultipleVersions(t *testing.T) {
	s := New()
	s.Put("p", map[string]any{"k": "v1"})
	s.Put("p", map[string]any{"k": "v2"})

	v, _ := s.Get("p")
	if v.Version != 2 || v.Data["k"] != "v2" {
		t.Fatalf("expected version 2 with v2, got %+v", v)
	}
}

func TestGet_Missing(t *testing.T) {
	s := New()
	if _, ok := s.Get("nope"); ok {
		t.Fatal("expected not found")
	}
}

func TestDelete(t *testing.T) {
	s := New()
	s.Put("x", map[string]any{"a": 1})

	if !s.Delete("x") {
		t.Fatal("delete should report removed")
	}

	if _, ok := s.Get("x"); ok {
		t.Fatal("expected gone after delete")
	}

	if s.Delete("x") {
		t.Fatal("second delete should be false")
	}
}

func TestList(t *testing.T) {
	s := New()
	s.Put("apps/foo", map[string]any{"k": "v"})
	s.Put("apps/bar", map[string]any{"k": "v"})
	s.Put("apps/sub/baz", map[string]any{"k": "v"})
	s.Put("other", map[string]any{"k": "v"})

	got := s.List("apps")
	sort.Strings(got)

	want := []string{"bar", "foo", "sub/"}
	if len(got) != len(want) {
		t.Fatalf("expected %v got %v", want, got)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected %v got %v", want, got)
		}
	}

	root := s.List("")
	sort.Strings(root)
	// root contains "apps/" and "other"
	if len(root) != 2 {
		t.Fatalf("root list unexpected: %v", root)
	}
}

func TestList_TrailingSlash(t *testing.T) {
	s := New()
	s.Put("apps/foo", map[string]any{"k": "v"})

	got := s.List("apps/")
	if len(got) != 1 || got[0] != "foo" {
		t.Fatalf("expected [foo], got %v", got)
	}
}

func TestConcurrentWrites(t *testing.T) {
	// Spawn many goroutines writing distinct keys; verify all are
	// readable afterwards.
	s := New()

	const n = 100

	var wg sync.WaitGroup
	wg.Add(n)

	for i := range n {
		go func(i int) {
			defer wg.Done()

			s.Put(fmt.Sprintf("k%d", i), map[string]any{"i": i})
		}(i)
	}

	wg.Wait()

	for i := range n {
		v, ok := s.Get(fmt.Sprintf("k%d", i))
		if !ok {
			t.Fatalf("missing k%d", i)
		}

		if v.Data["i"] != i {
			t.Fatalf("wrong value for k%d: %v", i, v.Data)
		}
	}
}

func TestConcurrentSameKey(t *testing.T) {
	// Many writers on the same key must each get a unique version
	// number and the latest version count must equal write count.
	s := New()

	const n = 50

	var wg sync.WaitGroup
	wg.Add(n)

	for i := range n {
		go func(i int) {
			defer wg.Done()

			s.Put("hot", map[string]any{"i": i})
		}(i)
	}

	wg.Wait()

	v, _ := s.Get("hot")
	if v.Version != n {
		t.Fatalf("expected version %d got %d", n, v.Version)
	}
}
