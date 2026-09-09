package service_test

import (
	"judge-backend/internal/app/identity_domain/auth/domain/entity/user/valueobject/role"
	sysRole "judge-backend/internal/app/identity_domain/user/domain/entity/user/valueobject/role"
	"testing"
)

// The auth domain and the user domain each carry their own copy of the role
// model: auth's is consulted on every request (JWT claims, the admin/
// contestant/read middleware), user's is what gets persisted and managed. They
// are kept in sync BY HAND, and the bridge between them (UserRoleService in
// this package) matches roles by name — so a divergence would not fail loudly,
// it would quietly authorize the wrong things.
//
// This test pins BOTH models to a hard-coded truth table — the table is the
// third leg, not just a mirror: comparing the two models only to each other
// would stay green if both drifted in lockstep (e.g. the same 4th role added
// on both sides). If it fails, someone changed a role or permission; extend
// the table deliberately (and think about existing persisted role bits
// before renumbering anything).
func TestRoleModelsAreInParity(t *testing.T) {
	// The truth table is HARD-CODED rather than copied from either side.
	expected := []struct {
		name       string
		id         int32
		read       bool
		contestant bool
		admin      bool
	}{
		{name: "admin", id: 1, read: true, contestant: false, admin: true},
		{name: "contestant", id: 2, read: false, contestant: true, admin: false},
		{name: "api_reader", id: 3, read: true, contestant: false, admin: false},
	}

	t.Run("both sides declare exactly the pinned roles, no more", func(t *testing.T) {
		authCount := len(role.Roles)
		userCount := len(sysRole.GetRoles())
		if authCount != userCount {
			t.Fatalf("auth declares %d roles, user declares %d — a role was added on one side only", authCount, userCount)
		}
		// Without this, adding a 4th role on BOTH sides passes silently —
		// counts match, the three named subtests still pass — while the new
		// role already satisfies the authorization middleware.
		if authCount != len(expected) {
			t.Fatalf("role models declare %d roles but the pinned truth table has %d — a role was added without extending the table", authCount, len(expected))
		}
	})

	authByName := func(name string) *role.Role {
		for _, r := range role.Roles {
			if r.Name() == name {
				return r
			}
		}
		return nil
	}

	for _, want := range expected {
		t.Run(want.name, func(t *testing.T) {
			authRole := authByName(want.name)
			if authRole == nil {
				t.Fatalf("auth side has no role named %q", want.name)
			}
			userRole, err := sysRole.GetRoleByName(want.name)
			if err != nil {
				t.Fatalf("user side has no role named %q: %v", want.name, err)
			}

			if got := int32(authRole.ID()); got != want.id {
				t.Errorf("auth %s ID = %d, want %d", want.name, got, want.id)
			}
			if got := userRole.ID(); got != want.id {
				t.Errorf("user %s ID = %d, want %d", want.name, got, want.id)
			}

			checks := []struct {
				perm     string
				authPerm role.Permission
				userPerm sysRole.Permission
				want     bool
			}{
				{"read", role.PermissionRead, sysRole.PermissionRead, want.read},
				{"contestant", role.PermissionContestant, sysRole.PermissionContestant, want.contestant},
				{"admin", role.PermissionAdmin, sysRole.PermissionAdmin, want.admin},
			}
			for _, c := range checks {
				if got := authRole.HasPermission(c.authPerm); got != c.want {
					t.Errorf("auth %s.HasPermission(%s) = %v, want %v", want.name, c.perm, got, c.want)
				}
				if got := userRole.HasPermission(c.userPerm); got != c.want {
					t.Errorf("user %s.HasPermission(%s) = %v, want %v", want.name, c.perm, got, c.want)
				}
			}
		})
	}
}
