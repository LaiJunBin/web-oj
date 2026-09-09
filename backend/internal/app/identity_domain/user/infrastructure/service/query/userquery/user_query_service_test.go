package userquery

import (
	"errors"
	"testing"
	"time"

	"judge-backend/internal/app/shared/infrastructure/logger"
	"judge-backend/internal/app/shared/model"
	"judge-backend/internal/app/identity_domain/user/application/service/query/userquery"
	"judge-backend/internal/app/identity_domain/user/domain/entity/user/valueobject/role"
	"judge-backend/internal/app/identity_domain/user/mocks"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type UserQueryServiceMocks struct {
	gormDB        *gorm.DB
	sqlmock       sqlmock.Sqlmock
	roleConverter *mocks.RoleConverter
}

func setupMocks(t *testing.T) *UserQueryServiceMocks {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	assert.NoError(t, err)

	return &UserQueryServiceMocks{
		gormDB:        gormDB,
		sqlmock:       mock,
		roleConverter: mocks.NewRoleConverter(t),
	}
}

func TestListUser(t *testing.T) {
	mocks := setupMocks(t)
	logger := logger.NewMockLogger()
	service := NewUserQueryService(mocks.gormDB, mocks.roleConverter, logger)

	t.Run("ListUser returns all users", func(t *testing.T) {
		mockUsers := []model.User{
			{ID: 1, Name: "John Doe", Username: "johndoe", Roles: 3, Password: "password", Model: gorm.Model{CreatedAt: time.Now(), UpdatedAt: time.Now()}},
		}
		mocks.roleConverter.On("ToRoles", mockUsers[0].Roles).Return(role.Roles{
			role.RoleAdmin,
			role.RoleContestant,
		}).Once()
		mocks.sqlmock.ExpectQuery("SELECT \\* FROM `users`").WillReturnRows(
			sqlmock.NewRows([]string{"id", "name", "username", "password", "roles", "created_at", "updated_at"}).
				AddRow(mockUsers[0].ID, mockUsers[0].Name, mockUsers[0].Username, mockUsers[0].Password, mockUsers[0].Roles, mockUsers[0].CreatedAt, mockUsers[0].UpdatedAt),
		)

		users, err := service.ListUser()
		assert.NoError(t, err)
		assert.Len(t, users, 1)

		expected := []userquery.User{
			{ID: 1, Name: "John Doe", Username: "johndoe", Roles: []string{"admin", "contestant"}, CreatedAt: mockUsers[0].CreatedAt, UpdatedAt: mockUsers[0].UpdatedAt},
		}
		assert.Equal(t, expected, users)
	})

	t.Run("ListUser returns error on DB failure", func(t *testing.T) {
		mocks.sqlmock.ExpectQuery("SELECT \\* FROM `users`").WillReturnError(errors.New("db error"))

		_, err := service.ListUser()
		assert.Error(t, err)
	})
}

func TestGetUserByID(t *testing.T) {
	mocks := setupMocks(t)
	logger := logger.NewMockLogger()
	service := NewUserQueryService(mocks.gormDB, mocks.roleConverter, logger)

	t.Run("GetUserByID returns user", func(t *testing.T) {
		mockUser := model.User{ID: 1, Name: "John Doe", Username: "johndoe", Roles: 3, Password: "password", Model: gorm.Model{CreatedAt: time.Now(), UpdatedAt: time.Now()}}
		mockRoles := role.Roles{role.RoleAdmin, role.RoleContestant}
		mocks.roleConverter.On("ToRoles", mockUser.Roles).Return(mockRoles).Once()
		mocks.sqlmock.ExpectQuery("SELECT").WillReturnRows(
			sqlmock.NewRows([]string{"id", "name", "username", "password", "roles", "created_at", "updated_at"}).
				AddRow(mockUser.ID, mockUser.Name, mockUser.Username, mockUser.Password, mockUser.Roles, mockUser.CreatedAt, mockUser.UpdatedAt),
		)

		user, err := service.GetUserByID(int32(mockUser.ID))
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, int32(mockUser.ID), user.ID)
		assert.Equal(t, mockUser.Name, user.Name)
		assert.Equal(t, mockUser.Username, user.Username)
		assert.Equal(t, []string{"admin", "contestant"}, user.Roles)
		assert.Equal(t, mockUser.CreatedAt, user.CreatedAt)
		assert.Equal(t, mockUser.UpdatedAt, user.UpdatedAt)
	})

	t.Run("GetUserByID returns nil when user not found", func(t *testing.T) {
		mocks.sqlmock.ExpectQuery("SELECT").WillReturnError(gorm.ErrRecordNotFound)

		user, err := service.GetUserByID(1)
		assert.ErrorIs(t, err, userquery.ErrorUserNotFound)
		assert.Nil(t, user)
	})

	t.Run("GetUserByID returns error on DB failure", func(t *testing.T) {
		mocks.sqlmock.ExpectQuery("SELECT").WillReturnError(errors.New("db error"))

		_, err := service.GetUserByID(1)
		assert.Error(t, err)
	})
}
