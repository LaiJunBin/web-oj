package password

import "errors"

var (
	ErrEmptyPassword = errors.New("password cannot be empty")
)

type Password string

func New(password string) (Password, error) {
	if password == "" {
		return "", ErrEmptyPassword
	}

	return Password(password), nil
}

func (p Password) String() string {
	return string(p)
}
