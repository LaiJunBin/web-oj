package roleconverter

import (
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/identity_domain/user/application/service/user/roleconverter"
	"judge-backend/internal/app/identity_domain/user/domain/entity/user/valueobject/role"
)

type RoleConverter struct {
	logger sharedtypes.Logger
}

func NewRoleConverter(logger sharedtypes.Logger) roleconverter.RoleConverter {
	return &RoleConverter{
		logger: logger,
	}
}

func (s *RoleConverter) ToBit(roles role.Roles) int {
	rolesBit := 0
	for _, role := range roles {
		rolesBit |= 1 << (role.ID() - 1)
	}

	// Debugf, printf-style: this runs on every authenticated request (the
	// authorization path re-reads the user), so Info was two garbled
	// key-value-into-printf lines per request.
	s.logger.Debugf("[RoleConverter] converted %d roles to bit %d", len(roles), rolesBit)
	return rolesBit
}

func (s *RoleConverter) ToRoles(rolesBit int) role.Roles {
	roles := role.Roles{}
	for _, role := range role.GetRoles() {
		bit := 1 << (role.ID() - 1)
		if rolesBit&bit != 0 {
			roles = append(roles, role)
		}
	}

	s.logger.Debugf("[RoleConverter] converted bit %d to %d roles", rolesBit, len(roles))
	return roles
}
