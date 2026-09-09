package user

import (
	"judge-backend/internal/app/identity_domain/auth/domain/entity/user/valueobject"
	"judge-backend/internal/app/identity_domain/auth/domain/entity/user/valueobject/role"
	sharedid "judge-backend/internal/app/identity_domain/shared/valueobject/userid"
)

type User struct {
	id       sharedid.UserID
	name     valueobject.Name
	username valueobject.Username
	password valueobject.Password
	roles    []*role.Role
}

func New(
	name string,
	username string,
	hashedPassword string,
) (*User, error) {
	nameVO, err := valueobject.NewName(name)
	if err != nil {
		return nil, err
	}

	usernameVO, err := valueobject.NewUsername(username)
	if err != nil {
		return nil, err
	}

	passwordVO, err := valueobject.NewPassword(hashedPassword)
	if err != nil {
		return nil, err
	}

	return &User{
		id:       sharedid.MustNew(0),
		name:     nameVO,
		username: usernameVO,
		password: passwordVO,
		roles:    []*role.Role{},
	}, nil
}

func MustNew(
	name string,
	username string,
	hashedPassword string,
) *User {
	user, err := New(name, username, hashedPassword)
	if err != nil {
		panic(err)
	}
	return user
}

func (u *User) IsNew() bool {
	return u.id.Equals(0)
}

func (u *User) GetID() int {
	return u.id.Value()
}

func (u *User) SetID(id int) error {
	uid, err := sharedid.New(id)
	if err != nil {
		return err
	}
	u.id = uid
	return nil
}

func (u *User) GetName() string {
	return u.name.String()
}

func (u *User) SetName(name string) error {
	nameVO, err := valueobject.NewName(name)
	if err != nil {
		return err
	}
	u.name = nameVO
	return nil
}

func (u *User) GetUsername() string {
	return u.username.String()
}

func (u *User) SetUsername(username string) error {
	usernameVO, err := valueobject.NewUsername(username)
	if err != nil {
		return err
	}
	u.username = usernameVO
	return nil
}

func (u *User) GetPassword() string {
	return u.password.String()
}

func (u *User) SetPassword(hashedPassword string) error {
	passwordVO, err := valueobject.NewPassword(hashedPassword)
	if err != nil {
		return err
	}
	u.password = passwordVO
	return nil
}

func (u *User) GetRoles() []*role.Role {
	return u.roles
}

func (u *User) GetRoleNames() []string {
	roleNames := make([]string, len(u.roles))
	for i, r := range u.roles {
		roleNames[i] = r.Name()
	}
	return roleNames
}

func (u *User) SetRoles(roles ...*role.Role) {
	u.roles = roles
}

func (u *User) HasRole(role *role.Role) bool {
	for _, r := range u.roles {
		if r.Equals(role) {
			return true
		}
	}
	return false
}

func (u *User) HasPermission(permission role.Permission) bool {
	for _, r := range u.roles {
		if r.HasPermission(permission) {
			return true
		}
	}
	return false
}
