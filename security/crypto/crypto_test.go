package crypto

import (
	"testing"

	utilsCrypto "github.com/tx7do/go-utils/crypto"
)

func TestTypeAlias_Cipher(t *testing.T) {
	key := []byte("1234567890abcdef") // 16 bytes
	c := utilsCrypto.NewAESCipher(key, nil)

	// Verify the alias works — c is a Cipher
	var _ Cipher = c
	if c.Name() == "" {
		t.Error("expected non-empty name")
	}
}

// ---------------------------------------------------------------------------
// Registry — RegisterCipher / GetCipher
// ---------------------------------------------------------------------------

func TestRegisterCipher_AES(t *testing.T) {
	key := []byte("1234567890abcdef")
	c := utilsCrypto.NewAESCipher(key, nil)

	RegisterCipher(c)

	got := GetCipher(c.Name())
	if got == nil {
		t.Fatal("GetCipher returned nil after registration")
	}

	// Encrypt + Decrypt round trip via registry lookup
	plain := []byte("hello, crypto registry!")
	encrypted, err := got.Encrypt(plain)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	decrypted, err := got.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if string(decrypted) != string(plain) {
		t.Errorf("round-trip mismatch: got %q, want %q", decrypted, plain)
	}
}

func TestGetCipher_NotFound(t *testing.T) {
	if c := GetCipher("nonexistent"); c != nil {
		t.Error("expected nil for unknown cipher name")
	}
}

func TestGetCipher_EmptyName(t *testing.T) {
	if c := GetCipher(""); c != nil {
		t.Error("expected nil for empty name")
	}
}

func TestRegisterCipher_Nil(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil cipher")
		}
	}()
	RegisterCipher(nil)
}

func TestRegisterCipher_CaseInsensitive(t *testing.T) {
	key := []byte("1234567890abcdef")
	c := utilsCrypto.NewAESCipher(key, nil)
	RegisterCipher(c)

	name := c.Name()
	// Look up with different casing
	upper := ""
	for _, ch := range name {
		if ch >= 'a' && ch <= 'z' {
			upper += string(ch - 32)
		} else {
			upper += string(ch)
		}
	}
	if got := GetCipher(upper); got == nil {
		t.Errorf("GetCipher(%q) returned nil, lookup should be case-insensitive", upper)
	}
}

// ---------------------------------------------------------------------------
// AES round-trip end-to-end
// ---------------------------------------------------------------------------

func TestAES_RoundTrip(t *testing.T) {
	key := []byte("1234567890abcdef")
	c := utilsCrypto.NewAESCipher(key, nil)

	plain := []byte("secret message")
	encrypted, err := c.Encrypt(plain)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if string(encrypted) == string(plain) {
		t.Error("encrypted data should differ from plaintext")
	}

	decrypted, err := c.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if string(decrypted) != string(plain) {
		t.Errorf("decrypted = %q, want %q", decrypted, plain)
	}
}
