package cipher

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
	"testing"
)

func TestEncrypt_HasPrefix(t *testing.T) {
	ct, err := Encrypt(base64.StdEncoding.EncodeToString([]byte("hello")))
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	if !strings.HasPrefix(ct, Prefix) {
		t.Fatalf("ciphertext %q lacks prefix %q", ct, Prefix)
	}
}

func TestEncrypt_Deterministic(t *testing.T) {
	// Property: encrypting the same plaintext repeatedly produces
	// identical ciphertext.
	for i := range 100 {
		buf := make([]byte, 1+i%64)
		if _, err := rand.Read(buf); err != nil {
			t.Fatalf("rand: %v", err)
		}

		pt := base64.StdEncoding.EncodeToString(buf)

		a, err := Encrypt(pt)
		if err != nil {
			t.Fatalf("encrypt a: %v", err)
		}

		b, err := Encrypt(pt)
		if err != nil {
			t.Fatalf("encrypt b: %v", err)
		}

		if a != b {
			t.Fatalf("non-deterministic encrypt: %q vs %q", a, b)
		}
	}
}

func TestEncryptDecrypt_Roundtrip(t *testing.T) {
	for i := range 100 {
		buf := make([]byte, 1+i%128)
		if _, err := rand.Read(buf); err != nil {
			t.Fatalf("rand: %v", err)
		}

		pt := base64.StdEncoding.EncodeToString(buf)

		ct, err := Encrypt(pt)
		if err != nil {
			t.Fatalf("encrypt: %v", err)
		}

		got, err := Decrypt(ct)
		if err != nil {
			t.Fatalf("decrypt: %v", err)
		}

		if got != pt {
			t.Fatalf("roundtrip mismatch: got %q want %q", got, pt)
		}
	}
}

func TestEncrypt_BadBase64(t *testing.T) {
	if _, err := Encrypt("!!!not base64!!!"); err == nil {
		t.Fatal("expected error for bad plaintext base64")
	}
}

func TestDecrypt_NoPrefix(t *testing.T) {
	if _, err := Decrypt("aGVsbG8="); err == nil {
		t.Fatal("expected error when prefix missing")
	}
}

func TestDecrypt_BadBase64Payload(t *testing.T) {
	if _, err := Decrypt(Prefix + "!!!"); err == nil {
		t.Fatal("expected error for malformed payload")
	}
}
