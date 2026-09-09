package hashservice_test

import (
	"judge-backend/internal/app/shared/infrastructure/logger"
	hashservice "judge-backend/internal/app/shared/infrastructure/services/hash_service"
	"testing"
)

func TestHashService_Hash(t *testing.T) {
	logger := logger.NewMockLogger()
	hs := hashservice.NewHashService(logger)

	text := "password123"
	hashed, err := hs.Hash(text)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(hashed) == 0 {
		t.Fatalf("expected hashed password to have length greater than 0")
	}

	if text == hashed {
		t.Fatalf("expected hashed password to be different from original")
	}
}

func TestHashService_Compare(t *testing.T) {
	logger := logger.NewMockLogger()
	hs := hashservice.NewHashService(logger)

	text := "password123"
	hashed, err := hs.Hash(text)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !hs.Compare(text, hashed) {
		t.Fatalf("expected hash comparison to succeed")
	}

	if hs.Compare("wrongpassword", hashed) {
		t.Fatalf("expected hash comparison to fail")
	}
}
