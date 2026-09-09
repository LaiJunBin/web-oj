package repository_test

import (
	"errors"
	"testing"

	"judge-backend/internal/app/shared/infrastructure/logger"
	"judge-backend/internal/app/shared/model"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/identity_domain/user/domain/entity/user"
	"judge-backend/internal/app/identity_domain/user/domain/entity/user/valueobject/role"
	"judge-backend/internal/app/identity_domain/user/infrastructure/repository"
	"judge-backend/internal/app/identity_domain/user/mocks"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type UserRepositoryMocks struct {
	gormDB        *gorm.DB
	sqlmock       sqlmock.Sqlmock
	roleConverter *mocks.RoleConverter
	logger        sharedtypes.Logger
}

func setupMocks(t *testing.T) *UserRepositoryMocks {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	assert.NoError(t, err)

	return &UserRepositoryMocks{
		gormDB:        gormDB,
		sqlmock:       mock,
		roleConverter: mocks.NewRoleConverter(t),
		logger:        logger.NewMockLogger(),
	}
}

func TestFindByID(t *testing.T) {
	mocks := setupMocks(t)
	repo := repository.NewUserRepository(mocks.gormDB, mocks.roleConverter, mocks.logger)

	t.Run("FindByID returns user", func(t *testing.T) {
		mockUser := &model.User{ID: 1, Name: "John Doe", Username: "johndoe", Roles: 3, Password: "password"}
		mockRoles := role.Roles{role.RoleAdmin, role.RoleContestant}
		mocks.roleConverter.On("ToRoles", mockUser.Roles).Return(mockRoles).Once()
		mocks.sqlmock.ExpectQuery("SELECT").WillReturnRows(
			sqlmock.NewRows([]string{"id", "name", "username", "password", "roles"}).
				AddRow(mockUser.ID, mockUser.Name, mockUser.Username, mockUser.Password, mockUser.Roles),
		)

		user, err := repo.FindByID(int(mockUser.ID))
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, mockUser.ID, user.GetID())
		assert.Equal(t, mockUser.Name, user.GetName())
		assert.Equal(t, mockUser.Username, user.GetUsername())
		assert.Equal(t, mockUser.Password, user.GetPassword())
		assert.Equal(t, mockRoles, user.GetRoles())
	})

	t.Run("FindByID returns nil when user not found", func(t *testing.T) {
		mocks.sqlmock.ExpectQuery("SELECT").WillReturnError(gorm.ErrRecordNotFound)

		user, err := repo.FindByID(1)
		assert.NoError(t, err)
		assert.Nil(t, user)
	})

	t.Run("FindByID returns error on DB failure", func(t *testing.T) {
		mocks.sqlmock.ExpectQuery("SELECT").WillReturnError(errors.New("db error"))

		_, err := repo.FindByID(1)
		assert.Error(t, err)
	})
}

func TestFindByUsername(t *testing.T) {
	mocks := setupMocks(t)
	repo := repository.NewUserRepository(mocks.gormDB, mocks.roleConverter, mocks.logger)
	t.Run("user found", func(t *testing.T) {
		mockUser := &model.User{ID: 1, Name: "User", Username: "username", Password: "password", Roles: 1}
		mockRoles := role.Roles{role.RoleAdmin}
		mocks.roleConverter.On("ToRoles", mockUser.Roles).Return(mockRoles).Once()
		mocks.sqlmock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "username", "password", "roles"}).AddRow(1, mockUser.Name, mockUser.Username, mockUser.Password, mockUser.Roles))

		user, err := repo.FindByUsername("username")

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, mockUser.ID, user.GetID())
		assert.Equal(t, mockUser.Name, user.GetName())
		assert.Equal(t, mockUser.Username, user.GetUsername())
		assert.Equal(t, mockUser.Password, user.GetPassword())
		assert.Equal(t, mockRoles, user.GetRoles())
	})

	t.Run("user not found", func(t *testing.T) {
		mocks.sqlmock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{}))
		user, err := repo.FindByUsername("username")
		assert.NoError(t, err)
		assert.Nil(t, user)
	})
}

