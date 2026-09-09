package role

import "errors"

var (
	RoleAdmin      = declareRole(1, "admin", PermissionAdmin|PermissionRead)
	RoleContestant = declareRole(2, "contestant", PermissionContestant)
	RoleAPIReader  = declareRole(3, "api_reader", PermissionRead)

	ErrRoleNotFound = errors.New("role not found")
)

type Role struct {
	id         RoleID
	name       RoleName
	permission Permission
}
type Roles []Role

var roles = Roles{}

func GetRoles() Roles {
	return roles
}

func (r Roles) HasRole(role Role) bool {
	for _, r := range r {
		if r.Equals(role) {
			return true
		}
	}
	return false
}

func (r Roles) HasPermission(permission Permission) bool {
	for _, role := range r {
		if role.HasPermission(permission) {
			return true
		}
	}
	return false
}

func newRole(id int, name string, permission int) (Role, error) {
	roleID, err := newRoleID(id)
	if err != nil {
		return Role{}, err
	}

	roleName, err := newRoleName(name)
	if err != nil {
		return Role{}, err
	}

	p, err := newPermission(permission)
	if err != nil {
		return Role{}, err
	}

	role := Role{
		id:         roleID,
		name:       roleName,
		permission: p,
	}

	return role, nil
}

func mustNewRole(id int, name string, permission int) Role {
	role, err := newRole(id, name, permission)
	if err != nil {
		panic(err)
	}
	return role
}

func declareRole(id int, name string, permission int) Role {
	role := mustNewRole(id, name, permission)
	roles = append(roles, role)
	return role
}

func (r Role) ID() int32 {
	return r.id.Value()
}

func (r Role) Name() string {
	return r.name.String()
}

func (r Role) Equals(role Role) bool {
	return r.id.Equals(role.id) && r.name.Equals(role.name) && r.permission.Equals(role.permission)
}

func GetRoleByName(name string) (Role, error) {
	for _, role := range roles {
		if role.Name() == name {
			return role, nil
		}
	}

	return Role{}, ErrRoleNotFound
}

func GetRoleByID(id int32) (Role, error) {
	for _, role := range roles {
		if role.ID() == id {
			return role, nil
		}
	}

	return Role{}, ErrRoleNotFound
}

func GetRolesByID(ids []int32) ([]Role, error) {
	roles := make([]Role, 0, len(ids))
	for _, id := range ids {
		role, err := GetRoleByID(id)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func GetRolesByNames(names []string) ([]Role, error) {
	roles := make([]Role, 0, len(names))
	for _, name := range names {
		role, err := GetRoleByName(name)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func (r Role) Validate() error {
	if err := r.name.Validate(); err != nil {
		return err
	}

	if err := r.permission.Validate(); err != nil {
		return err
	}

	return nil
}

func (r Role) HasPermission(permission Permission) bool {
	return r.permission.Value()&permission.Value() != 0
}
