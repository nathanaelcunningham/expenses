package security

import (
	"strings"
	"testing"
)

func TestGenerateSessionToken(t *testing.T) {
	token, err := GenerateSessionToken()
	if err != nil {
		t.Fatalf("Failed to generate session token: %v", err)
	}

	if token == "" {
		t.Fatal("Generated token is empty")
	}

	// Token should be base64url encoded (no padding, URL-safe characters)
	if strings.Contains(token, "=") {
		t.Error("Token contains padding characters")
	}

	if strings.Contains(token, "+") || strings.Contains(token, "/") {
		t.Error("Token contains non-URL-safe characters")
	}

	// Token should be long enough (32 bytes = 43 chars in base64url)
	if len(token) < 40 {
		t.Errorf("Token too short: %d characters", len(token))
	}

	t.Logf("Generated token: %s (length: %d)", token, len(token))
}

func TestGenerateSecureToken(t *testing.T) {
	tests := []struct {
		name       string
		byteLength int
		wantError  bool
	}{
		{"Valid 16 bytes", 16, false},
		{"Valid 32 bytes", 32, false},
		{"Valid 64 bytes", 64, false},
		{"Invalid 0 bytes", 0, true},
		{"Invalid negative", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateSecureToken(tt.byteLength)
			
			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if token == "" {
				t.Fatal("Generated token is empty")
			}

			// Verify token format
			if err := ValidateTokenFormat(token); err != nil {
				t.Errorf("Generated token has invalid format: %v", err)
			}

			t.Logf("Generated %d-byte token: %s", tt.byteLength, token)
		})
	}
}

func TestValidateTokenFormat(t *testing.T) {
	tests := []struct {
		name      string
		token     string
		wantError bool
	}{
		{"Valid token", "VGhpcyBpcyBhIHZhbGlkIHRva2Vu", false},
		{"Valid with URL-safe chars", "VGhpcyBpcyBhIHZhbGlkIHRva2Vu-_", false},
		{"Empty token", "", true},
		{"Invalid base64", "invalid!@#$", true},
		{"Base64 with padding", "VGVzdA==", true}, // We don't allow padding
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTokenFormat(tt.token)
			
			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestTokenUniqueness(t *testing.T) {
	tokens := make(map[string]bool)
	numTokens := 1000

	for i := 0; i < numTokens; i++ {
		token, err := GenerateSessionToken()
		if err != nil {
			t.Fatalf("Failed to generate token %d: %v", i, err)
		}

		if tokens[token] {
			t.Fatalf("Duplicate token generated: %s", token)
		}

		tokens[token] = true
	}

	t.Logf("Successfully generated %d unique tokens", numTokens)
}

func TestGetTokenByteLength(t *testing.T) {
	tg := NewTokenGenerator()
	
	token, err := tg.GenerateSecureToken(32)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	byteLength, err := tg.GetTokenByteLength(token)
	if err != nil {
		t.Fatalf("Failed to get token byte length: %v", err)
	}

	if byteLength != 32 {
		t.Errorf("Expected byte length 32, got %d", byteLength)
	}
}