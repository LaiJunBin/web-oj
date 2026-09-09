package role

import "errors"

var (
	ErrRoleNameRequired = errors.New("role name is required")
)

type RoleName struct {
	value string
}

func newRoleName(roleName string) (RoleName, error) {
	rn := RoleName{
		value: roleName,
	}

	if err := rn.Validate(); err != nil {
		return RoleName{}, err
	}

	return rn, nil
}

func (r RoleName) String() string {
	return r.value
}

func (r RoleName) Equals(other RoleName) bool {
	return r.value == other.value
}

func (r RoleName) Validate() error {
	if r.value == "" {
		return ErrRoleNameRequired
	}
	return nil
}
