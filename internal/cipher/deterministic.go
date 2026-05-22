// Package cipher implements deterministic, non-cryptographic
// transformations that mimic the shape of HashiCorp Vault's transit
// engine. The "ciphertext" is just base64 of the plaintext wrapped in a
// "vault:v1:" prefix, which means encrypt is deterministic and decrypt
// is a perfect inverse — suitable for local development only.
package cipher

import (
	"encoding/base64"
	"errors"
	"strings"
)

// Prefix is the Vault transit ciphertext envelope marker.
const Prefix = "vault:v1:"

// ErrInvalidCiphertext is returned when the input does not have the
// expected "vault:v1:" prefix or its payload is not valid base64.
var ErrInvalidCiphertext = errors.New("invalid ciphertext")

// ErrInvalidPlaintext is returned when the plaintext is not valid base64.
var ErrInvalidPlaintext = errors.New("invalid plaintext")

// Encrypt takes a base64-encoded plaintext, decodes it, and re-wraps
// the raw bytes with the Vault envelope. The transformation is
// deterministic: same input → same output.
func Encrypt(plaintextB64 string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(plaintextB64)
	if err != nil {
		return "", ErrInvalidPlaintext
	}

	return Prefix + base64.StdEncoding.EncodeToString(raw), nil
}

// Decrypt reverses Encrypt: strips the "vault:v1:" envelope, validates
// the inner base64, and returns the original base64 plaintext.
func Decrypt(ciphertext string) (string, error) {
	if !strings.HasPrefix(ciphertext, Prefix) {
		return "", ErrInvalidCiphertext
	}

	payload := strings.TrimPrefix(ciphertext, Prefix)

	raw, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return "", ErrInvalidCiphertext
	}

	return base64.StdEncoding.EncodeToString(raw), nil
}
