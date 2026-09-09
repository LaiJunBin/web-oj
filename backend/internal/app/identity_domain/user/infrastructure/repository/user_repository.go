package repository

import (
	"errors"
	"strings"
	"time"

	"judge-backend/internal/app/shared/model"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/identity_domain/user/application/service/user/roleconverter"
	"judge-backend/internal/app/identity_domain/user/domain/entity/user"
	"judge-backend/internal/app/identity_domain/user/domain/repository"

	mysqldriver "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

// isDuplicateKeyError centralizes the dialect-specific duplicate detection —
// a verbatim copy of the image repository's helper of the same name (kept in
// sync by hand; both carry this cross-reference): gorm's translated sentinel,
// MySQL error 1062, or sqlite's UNIQUE-constraint message.
func isDuplicateKeyError(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	var mysqlErr *mysqldriver.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return true
	}
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}

type UserRepository struct {
	db            *gorm.DB
	roleConverter roleconverter.RoleConverter
	logger        sharedtypes.Logger
}

func NewUserRepository(
	db *gorm.DB,
	roleConverter roleconverter.RoleConverter,
	logger sharedtypes.Logger,
) repository.UserRepository {
	return &UserRepository{
		db:            db,
		roleConverter: roleConverter,
		logger:        logger,
	}
}

func (r *UserRepository) toDomainUser(u *model.User) (*user.User, error) {
	domainUser, err := user.New(&user.UserConfig{
		Name:           u.Name,
		Username:       u.Username,
		HashedPassword: u.Password,
	})
	if err != nil {
		return nil, err
	}

	err = domainUser.SetID(u.ID)
	if err != nil {
		return nil, err
	}

	domainUser.SetRoles(r.roleConverter.ToRoles(u.Roles)...)

	return domainUser, nil
}

func (r *UserRepository) toModelUser(u *user.User) (*model.User, error) {
	user := &model.User{}
	if err := r.db.First(user, u.GetID()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return user, nil
		}

		return nil, err
	}

	return user, nil
}

func (r *UserRepository) FindByID(id int) (*user.User, error) {
	r.logger.Debugf("Starting to find user by ID", "id", id)

	u := &model.User{}
	if err := r.db.First(u, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.logger.Debugf("User not found", "id", id)
			return nil, nil
		}

		r.logger.Errorf("Error finding user by ID", "id", id, "error", err)
		return nil, err
	}

	r.logger.Debugf("Successfully found user by ID", "id", id)
	return r.toDomainUser(u)
}

func (r *UserRepository) FindByUsername(username string) (*user.User, error) {
	r.logger.Debugf("Starting to find user by username", "username", username)

	u := &model.User{}
	if err := r.db.Where("username = ?", username).First(u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.logger.Debugf("User not found", "username", username)
			return nil, nil
		}

		r.logger.Errorf("Error finding user by username", "username", username, "error", err)
		return nil, err
	}

	r.logger.Debugf("Successfully found user by username", "username", username)
	return r.toDomainUser(u)
}

func (r *UserRepository) Save(user *user.User) error {
	r.logger.Debugf("Starting to save user", "id", user.GetID(), "username", user.GetUsername())

	u, err := r.toModelUser(user)
	if err != nil {
		r.logger.Errorf("Error converting user to model", "id", user.GetID(), "error", err)
		return err
	}

	roles := r.roleConverter.ToBit(user.GetRoles())
	u.Name = user.GetName()
	u.Username = user.GetUsername()
	u.Password = user.GetPassword()
	u.Roles = roles

	if err := r.db.Save(u).Error; err != nil {
		r.logger.Errorf("Error saving user to database", "id", user.GetID(), "error", err)
		if isDuplicateKeyError(err) {
			return repository.ErrDuplicateUsername
		}
		return err
	}

	err = user.SetID(u.ID)
	if err != nil {
		r.logger.Errorf("Error setting user ID", "id", user.GetID(), "error", err)
		return err
	}

	r.logger.Debugf("Successfully saved user", "id", user.GetID(), "username", user.GetUsername())
	return nil
}

// Delete soft-deletes AND renames the username in ONE UPDATE: the unique
// username index has no partial-index WHERE clause (on any database), so a
// soft-deleted row that kept its name blocked recreating that username
// forever. Same scheme as the image repository's softDeleteFreeingTag. The
// entity was just resolved by the caller (delete_user's scoped FindByID), so
// GetUsername() is the live row's original name — no re-query needed. A
// RowsAffected of 0 (the row was concurrently soft-deleted already) keeps
// the old gorm-Delete idempotent-success semantics.
func (r *UserRepository) Delete(user *user.User) error {
	r.logger.Debugf("Starting to delete user", "id", user.GetID())

	res := r.db.Model(&model.User{}).Unscoped().
		Where("id = ? AND deleted_at IS NULL", user.GetID()).
		Updates(map[string]interface{}{
			"deleted_at": time.Now(),
			"username":   model.DeletedUsername(user.GetUsername(), user.GetID()),
		})
	if res.Error != nil {
		r.logger.Errorf("Error deleting user", "id", user.GetID(), "error", res.Error)
		return res.Error
	}

	r.logger.Debugf("Successfully deleted user", "id", user.GetID())
	return nil
}

func (r *UserRepository) SaveAll(users []*user.User) (err error) {
	r.logger.Debugf("Starting to save all users", "count", len(users))

	tx := r.db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
			r.logger.Errorf("Error saving all users, rolling back transaction", "error", err)
		}
	}()

	for i, u := range users {
		r.logger.Debugf("Processing user for save all", "index", i, "username", u.GetUsername())

		user := &model.User{}
		if err = tx.First(user, "username = ?", u.GetUsername()).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				r.logger.Errorf("Error finding user for save all", "index", i, "username", u.GetUsername(), "error", err)
				err = repository.NewSaveAllError(i)
				return
			}
		}

		if user.ID != 0 {
			r.logger.Debugf("User already exists for save all", "index", i, "username", u.GetUsername())
			err = repository.NewSaveAllExistingUserError(i)
			return
		}

		roles := r.roleConverter.ToBit(u.GetRoles())
		user.Name = u.GetName()
		user.Username = u.GetUsername()
		user.Password = u.GetPassword()
		user.Roles = roles

		if err = tx.Save(user).Error; err != nil {
			r.logger.Errorf("Error saving user in save all", "index", i, "username", u.GetUsername(), "error", err)
			if isDuplicateKeyError(err) {
				// A duplicate here is a LIVE row racing the scoped pre-check
				// above — report it as the existing-user case so the import
				// error names the username instead of blaming the row format.
				err = repository.NewSaveAllExistingUserError(i)
				return
			}
			err = repository.NewSaveAllError(i)
			return
		}

		err = u.SetID(user.ID)
		if err != nil {
			r.logger.Errorf("Error setting user ID in save all", "index", i, "username", u.GetUsername(), "error", err)
			err = repository.NewSaveAllError(i)
			return
		}
	}

	if err = tx.Commit().Error; err != nil {
		r.logger.Errorf("Error committing transaction for save all", "error", err)
		return
	}

	r.logger.Debugf("Successfully saved all users", "count", len(users))
	return
}