func TestSave(t *testing.T) {
	mocks := setupMocks(t)
	repo := repository.NewUserRepository(mocks.gormDB, mocks.roleConverter, mocks.logger)

	t.Run("Save inserts new user", func(t *testing.T) {
		mocks.sqlmock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{}))
		mockUser := user.MustNew("John Doe", "johndoe", "password")
		mocks.roleConverter.On("ToBit", mockUser.GetRoles()).Return(0).Once()
		mocks.sqlmock.ExpectBegin()
		mocks.sqlmock.ExpectExec("INSERT INTO `users`").WillReturnResult(sqlmock.NewResult(1, 1))
		mocks.sqlmock.ExpectCommit()

		err := repo.Save(mockUser)
		assert.NoError(t, err)
		assert.Equal(t, 1, mockUser.GetID())
	})

	t.Run("Save updates existing user", func(t *testing.T) {
		mockModelUser := &model.User{ID: 1, Name: "John Doe", Username: "johndoe", Password: "password", Roles: 0}
		mockUser := user.MustNew("John Doe - updated", "johndoe - updated", "password - updated")
		mockUser.SetID(1)
		mockUser.SetRoles(role.RoleAPIReader)
		mocks.roleConverter.On("ToBit", mockUser.GetRoles()).Return(4).Once()
		mocks.sqlmock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "username", "password", "roles"}).AddRow(mockModelUser.ID, mockModelUser.Name, mockModelUser.Username, mockModelUser.Password, mockModelUser.Roles))
		mocks.sqlmock.ExpectBegin()
		mocks.sqlmock.ExpectExec("UPDATE `users`").WillReturnResult(sqlmock.NewResult(1, 1))
		mocks.sqlmock.ExpectCommit()

		err := repo.Save(mockUser)
		assert.NoError(t, err)
		assert.Equal(t, 1, mockUser.GetID())

		assert.Equal(t, "John Doe - updated", mockUser.GetName())
		assert.Equal(t, "johndoe - updated", mockUser.GetUsername())
		assert.Equal(t, "password - updated", mockUser.GetPassword())
		assert.Equal(t, role.Roles{role.RoleAPIReader}, mockUser.GetRoles())
	})

	t.Run("Save returns error on DB failure", func(t *testing.T) {
		mockUser := user.MustNew("John Doe", "johndoe", "password")
		mocks.roleConverter.On("ToBit", mockUser.GetRoles()).Return(0).Once()
		mocks.sqlmock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{}))
		mocks.sqlmock.ExpectBegin()
		mocks.sqlmock.ExpectExec("INSERT INTO `users`").WillReturnError(errors.New("db error"))
		mocks.sqlmock.ExpectRollback()

		err := repo.Save(mockUser)
		assert.Error(t, err)
	})
}

func TestDelete(t *testing.T) {
	mocks := setupMocks(t)
	repo := repository.NewUserRepository(mocks.gormDB, mocks.roleConverter, mocks.logger)

	t.Run("Delete removes user", func(t *testing.T) {
		mockUser := user.MustNew("John Doe", "johndoe", "password")
		mockUser.SetID(1)
		mocks.sqlmock.ExpectBegin()
		mocks.sqlmock.ExpectExec("UPDATE `users`").WillReturnResult(sqlmock.NewResult(1, 1))
		mocks.sqlmock.ExpectCommit()

		err := repo.Delete(mockUser)
		assert.NoError(t, err)
	})

	t.Run("Delete returns error on DB failure", func(t *testing.T) {
		mockUser := user.MustNew("John Doe", "johndoe", "password")
		mockUser.SetID(1)
		mocks.sqlmock.ExpectBegin()
		mocks.sqlmock.ExpectExec("UPDATE `users`").WillReturnError(errors.New("db error"))
		mocks.sqlmock.ExpectRollback()

		err := repo.Delete(mockUser)
		assert.Error(t, err)
	})
}

