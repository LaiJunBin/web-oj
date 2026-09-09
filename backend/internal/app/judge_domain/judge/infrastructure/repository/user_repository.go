package repository

import (
	"judge-backend/internal/app/judge_domain/judge/domain/repository"
	"judge-backend/internal/app/shared/model"
	sharedtypes "judge-backend/internal/app/shared/types"

	"gorm.io/gorm"
)

type UserRepository struct {
	db     *gorm.DB
	logger sharedtypes.Logger
}

func NewUserRepository(
	db *gorm.DB,
	logger sharedtypes.Logger,
) repository.UserRepository {
	return &UserRepository{
		db:     db,
		logger: logger,
	}
}

func (r *UserRepository) FindUsernameByID(userID int32) (string, error) {
	user := &model.User{}
	// Unscoped: see the port's doc comment — soft-deleted accounts must still
	// resolve so their in-flight submissions can be judged.
	if err := r.db.Unscoped().Select("username", "deleted_at").Where("id = ?", userID).First(user).Error; err != nil {
		r.logger.Errorf("[UserRepository] Failed to find user %d: %v", userID, err)
		return "", err
	}
	// A soft-deleted row's username was renamed to free the unique index
	// (model.DeletedUsername); strip the suffix so SUBMIT_USERNAME keeps
	// addressing the per-seat resources created under the ORIGINAL name.
	// Only dead rows are stripped — a pre-guard live user named
	// "x__deleted_9" (creating one is rejected as reserved nowadays)
	// passes through untouched.
	if user.DeletedAt.Valid {
		return model.OriginalUsername(user.Username), nil
	}
	return user.Username, nil
}
