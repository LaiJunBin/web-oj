package role

import "errors"

type RoleID int

var (
	ErrInvalidRoleID = errors.New("invalid role id")
)

func newRoleID(id int) (RoleID, error) {
	rid := RoleID(id)
	if err := rid.Validate(); err != nil {
		return 0, err
	}

	return rid, nil
}

func (id RoleID) Value() int {
	return int(id)
}

func (id RoleID) Equals(other RoleID) bool {
	return id == other
}

func (id RoleID) Validate() error {
	if id <= 0 {
		return ErrInvalidRoleID
	}

	return nil
}
