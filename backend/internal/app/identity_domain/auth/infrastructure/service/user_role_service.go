package service

import (
	"judge-backend/internal/app/identity_domain/auth/application/service"
	"judge-backend/internal/app/identity_domain/auth/domain/entity/user"
	"judge-backend/internal/app/identity_domain/auth/domain/entity/user/valueobject/role"
	"judge-backend/internal/app/identity_domain/user/application/service/user/roleconverter"
	sysRole "judge-backend/internal/app/identity_domain/user/domain/entity/user/valueobject/role"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type UserRoleService struct {
	roleConverter roleconverter.RoleConverter
	logger        sharedtypes.Logger
}

func NewUserRoleService(
	roleConverter roleconverter.RoleConverter,
	logger sharedtypes.Logger,
) service.UserRoleService {
	return &UserRoleService{
		roleConverter: roleConverter,
		logger:        logger,
	}
}

func (s *UserRoleService) GetRolesBit(user *user.User) int {
	s.logger.Debugf("Getting roles bit for user: %v", user)
	roles := sysRole.Roles{}
	for _, r := range user.GetRoles() {
		role, err := sysRole.GetRoleByName(r.Name())
		if err != nil {
			s.logger.Warnf("Failed to get role by name: %s, error: %v", r.Name(), err)
			continue
		}
		roles = append(roles, role)
	}

	return s.roleConverter.ToBit(roles)
}

func (s *UserRoleService) SetRolesBit(user *user.User, rolesBit int) {
	s.logger.Debugf("Setting roles bit for user: %v with roles bit: %d", user, rolesBit)
	roles := []*role.Role{}
	sysRoles := s.roleConverter.ToRoles(rolesBit)
	for _, r := range sysRoles {
		for _, role := range role.Roles {
			if r.Name() == role.Name() {
				s.logger.Debugf("Matching role found: %s", r.Name())
				roles = append(roles, role)
				break
			}
		}
	}
	user.SetRoles(roles...)
}
