package repository_test

import (
	"judge-backend/internal/app/identity_domain/auth/domain/entity/user"
	"judge-backend/internal/app/identity_domain/auth/domain/entity/user/valueobject/role"
	"judge-backend/internal/app/identity_domain/auth/infrastructure/repository"
	"judge-backend/internal/app/identity_domain/auth/infrastructure/service"
	sysRole "judge-backend/internal/app/identity_domain/user/domain/entity/user/valueobject/role"
	"judge-backend/internal/app/identity_domain/user/mocks"
	"judge-backend/internal/app/shared/infrastructure/logger"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func newUserRepository() (*repository.UserRepository, sqlmock.Sqlmock, error) {
	db, sqlmock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}

	roleConverter := &mocks.RoleConverter{}
	roleConverter.On("ToRoles", mock.Anything).Return(sysRole.Roles{sysRole.RoleAdmin})
	logger := logger.NewMockLogger()
	service := service.NewUserRoleService(roleConverter, logger)
	r := repository.NewUserRepository(gormDB, service, logger)
	repo, ok := r.(*repository.UserRepository)
	if !ok {
		return nil, nil, err
	}

	return repo, sqlmock, nil
}

type expectedUser struct {
	Name     string
	Username string
	Password string
	ID       int
	Roles    []*role.Role
}

func assertUser(t *testing.T, user *user.User, expected expectedUser) {
	if user == nil {
		t.Fatalf("expected user, got nil")
	}

	if user.GetName() != expected.Name {
		t.Fatalf("expected name to be '%s', got %s", expected.Name, user.GetName())
	}

	if user.GetUsername() != expected.Username {
		t.Fatalf("expected username to be '%s', got %s", expected.Username, user.GetUsername())
	}

	if user.GetPassword() != expected.Password {
		t.Fatalf("expected password to be '%s', got %s", expected.Password, user.GetPassword())
	}

	if user.IsNew() {
		t.Fatalf("expected user to not be new")
	}

	if user.GetID() != expected.ID {
		t.Fatalf("expected id to be %d, got %d", expected.ID, user.GetID())
	}

	for _, role := range expected.Roles {
		if !user.HasRole(role) {
			t.Fatalf("expected user to have role %v", role)
		}
	}
}

func TestUserRepository(t *testing.T) {
	t.Run("TestFindByID", func(t *testing.T) {
		repo, mock, err := newUserRepository()
		if err != nil {
			t.Fatalf("error creating test repository: %v", err)
		}

		t.Run("user found", func(t *testing.T) {
			mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "username", "password", "roles"}).AddRow(1, "User", "username", "password", 1))
			user, err := repo.FindByID(1)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}

			expected := expectedUser{
				Name:     "User",
				Username: "username",
				Password: "password",
				ID:       1,
				Roles:    []*role.Role{role.RoleAdmin},
			}
			assertUser(t, user, expected)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("expected all mock expectations to be met, got %v", err)
			}
		})

		t.Run("user not found", func(t *testing.T) {
			mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{}))
			user, err := repo.FindByID(1)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}

			if user != nil {
				t.Fatalf("expected nil, got %v", user)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("expected all mock expectations to be met, got %v", err)
			}
		})
	})

	t.Run("TestFindByUsername", func(t *testing.T) {
		repo, mock, err := newUserRepository()
		if err != nil {
			t.Fatalf("error creating test repository: %v", err)
		}

		t.Run("user found", func(t *testing.T) {
			mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "username", "password", "roles"}).AddRow(1, "User", "username", "password", 1))
			user, err := repo.FindByUsername("username")
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}

			expected := expectedUser{
				Name:     "User",
				Username: "username",
				Password: "password",
				ID:       1,
				Roles:    []*role.Role{role.RoleAdmin},
			}
			assertUser(t, user, expected)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("expected all mock expectations to be met, got %v", err)
			}
		})

		t.Run("user not found", func(t *testing.T) {
			mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{}))
			user, err := repo.FindByUsername("username")
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}

			if user != nil {
				t.Fatalf("expected nil, got %v", user)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("expected all mock expectations to be met, got %v", err)
			}
		})
	})
}
