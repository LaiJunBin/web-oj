package role

import (
	"errors"
)

var (
	ErrInvalidPermissionValue = errors.New("invalid permission value")
)

const (
	PermissionRead       = 1 << iota // 1
	PermissionContestant             // 2
	PermissionAdmin                  // 4
)

type Permission int

func newPermission(value int) (Permission, error) {
	p := Permission(value)
	if err := p.Validate(); err != nil {
		return 0, err
	}

	return p, nil
}

func (p Permission) Value() int {
	return int(p)
}

func (p Permission) Equals(other Permission) bool {
	return p == other
}

func (p Permission) Validate() error {
	if p < 0 {
		return ErrInvalidPermissionValue
	}

	return nil
}
