package repository

import (
	"context"
	"fmt"

	refreshtoken "judge-backend/internal/app/identity_domain/auth/domain/entity/refresh_token"
	"judge-backend/internal/app/identity_domain/auth/domain/repository"
	sharedtypes "judge-backend/internal/app/shared/types"

	"github.com/redis/go-redis/v9"
)

type RefreshTokenRepository struct {
	rdb    *redis.Client
	logger sharedtypes.Logger
}

func NewRefreshTokenRepository(
	rdb *redis.Client,
	logger sharedtypes.Logger,
) repository.RefreshTokenRepository {
	return &RefreshTokenRepository{
		rdb:    rdb,
		logger: logger,
	}
}

func (r *RefreshTokenRepository) FindByToken(token string) (*refreshtoken.RefreshToken, error) {
	r.logger.Debugf("Finding refresh token")
	ctx := context.Background()
	t := r.rdb.Get(ctx, token)
	if t.Err() != nil {
		if t.Err() == redis.Nil {
			r.logger.Warnf("Refresh token not found")
			return nil, nil
		}
		r.logger.Errorf("Error retrieving refresh token: %v", t.Err())
		return nil, t.Err()
	}

	userID, err := t.Int()
	if err != nil {
		r.logger.Errorf("Error converting token to user ID: %v", err)
		return nil, err
	}

	ttl := r.rdb.TTL(ctx, token).Val().Seconds()
	if ttl < 0 {
		r.logger.Warnf("Token expired or invalid TTL for refresh token")
		return nil, nil
	}

	refreshToken, err := refreshtoken.NewRefreshToken(refreshtoken.RefreshTokenConfig{
		Token:          token,
		UserID:         userID,
		ExpirationTime: int64(ttl),
	})
	if err != nil {
		r.logger.Errorf("Error creating refresh token entity: %v", err)
		return nil, err
	}

	return refreshToken, nil
}

// userTokenIndexKey names the set of refresh tokens currently issued to one
// user. Tokens themselves are stored under their own opaque value, which
// answers "who is this token for?" but not "what is this user holding?" —
// and revoking an account's sessions needs the second question. Scanning the
// keyspace is not an option: this Redis is shared with the CSRF session
// store, so there is no safe pattern to match on.
//
// The set is an index, not the source of truth. A stale member costs one
// no-op DEL; a missing member is the case the index exists to prevent, so
// every write path below maintains it.
func userTokenIndexKey(userID int) string {
	return fmt.Sprintf("user_refresh_tokens:%d", userID)
}

func (r *RefreshTokenRepository) Save(token *refreshtoken.RefreshToken) error {
	r.logger.Debugf("Saving refresh token for user %d", token.UserID().Value())
	ctx := context.Background()
	err := r.rdb.Set(ctx, token.String(), token.UserID().Value(), token.ExpirationTime().Duration()).Err()
	if err != nil {
		r.logger.Errorf("Error saving refresh token: %v", err)
		return err
	}

	return r.indexToken(ctx, token)
}

// indexToken adds a token to its user's set and pushes the set's own expiry
// out to match. Without the expiry the index would be the one key that
// outlives every token it points at, and would accumulate for every account
// that ever logged in.
func (r *RefreshTokenRepository) indexToken(ctx context.Context, token *refreshtoken.RefreshToken) error {
	key := userTokenIndexKey(token.UserID().Value())
	if err := r.rdb.SAdd(ctx, key, token.String()).Err(); err != nil {
		r.logger.Errorf("Error indexing refresh token for user %d: %v", token.UserID().Value(), err)
		return err
	}

	if err := r.rdb.Expire(ctx, key, token.ExpirationTime().Duration()).Err(); err != nil {
		r.logger.Errorf("Error setting expiry on refresh token index for user %d: %v", token.UserID().Value(), err)
		return err
	}

	return nil
}

func (r *RefreshTokenRepository) Delete(token *refreshtoken.RefreshToken) error {
	r.logger.Debugf("Deleting refresh token for user %d", token.UserID().Value())
	ctx := context.Background()
	err := r.rdb.Del(ctx, token.String()).Err()
	if err != nil {
		r.logger.Errorf("Error deleting refresh token: %v", err)
		return err
	}

	if err := r.rdb.SRem(ctx, userTokenIndexKey(token.UserID().Value()), token.String()).Err(); err != nil {
		r.logger.Errorf("Error de-indexing refresh token for user %d: %v", token.UserID().Value(), err)
		return err
	}

	return nil
}

func (r *RefreshTokenRepository) Replace(oldToken, newToken *refreshtoken.RefreshToken) error {
	r.logger.Debugf("Replacing refresh token for user %d", oldToken.UserID().Value())
	ctx := context.Background()
	err := r.rdb.Rename(ctx, oldToken.String(), newToken.String()).Err()
	if err != nil {
		r.logger.Errorf("Error renaming token: %v", err)
		return err
	}

	err = r.rdb.Set(ctx, newToken.String(), newToken.UserID().Value(), newToken.ExpirationTime().Duration()).Err()
	if err != nil {
		r.logger.Errorf("Error saving new token after replacement: %v", err)
		return err
	}

	// Rotation retires one token and issues another, so the index has to
	// follow both halves — otherwise the rotated-away token lingers as a
	// stale member and, worse, the live one is missing when the account is
	// revoked.
	if err := r.rdb.SRem(ctx, userTokenIndexKey(oldToken.UserID().Value()), oldToken.String()).Err(); err != nil {
		r.logger.Errorf("Error de-indexing replaced refresh token for user %d: %v", oldToken.UserID().Value(), err)
		return err
	}

	return r.indexToken(ctx, newToken)
}

// DeleteAllByUserID revokes every session the user currently holds. It is the
// point of the index: an administrator resetting a compromised account's
// password needs the old credentials to stop working everywhere, not just on
// the device that happened to make the last request.
//
// Note this cannot reach the access token, which is a self-contained JWT and
// stays valid until it expires (auth.accessTokenTTLSeconds, 3600s by
// default). What it does guarantee is that the session cannot be renewed past
// that point.
func (r *RefreshTokenRepository) DeleteAllByUserID(userID int) error {
	r.logger.Debugf("Deleting all refresh tokens for user %d", userID)
	ctx := context.Background()
	key := userTokenIndexKey(userID)

	tokens, err := r.rdb.SMembers(ctx, key).Result()
	if err != nil {
		r.logger.Errorf("Error reading refresh token index for user %d: %v", userID, err)
		return err
	}

	if len(tokens) > 0 {
		if err := r.rdb.Del(ctx, tokens...).Err(); err != nil {
			r.logger.Errorf("Error deleting refresh tokens for user %d: %v", userID, err)
			return err
		}
	}

	if err := r.rdb.Del(ctx, key).Err(); err != nil {
		r.logger.Errorf("Error deleting refresh token index for user %d: %v", userID, err)
		return err
	}

	r.logger.Infof("Revoked %d refresh token(s) for user %d", len(tokens), userID)
	return nil
}
