package repository

import (
	"errors"
	"judge-backend/internal/app/identity_domain/auth/application/service"
	"judge-backend/internal/app/identity_domain/auth/domain/entity/user"
	"judge-backend/internal/app/identity_domain/auth/domain/repository"
	"judge-backend/internal/app/shared/model"
	sharedtypes "judge-backend/internal/app/shared/types"

	"gorm.io/gorm"
)

type UserRepository struct {
	db      *gorm.DB
	service service.UserRoleService
	logger  sharedtypes.Logger
}

func NewUserRepository(
	db *gorm.DB,
	service service.UserRoleService,
	logger sharedtypes.Logger,
) repository.UserRepository {
	return &UserRepository{
		db:      db,
		service: service,
		logger:  logger,
	}
}

func (r *UserRepository) toDomainUser(u *model.User) (*user.User, error) {
	r.logger.Debugf("Converting model.User to domain.User: ID=%d, Username=%s", u.ID, u.Username)
	domainUser, err := user.New(u.Name, u.Username, u.Password)
	if err != nil {
		r.logger.Errorf("Failed to create domain.User: %v", err)
		return nil, err
	}

	err = domainUser.SetID(u.ID)
	if err != nil {
		r.logger.Errorf("Failed to set ID for domain.User: %v", err)
		return nil, err
	}

	r.service.SetRolesBit(domainUser, u.Roles)
	return domainUser, nil
}

func (r *UserRepository) FindByID(id int) (*user.User, error) {
	r.logger.Debugf("Finding user by ID: %d", id)
	u := &model.User{}
	if err := r.db.First(u, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.logger.Warnf("User not found for ID: %d", id)
			return nil, nil
		}
		r.logger.Errorf("Database error while finding user by ID: %v", err)
		return nil, err
	}

	r.logger.Debugf("User found: ID=%d, Username=%s", u.ID, u.Username)
	return r.toDomainUser(u)
}

func (r *UserRepository) FindByUsername(username string) (*user.User, error) {
	r.logger.Debugf("Finding user by username: %s", username)
	u := &model.User{}
	if err := r.db.Where("username = ?", username).First(u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.logger.Warnf("User not found for username: %s", username)
			return nil, nil
		}
		r.logger.Errorf("Database error while finding user by username: %v", err)
		return nil, err
	}

	r.logger.Debugf("User found: ID=%d, Username=%s", u.ID, u.Username)
	return r.toDomainUser(u)
}
