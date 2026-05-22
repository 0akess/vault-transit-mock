package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/sys/health", http.NoBody)
	w := httptest.NewRecorder()
	Health(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200", w.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	for _, k := range []string{"initialized", "sealed", "standby", "version"} {
		if _, ok := body[k]; !ok {
			t.Fatalf("missing field %q in %v", k, body)
		}
	}

	if body["initialized"] != true || body["sealed"] != false {
		t.Fatalf("unexpected health flags: %v", body)
	}
}

func TestSealStatus(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/sys/seal-status", http.NoBody)
	w := httptest.NewRecorder()
	SealStatus(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200", w.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if body["sealed"] != false {
		t.Fatalf("expected sealed=false, got %v", body["sealed"])
	}

	for _, k := range []string{"type", "initialized", "sealed", "version"} {
		if _, ok := body[k]; !ok {
			t.Fatalf("missing field %q", k)
		}
	}
}
