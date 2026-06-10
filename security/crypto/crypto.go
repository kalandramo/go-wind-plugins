// Package crypto provides a pluggable cryptography abstraction for the
// go-wind framework.
//
// It re-exports the interfaces defined in [github.com/tx7do/go-utils/crypto]
// as type aliases, so that all modules within go-wind-plugins share a single
// stable reference point. It also provides a global registry (RegisterCipher /
// GetCipher) that mirrors the [encoding] package's RegisterCodec / GetCodec,
// enabling runtime lookup of crypto implementations by name.
//
// Available implementations (from go-utils/crypto):
//
//   - AES (NewAESCipher)
//   - RSA (NewRSACipher)
//   - SM2 (NewSM2Cipher)
//   - SM4 (NewSM4Cipher)
//   - HMAC (NewHMAC)
//   - SM3 / SHA256 (NewSM3Hasher)
//   - ECDSA (NewECDSASigner / NewECDSAVerifier)
//   - ECDH (NewECDH)
//
// Usage:
//
//	cipher, _ := crypto.NewAESCipher(key, nil)
//	encrypted, _ := cipher.Encrypt(plaintext)
//	decrypted, _ := cipher.Decrypt(encrypted)
//
// Registration and lookup:
//
//	crypto.RegisterCipher(cipher)
//	c := crypto.GetCipher("aes")
package crypto

import (
	"strings"
	"sync"

	"github.com/tx7do/go-utils/crypto"
)

// ---------------------------------------------------------------------------
// Interface aliases — re-exported from go-utils/crypto
// ---------------------------------------------------------------------------

// Cipher defines symmetric/asymmetric encryption-decryption.
type Cipher = crypto.Cipher

// Hasher defines hash computation (SHA256, SM3, …).
type Hasher = crypto.Hasher

// Signer defines digital signature generation (ECDSA, SM2, …).
type Signer = crypto.Signer

// Verifier defines digital signature verification.
type Verifier = crypto.Verifier

// KeyExchanger defines key agreement (ECDH, SM2 key exchange, …).
type KeyExchanger = crypto.KeyExchanger

// ---------------------------------------------------------------------------
// Global registry — mirrors encoding.RegisterCodec / GetCodec
// ---------------------------------------------------------------------------

var (
	cipherMu sync.RWMutex
	ciphers  = map[string]Cipher{}

	hasherMu sync.RWMutex
	hashers  = map[string]Hasher{}

	signerMu sync.RWMutex
	signers  = map[string]Signer{}

	verifierMu sync.RWMutex
	verifiers  = map[string]Verifier{}
)

// RegisterCipher registers a Cipher under its Name().
// The name is case-insensitive and stored in lower-case.
func RegisterCipher(c Cipher) {
	if c == nil {
		panic("crypto: cipher cannot be nil")
	}
	name := strings.ToLower(c.Name())
	if name == "" {
		panic("crypto: cipher name cannot be empty")
	}
	cipherMu.Lock()
	ciphers[name] = c
	cipherMu.Unlock()
}

// GetCipher returns the Cipher registered under name, or nil if not found.
// The name is case-insensitive.
func GetCipher(name string) Cipher {
	if name == "" {
		return nil
	}
	name = strings.ToLower(name)
	cipherMu.RLock()
	c := ciphers[name]
	cipherMu.RUnlock()
	return c
}

// RegisterHasher registers a Hasher under its Name().
func RegisterHasher(h Hasher) {
	if h == nil {
		panic("crypto: hasher cannot be nil")
	}
	name := strings.ToLower(h.Name())
	if name == "" {
		panic("crypto: hasher name cannot be empty")
	}
	hasherMu.Lock()
	hashers[name] = h
	hasherMu.Unlock()
}

// GetHasher returns the Hasher registered under name, or nil if not found.
func GetHasher(name string) Hasher {
	if name == "" {
		return nil
	}
	name = strings.ToLower(name)
	hasherMu.RLock()
	h := hashers[name]
	hasherMu.RUnlock()
	return h
}

// RegisterSigner registers a Signer under its Name().
func RegisterSigner(s Signer) {
	if s == nil {
		panic("crypto: signer cannot be nil")
	}
	name := strings.ToLower(s.Name())
	if name == "" {
		panic("crypto: signer name cannot be empty")
	}
	signerMu.Lock()
	signers[name] = s
	signerMu.Unlock()
}

// GetSigner returns the Signer registered under name, or nil if not found.
func GetSigner(name string) Signer {
	if name == "" {
		return nil
	}
	name = strings.ToLower(name)
	signerMu.RLock()
	s := signers[name]
	signerMu.RUnlock()
	return s
}

// RegisterVerifier registers a Verifier under its Name().
func RegisterVerifier(v Verifier) {
	if v == nil {
		panic("crypto: verifier cannot be nil")
	}
	name := strings.ToLower(v.Name())
	if name == "" {
		panic("crypto: verifier name cannot be empty")
	}
	verifierMu.Lock()
	verifiers[name] = v
	verifierMu.Unlock()
}

// GetVerifier returns the Verifier registered under name, or nil if not found.
func GetVerifier(name string) Verifier {
	if name == "" {
		return nil
	}
	name = strings.ToLower(name)
	verifierMu.RLock()
	v := verifiers[name]
	verifierMu.RUnlock()
	return v
}
