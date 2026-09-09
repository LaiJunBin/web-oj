package service

import (
	"errors"
	"judge-backend/internal/app/identity_domain/auth/domain/entity/user"
	"judge-backend/internal/app/identity_domain/auth/domain/entity/user/valueobject/role"
	jwttoken "judge-backend/internal/app/identity_domain/auth/domain/valueobject/jwt_token"
	"judge-backend/internal/app/identity_domain/auth/mocks"
	"judge-backend/internal/app/shared/infrastructure/logger"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/config"
	"testing"

	"github.com/stretchr/testify/mock"
)

type TokenServiceMocks struct {
	config         *config.Config
	encryptService *mocks.EncryptService
	randomService  *mocks.RandomService
	logger         sharedtypes.Logger
}

func newTokenServiceMocks() *TokenServiceMocks {
	return &TokenServiceMocks{
		config: &config.Config{
			Server: &config.ServerConfig{
				JwtKey: "secret",
			},
		},
		encryptService: &mocks.EncryptService{},
		randomService:  &mocks.RandomService{},
		logger:         logger.NewMockLogger(),
	}
}

func newTestTokenService() (*TokenService, *TokenServiceMocks) {
	mocks := newTokenServiceMocks()
	service, ok := NewTokenService(mocks.config, mocks.encryptService, mocks.randomService, mocks.logger).(*TokenService)
	if !ok {
		panic("failed to cast TokenService")
	}

	return service, mocks
}

type expectedClaims struct {
	UserID string
	Roles  []string
}

func assertClaims(t *testing.T, claims map[string]interface{}, expected expectedClaims) {
	if claims["user_id"] != expected.UserID {
		t.Fatalf("expected %v, got %v", expected.UserID, claims["user_id"])
	}

	roles, ok := claims["roles"].([]interface{})
	if !ok {
		t.Fatalf("expected []interface{}, got %T", claims["roles"])
	}

	if len(roles) != len(expected.Roles) {
		t.Fatalf("expected %v, got %v", len(expected.Roles), len(roles))
	}

	for i, role := range expected.Roles {
		if roles[i] != role {
			t.Fatalf("expected %v, got %v", role, roles[i])
		}
	}
}

func TestGenerateToken(t *testing.T) {
	tests := []struct {
		name           string
		userID         int
		roles          []*role.Role
		encryptReturn  string
		encryptError   error
		expectedClaims expectedClaims
		expectError    bool
	}{
		{
			name:          "should create a jwt token",
			userID:        1,
			encryptReturn: "encrypted_user01",
			expectedClaims: expectedClaims{
				UserID: "encrypted_user01",
				Roles:  []string{},
			},
		},
		{
			name:          "should return error when encrypting user id fails",
			userID:        1,
			encryptReturn: "",
			encryptError:  errors.New("encrypt error"),
			expectError:   true,
		},
		{
			name:          "generate with empty role",
			userID:        1,
			encryptReturn: "encrypted_user01",
			expectedClaims: expectedClaims{
				UserID: "encrypted_user01",
				Roles:  []string{},
			},
		},
		{
			name:          "generate with single role",
			userID:        2,
			roles:         []*role.Role{role.RoleAdmin},
			encryptReturn: "encrypted_user02",
			expectedClaims: expectedClaims{
				UserID: "encrypted_user02",
				Roles:  []string{"admin"},
			},
		},
		{
			name:          "generate with multiple roles",
			userID:        3,
			roles:         []*role.Role{role.RoleAdmin, role.RoleContestant, role.RoleAPIReader},
			encryptReturn: "encrypted_user03",
			expectedClaims: expectedClaims{
				UserID: "encrypted_user03",
				Roles:  []string{"admin", "contestant", "api_reader"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mocks := newTestTokenService()
			mocks.encryptService.On("Encrypt", mock.Anything).Return(tt.encryptReturn, tt.encryptError)

			user := user.MustNew("User", "user", "password123")
			user.SetID(tt.userID)
			user.SetRoles(tt.roles...)

			token, err := service.GenerateAccessToken(user, 1)
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if token != nil {
					t.Fatalf("expected nil, got %v", token)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if token == nil {
				t.Fatalf("expected token, got nil")
			}

			if token.String() == "" {
				t.Fatalf("expected token value, got empty string")
			}

			jwtToken, err := jwttoken.NewJwtTokenWithRaw(mocks.config.Server.JwtKey, token.String())
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if jwtToken == nil {
				t.Fatalf("expected jwt token, got nil")
			}

			claims := jwtToken.Claims()
			assertClaims(t, claims, tt.expectedClaims)
		})
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	service, mocks := newTestTokenService()
	mocks.randomService.On("RandString", mock.Anything).Return("refresh_token", nil)

	user := user.MustNew("User", "user", "password123")
	user.SetID(1)

	token, err := service.GenerateRefreshToken(user, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if token == nil {
		t.Fatalf("expected token, got nil")
	}

	if token.String() != "refresh_token" {
		t.Fatalf("expected refresh_token, got %v", token.String())
	}

	if token.UserID().Value() != 1 {
		t.Fatalf("expected 1, got %v", token.UserID().Value())
	}

	if token.ExpirationTime().Value() != 1 {
		t.Fatalf("expected 1, got %v", token.ExpirationTime().Value())
	}
}
