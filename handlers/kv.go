package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"vault-transit-mock/internal/storage"
)

// KV bundles the in-memory store with the HTTP entrypoints that
// implement Vault's KV v2 surface.
type KV struct {
	Store *storage.Store
}

// NewKV constructs a KV handler backed by a fresh storage.Store.
func NewKV() *KV {
	return &KV{Store: storage.New()}
}

func vaultTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

// Data serves /v1/secret/data/<path> — GET and POST.
func (h *KV) Data(w http.ResponseWriter, r *http.Request) {
	path := keyName(r.URL.Path, "/v1/secret/data/")
	if path == "" {
		writeError(w, http.StatusBadRequest, "missing path")

		return
	}

	switch r.Method {
	case http.MethodGet:
		h.read(w, path)
	case http.MethodPost, http.MethodPut:
		h.write(w, r, path)
	case http.MethodDelete:
		// Soft-delete latest: mirror Vault by removing all versions.
		// MVP keeps deletion semantics identical to the metadata route.
		if !h.Store.Delete(path) {
			writeError(w, http.StatusNotFound, "not found")

			return
		}

		w.WriteHeader(http.StatusNoContent)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// Metadata serves /v1/secret/metadata/<path> — LIST and DELETE.
func (h *KV) Metadata(w http.ResponseWriter, r *http.Request) {
	path := keyName(r.URL.Path, "/v1/secret/metadata/")
	// LIST is allowed with empty path (lists root).
	switch {
	case r.Method == "LIST" || (r.Method == http.MethodGet && r.URL.Query().Get("list") == "true"):
		h.list(w, path)
	case r.Method == http.MethodDelete:
		if path == "" {
			writeError(w, http.StatusBadRequest, "missing path")

			return
		}

		if !h.Store.Delete(path) {
			writeError(w, http.StatusNotFound, "not found")

			return
		}

		w.WriteHeader(http.StatusNoContent)
	case r.Method == http.MethodGet:
		// GET on metadata without ?list=true → return latest version meta.
		if path == "" {
			writeError(w, http.StatusBadRequest, "missing path")

			return
		}

		v, ok := h.Store.Get(path)
		if !ok {
			writeError(w, http.StatusNotFound, "not found")

			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"data": map[string]any{
				"current_version": v.Version,
				"oldest_version":  1,
				"created_time":    vaultTime(v.CreatedTime),
				"updated_time":    vaultTime(v.CreatedTime),
				"versions": map[string]any{
					strconv.Itoa(v.Version): map[string]any{
						"created_time":  vaultTime(v.CreatedTime),
						"deletion_time": "",
						"destroyed":     false,
					},
				},
			},
		})
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *KV) read(w http.ResponseWriter, path string) {
	v, ok := h.Store.Get(path)
	if !ok {
		writeError(w, http.StatusNotFound, "not found")

		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"data": v.Data,
			"metadata": map[string]any{
				"version":       v.Version,
				"created_time":  vaultTime(v.CreatedTime),
				"deletion_time": "",
				"destroyed":     v.Destroyed,
			},
		},
	})
}

func (h *KV) write(w http.ResponseWriter, r *http.Request, path string) {
	var req struct {
		Data map[string]any `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")

		return
	}

	if req.Data == nil {
		writeError(w, http.StatusBadRequest, "missing data field")

		return
	}

	v := h.Store.Put(path, req.Data)
	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"version":       v.Version,
			"created_time":  vaultTime(v.CreatedTime),
			"deletion_time": "",
			"destroyed":     false,
		},
	})
}

func (h *KV) list(w http.ResponseWriter, path string) {
	keys := h.Store.List(path)
	if len(keys) == 0 {
		writeError(w, http.StatusNotFound, "not found")

		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"keys": keys,
		},
	})
}
