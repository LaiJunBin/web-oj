package roleconverter

import (
	"judge-backend/internal/app/identity_domain/user/domain/entity/user/valueobject/role"
)

type RoleConverter interface {
	ToBit(roles role.Roles) int
	ToRoles(rolesBit int) role.Roles
}
