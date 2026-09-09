package service

import (
	"testing"
)

func TestRandString(t *testing.T) {
	randomService := NewRandomService()

	length := 10
	result := randomService.RandString(length)

	if len(result) != length {
		t.Errorf("Expected string length %d, but got %d", length, len(result))
	}

	for _, char := range result {
		if !isLetter(char) {
			t.Errorf("Expected only letters, but got '%c'", char)
		}
	}
}

func isLetter(char rune) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')
}

// TestRandString_UniqueAndCoversAlphabet is a coarse guard against a broken or
// deterministic generator (the reason this service was moved to crypto/rand):
// a large sample of tokens must all differ and, taken together, exercise every
// letter of the alphabet. A constant or tiny-period generator fails both.
func TestRandString_UniqueAndCoversAlphabet(t *testing.T) {
	randomService := NewRandomService()

	const samples = 1000
	const tokenLen = 64
	seen := make(map[string]struct{}, samples)
	charset := make(map[rune]struct{})

	for i := 0; i < samples; i++ {
		tok := randomService.RandString(tokenLen)
		if len(tok) != tokenLen {
			t.Fatalf("expected length %d, got %d", tokenLen, len(tok))
		}
		if _, dup := seen[tok]; dup {
			t.Fatalf("duplicate token produced within %d samples: %q", samples, tok)
		}
		seen[tok] = struct{}{}
		for _, c := range tok {
			charset[c] = struct{}{}
		}
	}

	if len(charset) != len(letterBytes) {
		t.Errorf("expected all %d letters to appear across %d samples, got %d distinct", len(letterBytes), samples, len(charset))
	}
}
