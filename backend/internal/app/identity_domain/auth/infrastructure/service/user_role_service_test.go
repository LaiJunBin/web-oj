package service_test

import (
	"judge-backend/internal/app/identity_domain/auth/domain/entity/user"
	"judge-backend/internal/app/identity_domain/auth/domain/entity/user/valueobject/role"
	"judge-backend/internal/app/identity_domain/auth/infrastructure/service"
	sysRole "judge-backend/internal/app/identity_domain/user/domain/entity/user/valueobject/role"
	"judge-backend/internal/app/identity_domain/user/mocks"
	"judge-backend/internal/app/shared/infrastructure/logger"
	"testing"
)

func TestUserRoleService_GetRolesBit(t *testing.T) {
	tests := []struct {
		roles    []*role.Role
		expected int
	}{
		{
			roles:    []*role.Role{},
			expected: 0,
		},
		{
			roles: []*role.Role{
				role.RoleAdmin,
			},
			expected: 1,
		},
		{
			roles: []*role.Role{
				role.RoleContestant,
			},
			expected: 2,
		},
		{
			roles: []*role.Role{
				role.RoleAdmin,
				role.RoleContestant,
			},
			expected: 3,
		},
		{
			roles: []*role.Role{
				role.RoleAdmin,
				role.RoleContestant,
				role.RoleAPIReader,
			},
			expected: 7,
		},
		{
			roles: []*role.Role{
				role.RoleAdmin,
				role.RoleContestant,
				role.RoleAPIReader,
			},
			expected: 7,
		},
	}

	user := user.MustNew("test", "test", "test")
	roleConverter := mocks.NewRoleConverter(t)
	logger := logger.NewMockLogger()
	service := service.NewUserRoleService(roleConverter, logger)
	for _, test := range tests {
		user.SetRoles(test.roles...)
		sysRoles := sysRole.Roles{}
		for _, r := range test.roles {
			sysRole, err := sysRole.GetRoleByName(r.Name())
			if err != nil {
				t.Fatalf("failed to get sysRole by name: %v", err)
			}
			sysRoles = append(sysRoles, sysRole)
		}
		roleConverter.On("ToBit", sysRoles).Return(test.expected).Once()
		rolesBit := service.GetRolesBit(user)

		if rolesBit != test.expected {
			t.Fatalf("expected roles bit to be %d, got %d", test.expected, rolesBit)
		}
	}
}

func TestUserRoleService_SetRolesBit(t *testing.T) {
	tests := []struct {
		roles    sysRole.Roles
		rolesBit int
	}{
		{
			roles:    sysRole.Roles{},
			rolesBit: 0,
		},
		{
			roles: sysRole.Roles{
				sysRole.RoleAdmin,
			},
			rolesBit: 1,
		},
		{
			roles: sysRole.Roles{
				sysRole.RoleContestant,
			},
			rolesBit: 2,
		},
		{
			roles: sysRole.Roles{
				sysRole.RoleAdmin,
				sysRole.RoleContestant,
			},
			rolesBit: 3,
		},
		{
			roles: sysRole.Roles{
				sysRole.RoleAdmin,
				sysRole.RoleContestant,
				sysRole.RoleAPIReader,
			},
			rolesBit: 7,
		},
		{
			roles: sysRole.Roles{
				sysRole.RoleAdmin,
				sysRole.RoleContestant,
				sysRole.RoleAPIReader,
			},
			rolesBit: 7,
		},
	}

	user := user.MustNew("test", "test", "test")
	roleConverter := mocks.NewRoleConverter(t)
	logger := logger.NewMockLogger()
	service := service.NewUserRoleService(roleConverter, logger)
	for _, test := range tests {
		roleConverter.On("ToRoles", test.rolesBit).Return(test.roles).Once()
		service.SetRolesBit(user, test.rolesBit)
		roles := user.GetRoles()
		if len(roles) != len(test.roles) {
			t.Fatalf("expected roles to have length %d, got %d", len(test.roles), len(roles))
		}

		for i, role := range roles {
			if role.Name() != test.roles[i].Name() {
				t.Fatalf("expected role at index %d to be %s, got %s", i, test.roles[i].Name(), role.Name())
			}
		}
	}
}
