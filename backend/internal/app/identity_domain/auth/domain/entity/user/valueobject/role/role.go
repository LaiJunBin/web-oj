package role

var (
	RoleAdmin      = declareRole(1, "admin", PermissionAdmin|PermissionRead)
	RoleContestant = declareRole(2, "contestant", PermissionContestant)
	RoleAPIReader  = declareRole(3, "api_reader", PermissionRead)
)

var Roles = []*Role{}

type Role struct {
	id         RoleID
	name       RoleName
	permission Permission
}

func newRole(id int, name string, permission int) (*Role, error) {
	roleID, err := newRoleID(id)
	if err != nil {
		return nil, err
	}

	roleName, err := newRoleName(name)
	if err != nil {
		return nil, err
	}

	p, err := newPermission(permission)
	if err != nil {
		return nil, err
	}

	role := Role{
		id:         roleID,
		name:       roleName,
		permission: p,
	}

	return &role, nil
}

func mustNewRole(id int, name string, permission int) *Role {
	role, err := newRole(id, name, permission)
	if err != nil {
		panic(err)
	}
	return role
}

func declareRole(id int, name string, permission int) *Role {
	role := mustNewRole(id, name, permission)
	Roles = append(Roles, role)
	return role
}

func (r *Role) ID() int {
	return r.id.Value()
}

func (r *Role) Name() string {
	return r.name.String()
}

func (r *Role) Equals(role *Role) bool {
	return r.id.Equals(role.id) && r.name.Equals(role.name) && r.permission.Equals(role.permission)
}

func GetRoleByName(name string) *Role {
	for _, role := range Roles {
		if role.Name() == name {
			return role
		}
	}

	return nil
}

func (r *Role) Validate() error {
	if err := r.name.Validate(); err != nil {
		return err
	}

	if err := r.permission.Validate(); err != nil {
		return err
	}

	return nil
}

func (r *Role) HasPermission(permission Permission) bool {
	return r.permission.Value()&permission.Value() != 0
}
