package refreshtoken

import (
	"errors"
	sharedid "judge-backend/internal/app/identity_domain/shared/valueobject/userid"
	"time"
)

var (
	ErrEmptyToken   = errors.New("empty token")
	ErrTimeDuration = errors.New("time duration must be greater than 0")
)

type Second = int64
type ExpirationTime struct {
	time Second
}

func (e ExpirationTime) Value() int64 {
	return e.time
}

func (e ExpirationTime) Duration() time.Duration {
	return time.Duration(e.time * int64(time.Second))
}

func (e ExpirationTime) IsExpired() bool {
	return e.time <= 0
}

func (e ExpirationTime) String() string {
	return time.Now().Add(e.Duration()).String()
}

func (e ExpirationTime) Equal(expirationTime ExpirationTime) bool {
	return e.time == expirationTime.time
}

type RefreshToken struct {
	userID         sharedid.UserID
	token          string
	expirationTime ExpirationTime
}

type RefreshTokenConfig struct {
	Token          string
	UserID         int
	ExpirationTime Second
}

func NewRefreshToken(config RefreshTokenConfig) (*RefreshToken, error) {
	userID, err := sharedid.New(config.UserID)
	if err != nil {
		return nil, err
	}

	if config.Token == "" {
		return nil, ErrEmptyToken
	}

	return &RefreshToken{
		userID:         userID,
		token:          config.Token,
		expirationTime: ExpirationTime{time: config.ExpirationTime},
	}, nil
}

func MustNewRefreshToken(config RefreshTokenConfig) *RefreshToken {
	token, err := NewRefreshToken(config)
	if err != nil {
		panic(err)
	}
	return token
}

func (r *RefreshToken) UserID() sharedid.UserID {
	return r.userID
}

func (r *RefreshToken) String() string {
	return r.token
}

func (r *RefreshToken) IsExpired() bool {
	return r.expirationTime.IsExpired()
}

func (r *RefreshToken) ExpirationTime() ExpirationTime {
	return r.expirationTime
}
