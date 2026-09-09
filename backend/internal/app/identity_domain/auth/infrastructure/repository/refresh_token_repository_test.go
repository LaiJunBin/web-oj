package repository_test

import (
	refreshtoken "judge-backend/internal/app/identity_domain/auth/domain/entity/refresh_token"
	"judge-backend/internal/app/identity_domain/auth/infrastructure/repository"
	"judge-backend/internal/app/shared/infrastructure/logger"
	"strconv"
	"testing"

	"github.com/go-redis/redismock/v9"
)

func newRefreshTokenRepository() (*repository.RefreshTokenRepository, redismock.ClientMock) {
	db, mock := redismock.NewClientMock()
	logger := logger.NewMockLogger()
	r := repository.NewRefreshTokenRepository(db, logger)
	repo, ok := r.(*repository.RefreshTokenRepository)
	if !ok {
		panic("error creating test repository")
	}

	return repo, mock
}

type expectedRefreshToken struct {
	Token          string
	UserID         int
	ExpirationTime refreshtoken.Second
}

func assertRefreshToken(t *testing.T, token *refreshtoken.RefreshToken, expected expectedRefreshToken) {
	if token == nil {
		t.Fatalf("expected token, got nil")
	}

	if token.String() != expected.Token {
		t.Fatalf("expected token to be '%s', got %s", expected.Token, token.String())
	}

	if token.UserID().Value() != expected.UserID {
		t.Fatalf("expected user id to be %d, got %d", expected.UserID, token.UserID().Value())
	}

	if token.ExpirationTime().Value() != expected.ExpirationTime {
		t.Fatalf("expected expiration time to be %d, got %d", expected.ExpirationTime, token.ExpirationTime().Value())
	}
}

func TestRefreshTokenRepository(t *testing.T) {
	t.Run("TestFindByToken", func(t *testing.T) {
		repo, mock := newRefreshTokenRepository()

		t.Run("Success", func(t *testing.T) {
			refreshToken := refreshtoken.MustNewRefreshToken(refreshtoken.RefreshTokenConfig{
				Token:          "token",
				UserID:         1,
				ExpirationTime: 3600,
			})
			mock.ExpectGet(refreshToken.String()).SetVal(strconv.Itoa(refreshToken.UserID().Value()))
			mock.ExpectTTL(refreshToken.String()).SetVal(refreshToken.ExpirationTime().Duration())

			token, err := repo.FindByToken("token")
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}

			expected := expectedRefreshToken{
				Token:          "token",
				UserID:         1,
				ExpirationTime: 3600,
			}
			assertRefreshToken(t, token, expected)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("expected all mock expectations to be met, got %v", err)
			}
		})

		t.Run("Expired", func(t *testing.T) {
			mock.ExpectGet("token").SetVal("1")
			mock.ExpectTTL("token").SetVal(-1)

			token, err := repo.FindByToken("token")
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}

			if token != nil {
				t.Fatalf("expected nil, got %v", token)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("expected all mock expectations to be met, got %v", err)
			}
		})

		t.Run("NotFound", func(t *testing.T) {
			mock.ExpectGet("token").RedisNil()

			token, err := repo.FindByToken("token")
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}

			if token != nil {
				t.Fatalf("expected nil, got %v", token)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("expected all mock expectations to be met, got %v", err)
			}
		})
	})

	t.Run("TestSave", func(t *testing.T) {
		repo, mock := newRefreshTokenRepository()
		refreshToken := refreshtoken.MustNewRefreshToken(refreshtoken.RefreshTokenConfig{
			Token:          "token",
			UserID:         1,
			ExpirationTime: 3600,
		})
		mock.ExpectSet(refreshToken.String(), refreshToken.UserID().Value(), refreshToken.ExpirationTime().Duration()).SetVal("OK")
		mock.ExpectSAdd("user_refresh_tokens:1", refreshToken.String()).SetVal(1)
		mock.ExpectExpire("user_refresh_tokens:1", refreshToken.ExpirationTime().Duration()).SetVal(true)

		err := repo.Save(refreshToken)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("expected all mock expectations to be met, got %v", err)
		}
	})

	t.Run("TestDelete", func(t *testing.T) {
		repo, mock := newRefreshTokenRepository()
		refreshToken := refreshtoken.MustNewRefreshToken(refreshtoken.RefreshTokenConfig{
			Token:          "token",
			UserID:         1,
			ExpirationTime: 3600,
		})

		mock.ExpectDel(refreshToken.String()).SetVal(1)
		mock.ExpectSRem("user_refresh_tokens:1", refreshToken.String()).SetVal(1)

		err := repo.Delete(refreshToken)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("expected all mock expectations to be met, got %v", err)
		}
	})

	t.Run("TestReplace", func(t *testing.T) {
		repo, mock := newRefreshTokenRepository()
		oldToken := refreshtoken.MustNewRefreshToken(refreshtoken.RefreshTokenConfig{
			Token:          "token",
			UserID:         1,
			ExpirationTime: 3600,
		})
		newToken := refreshtoken.MustNewRefreshToken(refreshtoken.RefreshTokenConfig{
			Token:          "new_token",
			UserID:         1,
			ExpirationTime: 3500,
		})

		mock.ExpectRename(oldToken.String(), newToken.String()).SetVal("OK")
		mock.ExpectSet(newToken.String(), newToken.UserID().Value(), newToken.ExpirationTime().Duration()).SetVal("OK")
		mock.ExpectSRem("user_refresh_tokens:1", oldToken.String()).SetVal(1)
		mock.ExpectSAdd("user_refresh_tokens:1", newToken.String()).SetVal(1)
		mock.ExpectExpire("user_refresh_tokens:1", newToken.ExpirationTime().Duration()).SetVal(true)

		err := repo.Replace(oldToken, newToken)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("expected all mock expectations to be met, got %v", err)
		}
	})


	// Revocation reads the index rather than the keyspace: this Redis is
	// shared with the CSRF session store, so there is no pattern that safely
	// matches "refresh tokens and nothing else".
	t.Run("TestDeleteAllByUserID", func(t *testing.T) {
		repo, mock := newRefreshTokenRepository()

		mock.ExpectSMembers("user_refresh_tokens:7").SetVal([]string{"token_a", "token_b"})
		mock.ExpectDel("token_a", "token_b").SetVal(2)
		mock.ExpectDel("user_refresh_tokens:7").SetVal(1)

		if err := repo.DeleteAllByUserID(7); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("expected all mock expectations to be met, got %v", err)
		}
	})

	// An account that never logged in has no index; revoking it is a no-op
	// rather than an error, so a password reset does not fail on it.
	t.Run("TestDeleteAllByUserIDWithNoSessions", func(t *testing.T) {
		repo, mock := newRefreshTokenRepository()

		mock.ExpectSMembers("user_refresh_tokens:8").SetVal([]string{})
		mock.ExpectDel("user_refresh_tokens:8").SetVal(0)

		if err := repo.DeleteAllByUserID(8); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("expected all mock expectations to be met, got %v", err)
		}
	})
}
