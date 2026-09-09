package hashservice

import (
	"errors"
	"testing"

	"judge-backend/internal/app/shared/infrastructure/logger"
	sharedmocks "judge-backend/internal/app/shared/mocks"

	"github.com/stretchr/testify/assert"
)

func TestHashService_Hash(t *testing.T) {
	tests := []struct {
		name          string
		text          string
		mockHash      string
		mockError     error
		expectedHash  string
		expectedError error
	}{
		{
			name:          "Successfully hash text",
			text:          "test123",
			mockHash:      "hashed_test123",
			mockError:     nil,
			expectedHash:  "hashed_test123",
			expectedError: nil,
		},
		{
			name:          "Hash service returns error",
			text:          "test123",
			mockHash:      "",
			mockError:     errors.New("hash error"),
			expectedHash:  "",
			expectedError: errors.New("hash error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHashService := sharedmocks.NewHashService(t)
			mockHashService.On("Hash", tt.text).Return(tt.mockHash, tt.mockError)

			logger := logger.NewMockLogger()
			service := NewHashService(mockHashService, logger)

			hash, err := service.Hash(tt.text)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Empty(t, hash)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedHash, hash)
			}

			mockHashService.AssertExpectations(t)
		})
	}
}

func TestHashService_Compare(t *testing.T) {
	tests := []struct {
		name           string
		text           string
		hashed         string
		mockResult     bool
		expectedResult bool
	}{
		{
			name:           "Text matches hash",
			text:           "test123",
			hashed:         "hashed_test123",
			mockResult:     true,
			expectedResult: true,
		},
		{
			name:           "Text does not match hash",
			text:           "test123",
			hashed:         "hashed_test123",
			mockResult:     false,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHashService := sharedmocks.NewHashService(t)
			mockHashService.On("Compare", tt.text, tt.hashed).Return(tt.mockResult)

			logger := logger.NewMockLogger()
			service := NewHashService(mockHashService, logger)

			result := service.Compare(tt.text, tt.hashed)

			assert.Equal(t, tt.expectedResult, result)

			mockHashService.AssertExpectations(t)
		})
	}
}
