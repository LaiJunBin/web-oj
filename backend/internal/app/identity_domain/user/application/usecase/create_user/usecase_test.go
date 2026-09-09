package createuser_test

import (
	createuser "judge-backend/internal/app/identity_domain/user/application/usecase/create_user"
	"judge-backend/internal/app/identity_domain/user/domain/entity/user"
	"judge-backend/internal/app/identity_domain/user/mocks"
	"judge-backend/internal/app/shared/infrastructure/logger"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type usecaseMocks struct {
	userRepository *mocks.UserRepository
	hashService    *mocks.HashService
}

func setupMocks(t *testing.T) *usecaseMocks {
	return &usecaseMocks{
		userRepository: mocks.NewUserRepository(t),
		hashService:    mocks.NewHashService(t),
	}
}

func TestExecute(t *testing.T) {
	mocks := setupMocks(t)
	logger := logger.NewMockLogger()
	usecase := createuser.NewUsecase(mocks.userRepository, mocks.hashService, logger)

	t.Run("successful create user", func(t *testing.T) {
		mocks.userRepository.On("FindByUsername", "johndoe").Return(nil, nil).Once()
		mocks.hashService.On("Hash", "password").Return("hashed_password", nil).Once()
		mocks.userRepository.On("Save", mock.Anything).Return(nil).Once()
		input := &createuser.Input{
			Name:     "John Doe",
			Username: "johndoe",
			Password: "password",
			Roles:    []int32{1, 2},
		}

		_, err := usecase.Execute(input)

		assert.NoError(t, err)
	})

	t.Run("rejects a reserved-marker username before touching anything", func(t *testing.T) {
		// The rename-on-delete namespace is reserved (see the usecase's
		// ErrorUsernameReserved). Strict shared mocks double as the negative
		// assertions: no expectations are registered here, so any
		// FindByUsername/Hash/Save call would fail the test.
		input := &createuser.Input{
			Name:     "Evil",
			Username: "evil__deleted_7",
			Password: "password",
			Roles:    []int32{1},
		}

		_, err := usecase.Execute(input)

		assert.ErrorIs(t, err, createuser.ErrorUsernameReserved)
	})

	t.Run("failed create user with duplicate username", func(t *testing.T) {
		mockUser := user.MustNew("John Doe", "johndoe", "hashed_password")
		mocks.userRepository.On("FindByUsername", "john").Return(mockUser, nil).Once()
		input := &createuser.Input{
			Name:     "John Doe",
			Username: "john",
			Password: "password",
			Roles:    []int32{},
		}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
	})

	t.Run("failed create user with internal error", func(t *testing.T) {
		mocks.userRepository.On("FindByUsername", "johndoe").Return(nil, assert.AnError).Once()
		input := &createuser.Input{
			Name:     "John Doe",
			Username: "johndoe",
			Password: "password",
			Roles:    []int32{},
		}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
	})

	t.Run("failed create user with invalid role", func(t *testing.T) {
		mocks.userRepository.On("FindByUsername", "johndoe").Return(nil, nil).Once()
		mocks.hashService.On("Hash", "password").Return("hashed_password", nil).Once()
		input := &createuser.Input{
			Name:     "John Doe",
			Username: "johndoe",
			Password: "password",
			Roles:    []int32{1, 2, 100},
		}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
	})
}
