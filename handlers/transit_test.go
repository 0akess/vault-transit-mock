package handlers

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTransitKeys_Create(t *testing.T) {
	for _, method := range []string{http.MethodPost, http.MethodPut} {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/v1/transit/keys/foo", http.NoBody)
			w := httptest.NewRecorder()
			TransitKeys(w, req)

			if w.Code != http.StatusNoContent {
				t.Fatalf("status: got %d want 204", w.Code)
			}
		})
	}
}

func TestTransitKeys_Get(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/transit/keys/foo", http.NoBody)
	w := httptest.NewRecorder()
	TransitKeys(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: %d", w.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("missing data: %v", body)
	}

	if data["name"] != "foo" || data["type"] != "aes256-gcm96" {
		t.Fatalf("unexpected key metadata: %v", data)
	}

	if _, ok := data["latest_version"]; !ok {
		t.Fatalf("missing latest_version")
	}
}

func TestTransitKeys_MissingName(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/transit/keys/", http.NoBody)
	w := httptest.NewRecorder()
	TransitKeys(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestTransitKeys_WrongMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/v1/transit/keys/foo", http.NoBody)
	w := httptest.NewRecorder()
	TransitKeys(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestTransitEncrypt(t *testing.T) {
	pt := base64.StdEncoding.EncodeToString([]byte("hello"))

	for _, method := range []string{http.MethodPost, http.MethodPut} {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/v1/transit/encrypt/foo",
				strings.NewReader(`{"plaintext":"`+pt+`"}`))
			w := httptest.NewRecorder()
			TransitEncrypt(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("status: %d body=%s", w.Code, w.Body.String())
			}

			var body map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode: %v", err)
			}

			data := body["data"].(map[string]any)

			ct, _ := data["ciphertext"].(string)
			if !strings.HasPrefix(ct, "vault:v1:") {
				t.Fatalf("bad ciphertext: %q", ct)
			}

			if _, ok := data["key_version"]; !ok {
				t.Fatalf("missing key_version")
			}
		})
	}
}

func TestTransitEncrypt_BadBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/transit/encrypt/foo",
		strings.NewReader(`{not json`))
	w := httptest.NewRecorder()
	TransitEncrypt(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestTransitEncrypt_EmptyPlaintext(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/transit/encrypt/foo",
		strings.NewReader(`{"plaintext":""}`))
	w := httptest.NewRecorder()
	TransitEncrypt(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestTransitEncrypt_BadPlaintextBase64(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/transit/encrypt/foo",
		strings.NewReader(`{"plaintext":"!!!"}`))
	w := httptest.NewRecorder()
	TransitEncrypt(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestTransitEncrypt_MissingName(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/transit/encrypt/",
		strings.NewReader(`{"plaintext":"aGVsbG8="}`))
	w := httptest.NewRecorder()
	TransitEncrypt(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestTransitEncrypt_WrongMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/transit/encrypt/foo", http.NoBody)
	w := httptest.NewRecorder()
	TransitEncrypt(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestTransitDecrypt_Roundtrip(t *testing.T) {
	pt := base64.StdEncoding.EncodeToString([]byte("roundtrip-payload"))

	encReq := httptest.NewRequest(http.MethodPost, "/v1/transit/encrypt/foo",
		strings.NewReader(`{"plaintext":"`+pt+`"}`))
	encW := httptest.NewRecorder()
	TransitEncrypt(encW, encReq)

	if encW.Code != http.StatusOK {
		t.Fatalf("encrypt failed: %d", encW.Code)
	}

	var encBody map[string]any

	_ = json.Unmarshal(encW.Body.Bytes(), &encBody)
	ct := encBody["data"].(map[string]any)["ciphertext"].(string)

	for _, method := range []string{http.MethodPost, http.MethodPut} {
		t.Run(method, func(t *testing.T) {
			decReq := httptest.NewRequest(method, "/v1/transit/decrypt/foo",
				strings.NewReader(`{"ciphertext":"`+ct+`"}`))
			decW := httptest.NewRecorder()
			TransitDecrypt(decW, decReq)

			if decW.Code != http.StatusOK {
				t.Fatalf("decrypt: %d body=%s", decW.Code, decW.Body.String())
			}

			var decBody map[string]any

			_ = json.Unmarshal(decW.Body.Bytes(), &decBody)
			if decBody["data"].(map[string]any)["plaintext"] != pt {
				t.Fatalf("plaintext mismatch")
			}
		})
	}
}

func TestTransitDecrypt_BadBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/transit/decrypt/foo",
		strings.NewReader(`oops`))
	w := httptest.NewRecorder()
	TransitDecrypt(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestTransitDecrypt_Empty(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/transit/decrypt/foo",
		strings.NewReader(`{"ciphertext":""}`))
	w := httptest.NewRecorder()
	TransitDecrypt(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestTransitDecrypt_BadCiphertext(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/transit/decrypt/foo",
		strings.NewReader(`{"ciphertext":"not-prefixed"}`))
	w := httptest.NewRecorder()
	TransitDecrypt(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestTransitDecrypt_MissingName(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/transit/decrypt/",
		strings.NewReader(`{"ciphertext":"vault:v1:aGVsbG8="}`))
	w := httptest.NewRecorder()
	TransitDecrypt(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestTransitDecrypt_WrongMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/transit/decrypt/foo", http.NoBody)
	w := httptest.NewRecorder()
	TransitDecrypt(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}