func TestSaveAll(t *testing.T) {
	mocks := setupMocks(t)
	repo := repository.NewUserRepository(mocks.gormDB, mocks.roleConverter, mocks.logger)

	t.Run("SaveAll inserts new users", func(t *testing.T) {
		mockUser1 := user.MustNew("John Doe", "johndoe", "password")
		mockUser2 := user.MustNew("Jane Doe", "janedoe", "password")
		mockUser1.SetRoles(role.RoleAdmin)
		mockUser2.SetRoles(role.RoleAPIReader)
		mockUsers := []*user.User{
			mockUser1,
			mockUser2,
		}
		mocks.sqlmock.ExpectBegin()
		mocks.sqlmock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{}))
		mocks.roleConverter.On("ToBit", mockUser2.GetRoles()).Return(4).Once()
		mocks.sqlmock.ExpectExec("INSERT INTO `users`").WillReturnResult(sqlmock.NewResult(1, 1))

		mocks.sqlmock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{}))
		mocks.roleConverter.On("ToBit", mockUser1.GetRoles()).Return(1).Once()
		mocks.sqlmock.ExpectExec("INSERT INTO `users`").WillReturnResult(sqlmock.NewResult(2, 1))
		mocks.sqlmock.ExpectCommit()

		err := repo.SaveAll(mockUsers)
		assert.NoError(t, err)
		assert.Equal(t, 1, mockUsers[0].GetID())
		assert.Equal(t, 2, mockUsers[1].GetID())
	})

	t.Run("SaveAll updates existing users", func(t *testing.T) {
		mockUser1 := user.MustNew("John Doe", "johndoe", "password")
		mockUser1.SetRoles(role.RoleAPIReader)

		mockModelUser2 := &model.User{ID: 2, Name: "Jane Doe", Username: "janedoe", Password: "password", Roles: 0}
		mockUser2 := user.MustNew("Jane Doe - Updated", "janedoe", "password")
		mockUser2.SetID(2)
		mockUser2.SetRoles(role.RoleAdmin)

		mocks.sqlmock.ExpectBegin()
		mocks.sqlmock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{}))
		mocks.roleConverter.On("ToBit", mockUser1.GetRoles()).Return(4).Once()
		mocks.sqlmock.ExpectExec("INSERT INTO `users`").WillReturnResult(sqlmock.NewResult(1, 1))

		mocks.sqlmock.ExpectQuery("SELECT").WillReturnRows(
			sqlmock.NewRows([]string{"id", "name", "username", "password", "roles"}).
				AddRow(mockModelUser2.ID, mockModelUser2.Name, mockModelUser2.Username, mockModelUser2.Password, mockModelUser2.Roles),
		)
		mocks.sqlmock.ExpectRollback()

		err := repo.SaveAll([]*user.User{mockUser1, mockUser2})
		assert.Error(t, err)
	})

	t.Run("SaveAll returns error should rollback on DB failure", func(t *testing.T) {
		mockUsers := []*user.User{
			user.MustNew("John Doe", "johndoe", "password"),
			user.MustNew("Jane Doe", "janedoe", "password"),
		}
		mocks.sqlmock.ExpectBegin()
		mocks.sqlmock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{}))
		mocks.roleConverter.On("ToBit", mockUsers[0].GetRoles()).Return(0).Once()
		mocks.sqlmock.ExpectExec("INSERT INTO `users`").WillReturnResult(sqlmock.NewResult(1, 1))
		mocks.sqlmock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{}))
		mocks.roleConverter.On("ToBit", mockUsers[1].GetRoles()).Return(0).Once()
		mocks.sqlmock.ExpectExec("INSERT INTO `users`").WillReturnError(errors.New("db error"))
		mocks.sqlmock.ExpectRollback()

		err := repo.SaveAll(mockUsers)
		assert.Error(t, err)
	})
}
