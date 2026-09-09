package jwttoken

import (
	"errors"
	"judge-backend/internal/app/identity_domain/auth/domain/entity/user/valueobject/role"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenExpired     = errors.New("JWT token has expired")
	ErrUnexpectedMethod = errors.New("unexpected signing method")
	ErrInvalidToken     = errors.New("invalid token")
)

type Second = int64
type JwtTokenConfig struct {
	Secret          string
	EncryptedUserID string
	Roles           []*role.Role
	ExpirationTime  Second
}

type JwtToken struct {
	token  *jwt.Token
	secret string
}

func NewJwtToken(config JwtTokenConfig) (*JwtToken, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}
	// exp must live in the claims (not the header): golang-jwt — like every
	// standard JWT library — only ever validates the registered "exp" claim.
	// A header-only exp is never checked by the library itself, so any code
	// that parses a token without also calling this package's IsExpired()
	// would silently accept an expired token.
	claims["exp"] = time.Now().Add(time.Duration(config.ExpirationTime) * time.Second).Unix()
	claims["user_id"] = config.EncryptedUserID
	roleNames := make([]string, len(config.Roles))
	for i, r := range config.Roles {
		roleNames[i] = r.Name()
	}
	claims["roles"] = roleNames

	tokenString, err := token.SignedString([]byte(config.Secret))
	if err != nil {
		return nil, err
	}
	// Skip claims validation on this self-verification round-trip: we just
	// signed this token ourselves (an already-expired ExpirationTime is only
	// ever used in tests), so there is nothing to defend against here. The
	// untrusted-input path, NewJwtTokenWithRaw, deliberately does NOT skip
	// validation — that is what makes the library reject expired tokens
	// on its own now, independent of IsExpired().
	token, err = jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Secret), nil
	}, jwt.WithoutClaimsValidation())
	if err != nil {
		return nil, err
	}

	return &JwtToken{
		token:  token,
		secret: config.Secret,
	}, nil
}

func MustNewJwtToken(config JwtTokenConfig) *JwtToken {
	token, err := NewJwtToken(config)
	if err != nil {
		panic(err)
	}
	return token
}

func NewJwtTokenWithRaw(secret, raw string) (*JwtToken, error) {
	// Pin the accepted algorithm to HS256 at the parse itself. Validate() also
	// asserts an HMAC method, but pinning here means any caller that parses a
	// token without going through Validate() still cannot be fed an alg it did
	// not intend (e.g. "none", or an asymmetric alg used in a key-confusion
	// attack) — defence in depth for a security-critical primitive.
	token, err := jwt.Parse(raw, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return nil, err
	}

	return &JwtToken{
		token:  token,
		secret: secret,
	}, nil
}

func (j *JwtToken) IsExpired() bool {
	exp, ok := j.Claims()["exp"].(float64)
	if !ok {
		return true
	}

	expiredTime := time.Unix(int64(exp), 0)
	return time.Now().After(expiredTime)
}

func (j *JwtToken) Claims() map[string]interface{} {
	claims, ok := j.token.Claims.(jwt.MapClaims)
	if !ok {
		return map[string]interface{}{}
	}
	return claims
}

func (j *JwtToken) Validate() error {
	if j.IsExpired() {
		return ErrTokenExpired
	}

	if _, ok := j.token.Method.(*jwt.SigningMethodHMAC); !ok {
		return ErrUnexpectedMethod
	}

	if !j.token.Valid {
		return ErrInvalidToken
	}

	_, err := jwt.Parse(j.token.Raw, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.secret), nil
	})

	if err != nil {
		return ErrInvalidToken
	}

	return nil
}

func (j *JwtToken) String() string {
	return j.token.Raw
}

func (j *JwtToken) UserID() string {
	userID, ok := j.Claims()["user_id"].(string)
	if !ok {
		return ""
	}
	return userID
}

func (j *JwtToken) MaxAge() int64 {
	exp, _ := j.Claims()["exp"].(float64)
	return int64(exp) - time.Now().Unix()
}
