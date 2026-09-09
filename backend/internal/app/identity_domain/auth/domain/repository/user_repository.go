package repository

import (
	"judge-backend/internal/app/identity_domain/auth/domain/entity/user"
)

type UserRepository interface {
	FindByID(id int) (*user.User, error)
	FindByUsername(username string) (*user.User, error)
}
