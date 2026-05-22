package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"vault-transit-mock/internal/cipher"
)

// keyName extracts the trailing path segment from a URL like
// /v1/transit/encrypt/<name>. Returns empty string if missing.
func keyName(path, prefix string) string {
	rest := strings.TrimPrefix(path, prefix)
	rest = strings.Trim(rest, "/")

	return rest
}

// TransitKeys serves GET (metadata) and POST (create) on
// /v1/transit/keys/<name>. State is not persisted; metadata is
// synthetic.
func TransitKeys(w http.ResponseWriter, r *http.Request) {
	name := keyName(r.URL.Path, "/v1/transit/keys/")
	if name == "" {
		writeError(w, http.StatusBadRequest, "missing key name")

		return
	}

	switch r.Method {
	case http.MethodPost, http.MethodPut:
		w.WriteHeader(http.StatusNoContent)
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]any{
			"data": map[string]any{
				"name":                   name,
				"type":                   "aes256-gcm96",
				"latest_version":         1,
				"min_decryption_version": 1,
				"min_encryption_version": 0,
				"deletion_allowed":       false,
				"derived":                false,
				"exportable":             false,
				"keys": map[string]any{
					"1": time.Now().UTC().Unix(),
				},
			},
		})
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// TransitEncrypt handles POST and PUT /v1/transit/encrypt/<name>.
func TransitEncrypt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")

		return
	}

	if keyName(r.URL.Path, "/v1/transit/encrypt/") == "" {
		writeError(w, http.StatusBadRequest, "missing key name")

		return
	}

	var req struct {
		Plaintext string `json:"plaintext"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")

		return
	}

	if req.Plaintext == "" {
		writeError(w, http.StatusBadRequest, "missing plaintext")

		return
	}

	ct, err := cipher.Encrypt(req.Plaintext)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())

		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"ciphertext":  ct,
			"key_version": 1,
		},
	})
}

// TransitDecrypt handles POST and PUT /v1/transit/decrypt/<name>.
func TransitDecrypt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")

		return
	}

	if keyName(r.URL.Path, "/v1/transit/decrypt/") == "" {
		writeError(w, http.StatusBadRequest, "missing key name")

		return
	}

	var req struct {
		Ciphertext string `json:"ciphertext"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")

		return
	}

	if req.Ciphertext == "" {
		writeError(w, http.StatusBadRequest, "missing ciphertext")

		return
	}

	pt, err := cipher.Decrypt(req.Ciphertext)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())

		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"plaintext": pt,
		},
	})
}
