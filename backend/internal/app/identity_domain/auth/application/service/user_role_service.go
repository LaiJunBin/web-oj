package service

import "judge-backend/internal/app/identity_domain/auth/domain/entity/user"

type UserRoleService interface {
	GetRolesBit(user *user.User) int
	SetRolesBit(user *user.User, rolesBit int)
}
