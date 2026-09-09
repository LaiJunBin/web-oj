package user

import (
	"judge-backend/internal/app/identity_domain/user/domain/entity/user/valueobject/name"
	"judge-backend/internal/app/identity_domain/user/domain/entity/user/valueobject/password"
	"judge-backend/internal/app/identity_domain/user/domain/entity/user/valueobject/role"
	userid "judge-backend/internal/app/identity_domain/shared/valueobject/userid"
	"judge-backend/internal/app/identity_domain/user/domain/entity/user/valueobject/username"
)

type User struct {
	id       userid.UserID
	name     name.Name
	username username.Username
	password password.Password
	roles    []role.Role
	groupIDs []int
}

type UserConfig struct {
	Name           string
	Username       string
	HashedPassword string
}

func New(config *UserConfig) (*User, error) {
	nameVO, err := name.New(config.Name)
	if err != nil {
		return nil, err
	}

	usernameVO, err := username.New(config.Username)
	if err != nil {
		return nil, err
	}

	passwordVO, err := password.New(config.HashedPassword)
	if err != nil {
		return nil, err
	}

	return &User{
		id:       userid.MustNew(0),
		name:     nameVO,
		username: usernameVO,
		password: passwordVO,
		roles:    []role.Role{},
		groupIDs: []int{},
	}, nil
}

func MustNew(
	name string,
	username string,
	hashedPassword string,
) *User {
	user, err := New(&UserConfig{
		Name:           name,
		Username:       username,
		HashedPassword: hashedPassword,
	})
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
	uid, err := userid.New(id)
	if err != nil {
		return err
	}
	u.id = uid
	return nil
}

func (u *User) GetName() string {
	return u.name.String()
}

func (u *User) SetName(newName string) error {
	nameVO, err := name.New(newName)
	if err != nil {
		return err
	}
	u.name = nameVO
	return nil
}

func (u *User) GetUsername() string {
	return u.username.String()
}

func (u *User) SetUsername(newUsername string) error {
	usernameVO, err := username.New(newUsername)
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
	passwordVO, err := password.New(hashedPassword)
	if err != nil {
		return err
	}
	u.password = passwordVO
	return nil
}

func (u *User) GetRoles() role.Roles {
	return u.roles
}

func (u *User) GetRoleNames() []string {
	roleNames := make([]string, len(u.roles))
	for i, r := range u.roles {
		roleNames[i] = r.Name()
	}
	return roleNames
}

func (u *User) SetRoles(roles ...role.Role) {
	u.roles = roles
}

func (u *User) HasRole(role role.Role) bool {
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


func (u *User) GetGroupIDs() []int {
	return u.groupIDs
}

func (u *User) SetGroupIDs(groupIDs []int) {
	u.groupIDs = groupIDs
}
