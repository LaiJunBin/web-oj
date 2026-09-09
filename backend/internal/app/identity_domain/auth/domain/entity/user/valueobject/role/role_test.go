package role

import (
	"testing"
)

func TestRoleCreation(t *testing.T) {
	role, err := newRole(1, "admin", PermissionRead|PermissionAdmin)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	if role.Name() != "admin" {
		t.Errorf("expected Name to be 'admin', got %s", role.Name())
	}

	if !role.permission.Equals(PermissionRead | PermissionAdmin) {
		t.Errorf("expected Permissions to be %d, got %d", PermissionRead|PermissionAdmin, role.permission)
	}
}

func TestRoleEquality(t *testing.T) {
	tests := []struct {
		role1, role2 *Role
		expected     bool
	}{
		{mustNewRole(1, "admin", PermissionRead), mustNewRole(1, "admin", PermissionRead), true},
		{mustNewRole(1, "admin", PermissionRead), mustNewRole(1, "contestant", PermissionRead), false},
	}

	for _, tt := range tests {
		if tt.role1.Equals(tt.role2) != tt.expected {
			t.Errorf("expected %v.equals(%v) to be %v", tt.role1, tt.role2, tt.expected)
		}
	}
}

func TestRolePermissionCheck(t *testing.T) {
	role := mustNewRole(1, "admin", PermissionRead)

	tests := []struct {
		permission Permission
		expected   bool
	}{
		{PermissionRead, true},
		{PermissionAdmin, false},
	}

	for _, tt := range tests {
		if role.HasPermission(tt.permission) != tt.expected {
			t.Errorf("expected role to have permission %v: %v", tt.permission, tt.expected)
		}
	}
}

func TestRoleSpecificPermissions(t *testing.T) {
	tests := []struct {
		role       *Role
		permission Permission
		expected   bool
	}{
		{RoleAPIReader, PermissionRead, true},
		{RoleAdmin, PermissionRead | PermissionAdmin, true},
		{RoleAPIReader, PermissionAdmin, false},
		{RoleContestant, PermissionAdmin, false},
	}

	for _, tt := range tests {
		if tt.role.HasPermission(tt.permission) != tt.expected {
			t.Errorf("expected %v to have permission %v: %v", tt.role, tt.permission, tt.expected)
		}
	}
}
