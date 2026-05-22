package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newKV(t *testing.T) *KV {
	t.Helper()

	return NewKV()
}

func TestKV_WriteRead(t *testing.T) {
	h := newKV(t)
	body := `{"data":{"foo":"bar","n":1}}`
	req := httptest.NewRequest(http.MethodPost, "/v1/secret/data/myapp", strings.NewReader(body))
	w := httptest.NewRecorder()
	h.Data(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("write: %d body=%s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/secret/data/myapp", http.NoBody)
	w = httptest.NewRecorder()
	h.Data(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("read: %d", w.Code)
	}

	var out map[string]any

	_ = json.Unmarshal(w.Body.Bytes(), &out)
	data := out["data"].(map[string]any)

	inner := data["data"].(map[string]any)
	if inner["foo"] != "bar" {
		t.Fatalf("data mismatch: %v", inner)
	}

	meta := data["metadata"].(map[string]any)
	for _, k := range []string{"version", "created_time", "deletion_time", "destroyed"} {
		if _, ok := meta[k]; !ok {
			t.Fatalf("missing meta field %q in %v", k, meta)
		}
	}
}

func TestKV_Read_NotFound(t *testing.T) {
	h := newKV(t)
	req := httptest.NewRequest(http.MethodGet, "/v1/secret/data/missing", http.NoBody)
	w := httptest.NewRecorder()
	h.Data(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestKV_Write_BadBody(t *testing.T) {
	h := newKV(t)
	req := httptest.NewRequest(http.MethodPost, "/v1/secret/data/x", strings.NewReader(`xxx`))
	w := httptest.NewRecorder()
	h.Data(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestKV_Write_NoData(t *testing.T) {
	h := newKV(t)
	req := httptest.NewRequest(http.MethodPost, "/v1/secret/data/x", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	h.Data(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestKV_MissingPath(t *testing.T) {
	h := newKV(t)
	req := httptest.NewRequest(http.MethodGet, "/v1/secret/data/", http.NoBody)
	w := httptest.NewRecorder()
	h.Data(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestKV_Data_WrongMethod(t *testing.T) {
	h := newKV(t)
	req := httptest.NewRequest(http.MethodPatch, "/v1/secret/data/x", http.NoBody)
	w := httptest.NewRecorder()
	h.Data(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestKV_Data_Delete(t *testing.T) {
	h := newKV(t)
	body := `{"data":{"k":"v"}}`
	req := httptest.NewRequest(http.MethodPost, "/v1/secret/data/x", strings.NewReader(body))
	w := httptest.NewRecorder()
	h.Data(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("write failed: %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodDelete, "/v1/secret/data/x", http.NoBody)
	w = httptest.NewRecorder()
	h.Data(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodDelete, "/v1/secret/data/x", http.NoBody)
	w = httptest.NewRecorder()
	h.Data(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 on second delete, got %d", w.Code)
	}
}

func TestKV_Versions(t *testing.T) {
	h := newKV(t)

	for _, val := range []string{"a", "b", "c"} {
		req := httptest.NewRequest(http.MethodPost, "/v1/secret/data/v",
			strings.NewReader(`{"data":{"k":"`+val+`"}}`))
		w := httptest.NewRecorder()
		h.Data(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("write %s: %d", val, w.Code)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/secret/data/v", http.NoBody)
	w := httptest.NewRecorder()
	h.Data(w, req)

	var out map[string]any

	_ = json.Unmarshal(w.Body.Bytes(), &out)
	data := out["data"].(map[string]any)

	inner := data["data"].(map[string]any)
	if inner["k"] != "c" {
		t.Fatalf("expected latest c, got %v", inner["k"])
	}

	meta := data["metadata"].(map[string]any)
	if v, _ := meta["version"].(float64); int(v) != 3 {
		t.Fatalf("expected version 3, got %v", meta["version"])
	}
}

func TestKV_Metadata_ListHTTPMethod(t *testing.T) {
	h := newKV(t)

	for _, p := range []string{"apps/foo", "apps/bar"} {
		req := httptest.NewRequest(http.MethodPost, "/v1/secret/data/"+p,
			strings.NewReader(`{"data":{"k":"v"}}`))
		w := httptest.NewRecorder()
		h.Data(w, req)
	}

	req := httptest.NewRequest("LIST", "/v1/secret/metadata/apps", http.NoBody)
	w := httptest.NewRecorder()
	h.Metadata(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("list: %d body=%s", w.Code, w.Body.String())
	}

	var out map[string]any

	_ = json.Unmarshal(w.Body.Bytes(), &out)

	keys := out["data"].(map[string]any)["keys"].([]any)
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %v", keys)
	}
}

func TestKV_Metadata_ListQuery(t *testing.T) {
	h := newKV(t)
	req := httptest.NewRequest(http.MethodPost, "/v1/secret/data/apps/foo",
		strings.NewReader(`{"data":{"k":"v"}}`))
	w := httptest.NewRecorder()
	h.Data(w, req)
	req = httptest.NewRequest(http.MethodGet, "/v1/secret/metadata/apps?list=true", http.NoBody)
	w = httptest.NewRecorder()
	h.Metadata(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestKV_Metadata_ListEmpty(t *testing.T) {
	h := newKV(t)
	req := httptest.NewRequest("LIST", "/v1/secret/metadata/empty", http.NoBody)
	w := httptest.NewRecorder()
	h.Metadata(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestKV_Metadata_Delete(t *testing.T) {
	h := newKV(t)
	req := httptest.NewRequest(http.MethodPost, "/v1/secret/data/x",
		strings.NewReader(`{"data":{"k":"v"}}`))
	w := httptest.NewRecorder()
	h.Data(w, req)
	req = httptest.NewRequest(http.MethodDelete, "/v1/secret/metadata/x", http.NoBody)
	w = httptest.NewRecorder()
	h.Metadata(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/secret/data/x", http.NoBody)
	w = httptest.NewRecorder()
	h.Data(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after metadata delete, got %d", w.Code)
	}
}

func TestKV_Metadata_DeleteMissing(t *testing.T) {
	h := newKV(t)
	req := httptest.NewRequest(http.MethodDelete, "/v1/secret/metadata/no-such", http.NoBody)
	w := httptest.NewRecorder()
	h.Metadata(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestKV_Metadata_DeleteMissingPath(t *testing.T) {
	h := newKV(t)
	req := httptest.NewRequest(http.MethodDelete, "/v1/secret/metadata/", http.NoBody)
	w := httptest.NewRecorder()
	h.Metadata(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestKV_Metadata_Get(t *testing.T) {
	h := newKV(t)
	req := httptest.NewRequest(http.MethodPost, "/v1/secret/data/x",
		strings.NewReader(`{"data":{"k":"v"}}`))
	w := httptest.NewRecorder()
	h.Data(w, req)

	req = httptest.NewRequest(http.MethodGet, "/v1/secret/metadata/x", http.NoBody)
	w = httptest.NewRecorder()
	h.Metadata(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var out map[string]any

	_ = json.Unmarshal(w.Body.Bytes(), &out)

	data := out["data"].(map[string]any)
	for _, k := range []string{"current_version", "created_time", "versions"} {
		if _, ok := data[k]; !ok {
			t.Fatalf("missing %q in %v", k, data)
		}
	}
}

func TestKV_Metadata_GetMissingPath(t *testing.T) {
	h := newKV(t)
	req := httptest.NewRequest(http.MethodGet, "/v1/secret/metadata/", http.NoBody)
	w := httptest.NewRecorder()
	h.Metadata(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestKV_Metadata_GetNotFound(t *testing.T) {
	h := newKV(t)
	req := httptest.NewRequest(http.MethodGet, "/v1/secret/metadata/missing", http.NoBody)
	w := httptest.NewRecorder()
	h.Metadata(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestKV_Metadata_WrongMethod(t *testing.T) {
	h := newKV(t)
	req := httptest.NewRequest(http.MethodPatch, "/v1/secret/metadata/x", http.NoBody)
	w := httptest.NewRecorder()
	h.Metadata(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}
