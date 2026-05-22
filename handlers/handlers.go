// Package handlers contains HTTP handlers that implement a
// drop-in-compatible subset of HashiCorp Vault's HTTP API.
package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
)

const tokenBytes = 16

// Version is the mock's reported version, surfaced via /v1/sys/health.
// It is settable from main at startup; tests can override.
var Version = "mock-1.0.0"

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	//nolint:errcheck // headers are already flushed; no useful recovery for a mock response
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]any{"errors": []string{msg}})
}

func randomToken() string {
	b := make([]byte, tokenBytes)
	_, _ = rand.Read(b)

	return hex.EncodeToString(b)
}
