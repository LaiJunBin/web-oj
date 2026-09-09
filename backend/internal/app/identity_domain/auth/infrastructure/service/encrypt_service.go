package service

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"judge-backend/internal/app/identity_domain/auth/application/service"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/config"
)

type EncryptService struct {
	key    string
	logger sharedtypes.Logger
}

func NewEncryptService(config *config.Config, logger sharedtypes.Logger) service.EncryptService {
	return &EncryptService{
		key:    config.Server.AppKey,
		logger: logger,
	}
}

func pKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func pKCS7UnPadding(origData []byte) ([]byte, error) {
	length := len(origData)
	unpadding := int(origData[length-1])
	if unpadding > length {
		return nil, errors.New("unpadding error")
	}
	return origData[:(length - unpadding)], nil
}

func (s *EncryptService) Encrypt(value string) (string, error) {
	s.logger.Debugf("Encrypting value: %s", value)
	block, err := aes.NewCipher([]byte(s.key[:16]))
	if err != nil {
		s.logger.Errorf("Failed to create cipher for encryption, error: %v", err)
		return "", err
	}
	plaintext := pKCS7Padding([]byte(value), block.BlockSize())
	ciphertext := make([]byte, block.BlockSize()+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		s.logger.Errorf("Failed to generate IV for encryption, error: %v", err)
		return "", err
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)
	s.logger.Debugf("Value encrypted successfully: %s", value)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (s *EncryptService) Decrypt(value string) (string, error) {
	s.logger.Debugf("Decrypting value: %s", value)
	ciphertext, _ := base64.StdEncoding.DecodeString(value)
	block, err := aes.NewCipher([]byte(s.key[:16]))
	if err != nil {
		s.logger.Errorf("Failed to create cipher for decryption, error: %v", err)
		return "", err
	}
	if len(ciphertext) < aes.BlockSize {
		s.logger.Errorf("Failed to decrypt value: %s, error: ciphertext too short", value)
		return "", errors.New("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)
	unpadding, err := pKCS7UnPadding(ciphertext)
	if err != nil {
		s.logger.Errorf("Failed to decrypt value: %s, error: %v", value, err)
		return "", err
	}
	s.logger.Debugf("Value decrypted successfully: %s", value)
	return string(unpadding), nil
}
