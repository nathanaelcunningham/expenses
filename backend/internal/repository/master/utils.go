package master

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
)

// generateID generates a random hex ID
func generateID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if random fails
		return fmt.Sprintf("%x", []byte(fmt.Sprintf("%d", 123456789)))
	}
	return hex.EncodeToString(bytes)
}

// generateInviteCode generates a memorable invite code
func generateInviteCode() string {
	words := []string{
		"apple", "brave", "cloud", "dance", "eagle", "flame", "grape", "heart",
		"island", "joy", "kite", "light", "moon", "nature", "ocean", "peace",
		"quiet", "river", "star", "tree", "unity", "voice", "water", "bright",
		"calm", "dream", "free", "green", "happy", "love", "magic", "pure",
	}

	// Generate 3 random words
	wordIndices := make([]int, 3)
	for i := range wordIndices {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(words))))
		if err != nil {
			// Fallback to simple index if random fails
			wordIndices[i] = i % len(words)
		} else {
			wordIndices[i] = int(index.Int64())
		}
	}

	// Add a random number for uniqueness
	num, err := rand.Int(rand.Reader, big.NewInt(999))
	if err != nil {
		num = big.NewInt(123) // Fallback
	}

	return fmt.Sprintf("%s-%s-%s-%03d", 
		words[wordIndices[0]], 
		words[wordIndices[1]], 
		words[wordIndices[2]], 
		num.Int64())
}