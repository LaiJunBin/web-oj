package roleconverter_test

import (
	"judge-backend/internal/app/shared/infrastructure/logger"
	"judge-backend/internal/app/identity_domain/user/domain/entity/user/valueobject/role"
	"judge-backend/internal/app/identity_domain/user/infrastructure/service/user/roleconverter"
	"testing"
)

func TestRoleConverter_ToBit(t *testing.T) {
	tests := []struct {
		roles    role.Roles
		expected int
	}{
		{
			roles:    role.Roles{},
			expected: 0,
		},
		{
			roles: role.Roles{
				role.RoleAdmin,
			},
			expected: 1,
		},
		{
			roles: role.Roles{
				role.RoleContestant,
			},
			expected: 2,
		},
		{
			roles: role.Roles{
				role.RoleAdmin,
				role.RoleContestant,
			},
			expected: 3,
		},
		{
			roles: role.Roles{
				role.RoleAdmin,
				role.RoleContestant,
				role.RoleAPIReader,
			},
			expected: 7,
		},
		{
			roles: role.Roles{
				role.RoleAdmin,
				role.RoleContestant,
				role.RoleAPIReader,
			},
			expected: 7,
		},
	}

	logger := logger.NewMockLogger()
	service := roleconverter.NewRoleConverter(logger)
	for _, test := range tests {
		rolesBit := service.ToBit(test.roles)

		if rolesBit != test.expected {
			t.Fatalf("expected roles bit to be %d, got %d", test.expected, rolesBit)
		}
	}
}

func TestRoleConverter_SetRolesBit(t *testing.T) {
	tests := []struct {
		roles    role.Roles
		rolesBit int
	}{
		{
			roles:    role.Roles{},
			rolesBit: 0,
		},
		{
			roles: role.Roles{
				role.RoleAdmin,
			},
			rolesBit: 1,
		},
		{
			roles: role.Roles{
				role.RoleContestant,
			},
			rolesBit: 2,
		},
		{
			roles: role.Roles{
				role.RoleAdmin,
				role.RoleContestant,
			},
			rolesBit: 3,
		},
		{
			roles: role.Roles{
				role.RoleAdmin,
				role.RoleContestant,
				role.RoleAPIReader,
			},
			rolesBit: 7,
		},
		{
			roles: role.Roles{
				role.RoleAdmin,
				role.RoleContestant,
				role.RoleAPIReader,
			},
			rolesBit: 7,
		},
	}

	logger := logger.NewMockLogger()
	service := roleconverter.NewRoleConverter(logger)
	for _, test := range tests {
		roles := service.ToRoles(test.rolesBit)

		if len(roles) != len(test.roles) {
			t.Fatalf("expected roles to have length %d, got %d", len(test.roles), len(roles))
		}

		for i, role := range roles {
			if role != test.roles[i] {
				t.Fatalf("expected role at index %d to be %v, got %v", i, test.roles[i], role)
			}
		}
	}
}
