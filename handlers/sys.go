package handlers

import "net/http"

// Health responds to GET /v1/sys/health. Always reports unsealed
// and initialized, which is what dev clients expect from a happy mock.
func Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"initialized": true,
		"sealed":      false,
		"standby":     false,
		"version":     Version,
	})
}

// SealStatus responds to GET /v1/sys/seal-status with a synthetic
// "always-unsealed" payload.
func SealStatus(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"type":         "shamir",
		"initialized":  true,
		"sealed":       false,
		"t":            1,
		"n":            1,
		"progress":     0,
		"nonce":        "",
		"version":      Version,
		"cluster_name": "vault-transit-mock",
	})
}
