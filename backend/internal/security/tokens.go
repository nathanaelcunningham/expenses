package security

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// TokenGenerator provides secure token generation utilities
type TokenGenerator struct{}

// NewTokenGenerator creates a new token generator
func NewTokenGenerator() *TokenGenerator {
	return &TokenGenerator{}
}

// GenerateSessionToken generates a cryptographically secure session token
// Returns a base64url-encoded string with 128 bits of entropy
func (tg *TokenGenerator) GenerateSessionToken() (string, error) {
	// Generate 32 bytes (256 bits) of random data for high entropy
	// This provides significantly more security than needed but ensures future-proofing
	bytes := make([]byte, 32)
	
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	
	// Use base64url encoding to make tokens URL-safe and remove padding
	token := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(bytes)
	
	return token, nil
}

// GenerateSecureToken generates a secure token of specified byte length
// Returns a base64url-encoded string
func (tg *TokenGenerator) GenerateSecureToken(byteLength int) (string, error) {
	if byteLength <= 0 {
		return "", fmt.Errorf("byte length must be positive, got %d", byteLength)
	}
	
	bytes := make([]byte, byteLength)
	
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	
	// Use base64url encoding to make tokens URL-safe and remove padding
	token := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(bytes)
	
	return token, nil
}

// ValidateTokenFormat validates that a token is properly formatted
// This doesn't validate the token's authenticity, only its format
func (tg *TokenGenerator) ValidateTokenFormat(token string) error {
	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}
	
	// Check if the token is valid base64url
	_, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(token)
	if err != nil {
		return fmt.Errorf("invalid token format: %w", err)
	}
	
	return nil
}

// GetTokenByteLength returns the original byte length of a base64url-encoded token
func (tg *TokenGenerator) GetTokenByteLength(token string) (int, error) {
	decoded, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(token)
	if err != nil {
		return 0, fmt.Errorf("invalid token format: %w", err)
	}
	
	return len(decoded), nil
}

// Package-level convenience functions for common operations

// GenerateSessionToken generates a cryptographically secure session token
func GenerateSessionToken() (string, error) {
	tg := NewTokenGenerator()
	return tg.GenerateSessionToken()
}

// GenerateSecureToken generates a secure token of specified byte length
func GenerateSecureToken(byteLength int) (string, error) {
	tg := NewTokenGenerator()
	return tg.GenerateSecureToken(byteLength)
}

// ValidateTokenFormat validates that a token is properly formatted
func ValidateTokenFormat(token string) error {
	tg := NewTokenGenerator()
	return tg.ValidateTokenFormat(token)
}