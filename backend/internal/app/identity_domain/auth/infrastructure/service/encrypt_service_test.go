package service

import (
	"judge-backend/internal/app/identity_domain/auth/application/service"
	logger "judge-backend/internal/app/shared/infrastructure/logger"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/config"
	"testing"
)

func encryptAndAssert(t *testing.T, s service.EncryptService, value string) string {
	encrypted, err := s.Encrypt(value)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if encrypted == "" {
		t.Fatalf("expected encrypted value, got empty string")
	}

	if encrypted == value {
		t.Fatalf("expected encrypted value, got unencrypted value")
	}

	return encrypted
}

func decryptAndAssert(t *testing.T, s service.EncryptService, value string) string {
	decrypted, err := s.Decrypt(value)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if decrypted == value {
		t.Fatalf("expected decrypted value, got encrypted value")
	}

	return decrypted
}

func TestEncrypt(t *testing.T) {
	mockLogger := logger.NewMockLogger()
	var logger sharedtypes.Logger = mockLogger
	config := &config.Config{
		Server: &config.ServerConfig{
			AppKey: "secretsecretsecretsecret",
		},
	}
	s := NewEncryptService(config, logger)

	t.Run("Test Encrypt short", func(t *testing.T) {
		encryptAndAssert(t, s, "a")
	})

	t.Run("Test Encrypt long", func(t *testing.T) {
		encryptAndAssert(t, s, "passwordpasswordpasswordpassword")
	})

	t.Run("Test Encrypt empty", func(t *testing.T) {
		encryptAndAssert(t, s, "")
	})
}

func TestDecrypt(t *testing.T) {
	mockLogger := logger.NewMockLogger()
	var logger sharedtypes.Logger = mockLogger
	config := &config.Config{
		Server: &config.ServerConfig{
			AppKey: "secretsecretsecretsecret",
		},
	}
	s := NewEncryptService(config, logger)

	t.Run("Test Decrypt short", func(t *testing.T) {
		encrypted := encryptAndAssert(t, s, "a")
		decrypted := decryptAndAssert(t, s, encrypted)
		if decrypted != "a" {
			t.Fatalf("expected a, got %v", decrypted)
		}
	})

	t.Run("Test Decrypt long", func(t *testing.T) {
		encrypted := encryptAndAssert(t, s, "passwordpasswordpasswordpassword")
		decrypted := decryptAndAssert(t, s, encrypted)
		if decrypted != "passwordpasswordpasswordpassword" {
			t.Fatalf("expected passwordpasswordpasswordpassword, got %v", decrypted)
		}
	})

	t.Run("Test Decrypt empty", func(t *testing.T) {
		encrypted := encryptAndAssert(t, s, "")
		decrypted := decryptAndAssert(t, s, encrypted)
		if decrypted != "" {
			t.Fatalf("expected empty string, got %v", decrypted)
		}
	})
}

func TestDecryptError(t *testing.T) {
	mockLogger := logger.NewMockLogger()
	var logger sharedtypes.Logger = mockLogger
	config := &config.Config{
		Server: &config.ServerConfig{
			AppKey: "secretsecretsecretsecret",
		},
	}
	s := NewEncryptService(config, logger)
	_, err := s.Decrypt("password")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	_, err = s.Decrypt("adfkasofjkawofwkqofjqagfoajfgowajfgoaf")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
