package valueobject

import "errors"

var (
	ErrInvalidNameLength = errors.New("name must be at least 2 characters long")
)

type Name string

func NewName(name string) (Name, error) {
	u := Name(name)
	if err := u.Validate(); err != nil {
		return "", err
	}

	return u, nil
}

func (u Name) String() string {
	return string(u)
}

func (u Name) Equals(other Name) bool {
	return u == other
}

func (u Name) Validate() error {
	if len(u) < 2 {
		return ErrInvalidNameLength
	}

	return nil
}
