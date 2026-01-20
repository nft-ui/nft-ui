package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"math/big"
	"os"
)

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// TokenGenerator handles token generation for quota queries
type TokenGenerator struct {
	salt     string
	hostname string
}

// NewTokenGenerator creates a new token generator
func NewTokenGenerator(salt string) *TokenGenerator {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}
	return &TokenGenerator{salt: salt, hostname: hostname}
}

// Generate creates an 8-character token for a given port
func (t *TokenGenerator) Generate(port int) string {
	// Create input string: "port:hostname"
	input := fmt.Sprintf("%d:%s", port, t.hostname)

	// HMAC-SHA256 with salt as key
	h := hmac.New(sha256.New, []byte(t.salt))
	h.Write([]byte(input))
	hash := h.Sum(nil)

	// Base62 encode and take first 8 characters
	encoded := t.base62Encode(hash)
	if len(encoded) > 8 {
		return encoded[:8]
	}
	return encoded
}

// FindQuotaByToken finds the quota that matches the given token
// Returns nil if not found
func (t *TokenGenerator) FindQuotaByToken(token string, quotas []QuotaRule) *QuotaRule {
	for i := range quotas {
		if t.Generate(quotas[i].Port) == token {
			return &quotas[i]
		}
	}
	return nil
}

// base62Encode encodes bytes to base62 string
func (t *TokenGenerator) base62Encode(data []byte) string {
	num := new(big.Int).SetBytes(data)
	base := big.NewInt(62)
	zero := big.NewInt(0)
	var result []byte

	for num.Cmp(zero) > 0 {
		mod := new(big.Int)
		num.DivMod(num, base, mod)
		result = append([]byte{base62Chars[mod.Int64()]}, result...)
	}

	// Pad to ensure minimum length of 8 characters
	for len(result) < 8 {
		result = append([]byte{'0'}, result...)
	}

	return string(result)
}

// IsValidTokenFormat checks if a token has valid format (8 alphanumeric chars)
func IsValidTokenFormat(token string) bool {
	if len(token) != 8 {
		return false
	}
	for _, c := range token {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return false
		}
	}
	return true
}
