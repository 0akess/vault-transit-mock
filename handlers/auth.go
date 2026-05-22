package handlers

import (
	"net/http"
	"time"
)

const defaultLeaseSeconds = 3600

func AppRoleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")

		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"auth": map[string]any{
			"client_token":   randomToken(),
			"accessor":       randomToken(),
			"policies":       []string{"default"},
			"token_policies": []string{"default"},
			"lease_duration": defaultLeaseSeconds,
			"renewable":      true,
		},
	})
}

func TokenLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")

		return
	}

	tok := r.Header.Get("X-Vault-Token")
	if tok == "" {
		tok = randomToken()
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"auth": map[string]any{
			"client_token":   tok,
			"accessor":       randomToken(),
			"policies":       []string{"default"},
			"token_policies": []string{"default"},
			"lease_duration": defaultLeaseSeconds,
			"renewable":      true,
		},
	})
}

func TokenLookupSelf(w http.ResponseWriter, r *http.Request) {
	tok := r.Header.Get("X-Vault-Token")
	if tok == "" {
		tok = "mock-token"
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"id":             tok,
			"accessor":       randomToken(),
			"ttl":            defaultLeaseSeconds,
			"creation_time":  time.Now().UTC().Unix(),
			"creation_ttl":   defaultLeaseSeconds,
			"display_name":   "mock",
			"policies":       []string{"default"},
			"token_policies": []string{"default"},
			"renewable":      true,
		},
	})
}

func TokenRenewSelf(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")

		return
	}

	tok := r.Header.Get("X-Vault-Token")
	if tok == "" {
		tok = randomToken()
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"auth": map[string]any{
			"client_token":   tok,
			"accessor":       randomToken(),
			"policies":       []string{"default"},
			"token_policies": []string{"default"},
			"lease_duration": defaultLeaseSeconds,
			"renewable":      true,
		},
	})
}
