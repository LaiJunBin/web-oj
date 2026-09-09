package valueobject

import "errors"

var (
	ErrInvalidUsernameLength = errors.New("username must be at least 3 characters long")
)

type Username string

func NewUsername(username string) (Username, error) {
	u := Username(username)
	if err := u.Validate(); err != nil {
		return "", err
	}

	return u, nil
}

func (u Username) String() string {
	return string(u)
}

func (u Username) Equals(other Username) bool {
	return u == other
}

func (u Username) Validate() error {
	if len(u) < 3 {
		return ErrInvalidUsernameLength
	}

	return nil
}
