package refreshtoken

import (
	"testing"
)

func TestNewRefreshToken(t *testing.T) {
	token := MustNewRefreshToken(RefreshTokenConfig{
		UserID:         1,
		ExpirationTime: 3600,
		Token:          "refresh_token",
	})

	tests := []struct {
		name      string
		got, want interface{}
	}{
		{"UserID", token.UserID().Value(), 1},
		{"IsExpired", token.IsExpired(), false},
		{"Token", token.String(), "refresh_token"},
		{"ExpirationTime", token.ExpirationTime().Value(), Second(3600)},
	}

	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s() = %v, want %v", tt.name, tt.got, tt.want)
		}
	}
}

func TestRefreshTokenExpired(t *testing.T) {
	token := MustNewRefreshToken(RefreshTokenConfig{
		UserID:         1,
		ExpirationTime: 0,
		Token:          "refresh_token",
	})

	if !token.IsExpired() {
		t.Errorf("IsExpired() = %v, want %v", token.IsExpired(), true)
	}
}
