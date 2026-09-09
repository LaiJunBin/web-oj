package user

import (
	"judge-backend/internal/app/identity_domain/user/domain/entity/user/valueobject/role"
	userid "judge-backend/internal/app/identity_domain/shared/valueobject/userid"
	"testing"
)

func createUser(t *testing.T) *User {
	user, err := New(&UserConfig{
		Name:           "user",
		Username:       "testuser",
		HashedPassword: "password",
	})

	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	return user
}

func assertRoles(t *testing.T, user *User, expectedRoles ...role.Role) {
	if len(user.roles) != len(expectedRoles) {
		t.Errorf("expected roles to have length %d, got %d", len(expectedRoles), len(user.roles))
	}
	for _, r := range expectedRoles {
		if !user.HasRole(r) {
			t.Errorf("expected user to have role %v", r)
		}
	}
}

func TestSetRoles(t *testing.T) {
	user := createUser(t)

	assertRoles(t, user)
	user.SetRoles(role.RoleAdmin)
	assertRoles(t, user, role.RoleAdmin)

	user.SetRoles(role.RoleAdmin, role.RoleContestant)
	assertRoles(t, user, role.RoleAdmin, role.RoleContestant)

	if user.HasRole(role.RoleAPIReader) {
		t.Errorf("expected user to not have role %v", role.RoleAPIReader)
	}

	user.SetRoles()
	assertRoles(t, user)
}

func TestIsNew(t *testing.T) {
	u := createUser(t)
	if !u.IsNew() {
		t.Errorf("expected user to be new")
	}

	u.id = userid.MustNew(1)
	if u.IsNew() {
		t.Errorf("expected user to not be new")
	}
}
