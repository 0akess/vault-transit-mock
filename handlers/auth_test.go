package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func decodeAuth(t *testing.T, body []byte) map[string]any {
	t.Helper()

	var r map[string]any
	if err := json.Unmarshal(body, &r); err != nil {
		t.Fatalf("decode: %v", err)
	}

	auth, ok := r["auth"].(map[string]any)
	if !ok {
		t.Fatalf("missing auth: %v", r)
	}

	return auth
}

func TestAppRoleLogin(t *testing.T) {
	for _, method := range []string{http.MethodPost, http.MethodPut} {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/v1/auth/approle/login",
				strings.NewReader(`{"role_id":"a","secret_id":"b"}`))
			w := httptest.NewRecorder()
			AppRoleLogin(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("status: got %d want 200", w.Code)
			}

			auth := decodeAuth(t, w.Body.Bytes())
			for _, k := range []string{"client_token", "accessor", "policies", "lease_duration", "renewable"} {
				if _, ok := auth[k]; !ok {
					t.Fatalf("missing %q in %v", k, auth)
				}
			}
		})
	}
}

func TestAppRoleLogin_WrongMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/auth/approle/login", http.NoBody)
	w := httptest.NewRecorder()
	AppRoleLogin(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestTokenLogin(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/token/login", http.NoBody)
	req.Header.Set("X-Vault-Token", "supplied-token")

	w := httptest.NewRecorder()
	TokenLogin(w, req)

	auth := decodeAuth(t, w.Body.Bytes())
	if auth["client_token"] != "supplied-token" {
		t.Fatalf("expected token to be echoed, got %v", auth["client_token"])
	}
}

func TestTokenLogin_NoHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/token/login", http.NoBody)
	w := httptest.NewRecorder()
	TokenLogin(w, req)

	auth := decodeAuth(t, w.Body.Bytes())
	if s, _ := auth["client_token"].(string); s == "" {
		t.Fatal("expected random token when header absent")
	}
}

func TestTokenLogin_WrongMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/auth/token/login", http.NoBody)
	w := httptest.NewRecorder()
	TokenLogin(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestTokenLookupSelf(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/auth/token/lookup-self", http.NoBody)
	req.Header.Set("X-Vault-Token", "abc")

	w := httptest.NewRecorder()
	TokenLookupSelf(w, req)

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("missing data: %v", body)
	}

	if data["id"] != "abc" {
		t.Fatalf("expected id=abc, got %v", data["id"])
	}

	for _, k := range []string{"accessor", "ttl", "creation_time", "policies"} {
		if _, ok := data[k]; !ok {
			t.Fatalf("missing %q", k)
		}
	}
}

func TestTokenLookupSelf_NoHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/auth/token/lookup-self", http.NoBody)
	w := httptest.NewRecorder()
	TokenLookupSelf(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestTokenRenewSelf(t *testing.T) {
	for _, method := range []string{http.MethodPost, http.MethodPut} {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/v1/auth/token/renew-self", http.NoBody)
			req.Header.Set("X-Vault-Token", "tok")

			w := httptest.NewRecorder()
			TokenRenewSelf(w, req)

			auth := decodeAuth(t, w.Body.Bytes())
			if auth["client_token"] != "tok" {
				t.Fatalf("expected token to be echoed: %v", auth)
			}

			if auth["renewable"] != true {
				t.Fatalf("expected renewable=true, got %v", auth["renewable"])
			}
		})
	}
}

func TestTokenRenewSelf_NoHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/token/renew-self", http.NoBody)
	w := httptest.NewRecorder()
	TokenRenewSelf(w, req)

	auth := decodeAuth(t, w.Body.Bytes())
	if s, _ := auth["client_token"].(string); s == "" {
		t.Fatal("expected random token")
	}
}

func TestTokenRenewSelf_WrongMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/auth/token/renew-self", http.NoBody)
	w := httptest.NewRecorder()
	TokenRenewSelf(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}
