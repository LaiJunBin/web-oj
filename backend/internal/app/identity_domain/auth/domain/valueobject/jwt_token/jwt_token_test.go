package jwttoken

import (
	"judge-backend/internal/app/identity_domain/auth/domain/entity/user/valueobject/role"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const secret = "secret"

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

func mustNewJwtToken(t *testing.T, config JwtTokenConfig) *JwtToken {
	t.Helper()
	token, err := NewJwtToken(config)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	return token
}

func TestJwtTokenCreation(t *testing.T) {
	token := mustNewJwtToken(t, JwtTokenConfig{
		Secret:          secret,
		EncryptedUserID: "1",
		Roles:           []*role.Role{},
		ExpirationTime:  1,
	})

	expected := expectedClaims{
		UserID: "1",
		Roles:  []string{},
	}
	assertClaims(t, token.Claims(), expected)
}

func TestJwtTokenCreationWithRaw(t *testing.T) {
	token := mustNewJwtToken(t, JwtTokenConfig{
		Secret:          secret,
		EncryptedUserID: "2",
		Roles:           []*role.Role{},
		ExpirationTime:  1,
	})
	raw := token.String()

	jwtToken, err := NewJwtTokenWithRaw(secret, raw)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	expected := expectedClaims{
		UserID: "2",
		Roles:  []string{},
	}
	assertClaims(t, jwtToken.Claims(), expected)

	if jwtToken.String() != raw {
		t.Fatalf("expected %v, got %v", raw, jwtToken.token.Raw)
	}
}

func TestJwtTokenExpired(t *testing.T) {
	token := mustNewJwtToken(t, JwtTokenConfig{
		Secret:          secret,
		EncryptedUserID: "1",
		Roles:           []*role.Role{},
		ExpirationTime:  -1,
	})

	if !token.IsExpired() {
		t.Fatalf("expected expired, got not expired")
	}
}

// TestJwtTokenWithRaw_RejectsExpiredTokenAtParseTime is a regression test for
// A6: exp used to live only in the JWT header, a place golang-jwt (like any
// standard JWT library) never validates — only THIS package's own IsExpired()
// enforced expiry, so any future code path that parsed a token and read its
// claims without also remembering to call IsExpired()/Validate() would have
// silently accepted a token that never expires. Now that exp lives in the
// claims, NewJwtTokenWithRaw (the path that parses untrusted, caller-supplied
// tokens) rejects an expired token by itself, at parse time — independent of
// whether the caller remembers to check expiry afterward.
func TestJwtTokenWithRaw_RejectsExpiredTokenAtParseTime(t *testing.T) {
	expired := mustNewJwtToken(t, JwtTokenConfig{
		Secret:          secret,
		EncryptedUserID: "1",
		Roles:           []*role.Role{},
		ExpirationTime:  -1,
	})

	_, err := NewJwtTokenWithRaw(secret, expired.String())
	if err == nil {
		t.Fatalf("expected NewJwtTokenWithRaw to reject an expired token by itself, got nil error")
	}
}

func TestJwtTokenValidate(t *testing.T) {
	t.Run("valid token", func(t *testing.T) {
		token := mustNewJwtToken(t, JwtTokenConfig{
			Secret:          secret,
			EncryptedUserID: "1",
			Roles:           []*role.Role{},
			ExpirationTime:  1,
		})

		if err := token.Validate(); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		t.Run("no method", func(t *testing.T) {
			token := mustNewJwtToken(t, JwtTokenConfig{
				Secret:          secret,
				EncryptedUserID: "1",
				Roles:           []*role.Role{},
				ExpirationTime:  1,
			})
			token.token.Method = nil
			if err := token.Validate(); err == nil {
				t.Fatalf("expected error, got nil")
			}
		})

		t.Run("invalid method", func(t *testing.T) {
			token := mustNewJwtToken(t, JwtTokenConfig{
				Secret:          secret,
				EncryptedUserID: "1",
				Roles:           []*role.Role{},
				ExpirationTime:  1,
			})
			token.token.Method = jwt.GetSigningMethod("NONE")
			if err := token.Validate(); err == nil {
				t.Fatalf("expected error, got %v", err)
			}
		})

		t.Run("set invalid", func(t *testing.T) {
			token := mustNewJwtToken(t, JwtTokenConfig{
				Secret:          secret,
				EncryptedUserID: "1",
				Roles:           []*role.Role{},
				ExpirationTime:  1,
			})
			token.token.Valid = false
			if err := token.Validate(); err == nil {
				t.Fatalf("expected error, got nil")
			}
		})

		t.Run("set expired to invalid", func(t *testing.T) {
			token := mustNewJwtToken(t, JwtTokenConfig{
				Secret:          secret,
				EncryptedUserID: "1",
				Roles:           []*role.Role{},
				ExpirationTime:  1,
			})
			token.token.Claims.(jwt.MapClaims)["exp"] = "invalid"
			if err := token.Validate(); err == nil {
				t.Fatalf("expected error, got nil")
			}
		})

		t.Run("delete exp", func(t *testing.T) {
			token := mustNewJwtToken(t, JwtTokenConfig{
				Secret:          secret,
				EncryptedUserID: "1",
				Roles:           []*role.Role{},
				ExpirationTime:  1,
			})
			delete(token.token.Claims.(jwt.MapClaims), "exp")
			if err := token.Validate(); err == nil {
				t.Fatalf("expected error, got nil")
			}
		})

		t.Run("set expired", func(t *testing.T) {
			token := mustNewJwtToken(t, JwtTokenConfig{
				Secret:          secret,
				EncryptedUserID: "1",
				Roles:           []*role.Role{},
				ExpirationTime:  1,
			})
			token.token.Claims.(jwt.MapClaims)["exp"] = float64(time.Now().Add(-1).Unix())
			if err := token.Validate(); err == nil {
				t.Fatalf("expected error, got nil")
			}
		})

		t.Run("set token to invalid", func(t *testing.T) {
			token := mustNewJwtToken(t, JwtTokenConfig{
				Secret:          secret,
				EncryptedUserID: "1",
				Roles:           []*role.Role{},
				ExpirationTime:  1,
			})
			token.token.Raw = "invalid"
			if err := token.Validate(); err == nil {
				t.Fatalf("expected error, got %v", err)
			}
		})

		t.Run("set token to another token with different secret", func(t *testing.T) {
			token := mustNewJwtToken(t, JwtTokenConfig{
				Secret:          secret,
				EncryptedUserID: "1",
				Roles:           []*role.Role{},
				ExpirationTime:  1,
			})
			fakeToken2 := mustNewJwtToken(t, JwtTokenConfig{
				Secret:          "secret2",
				EncryptedUserID: "1",
				Roles:           []*role.Role{},
				ExpirationTime:  1,
			})
			token.token.Raw = fakeToken2.token.Raw
			if err := token.Validate(); err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	})
}
