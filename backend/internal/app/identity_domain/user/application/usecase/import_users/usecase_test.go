package importusers_test

import (
	"bytes"
	"judge-backend/internal/app/shared/infrastructure/logger"
	importusers "judge-backend/internal/app/identity_domain/user/application/usecase/import_users"
	"judge-backend/internal/app/identity_domain/user/mocks"
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
	usecase := importusers.NewUsecase(mocks.userRepository, mocks.hashService, logger)

	t.Run("successful import a user", func(t *testing.T) {
		mockJSON := bytes.NewBufferString(`[{ "name": "John Doe", "username": "johndoe", "password": "password", "roles": ["admin"] }]`)
		mocks.hashService.On("Hash", "password").Return("hashed_password", nil).Once()
		mocks.userRepository.On("SaveAll", mock.Anything).Return(nil).Once()
		input := &importusers.Input{JSONBuf: mockJSON}

		_, err := usecase.Execute(input)

		assert.NoError(t, err)
	})

	t.Run("successful import multiple users", func(t *testing.T) {
		mockJSON := bytes.NewBufferString(`[
			{ "name": "User01", "username": "user01", "password": "1234", "roles": ["admin"] },
			{ "name": "User02", "username": "user02", "password": "1234", "roles": ["contestant"] },
			{ "name": "User03", "username": "user03", "password": "1234", "roles": ["admin", "contestant"] },
			{ "name": "User04", "username": "user04", "password": "1234", "roles": ["api_reader"] }
		]`)
		mocks.hashService.On("Hash", "1234").Return("hashed_password", nil).Times(4)
		mocks.userRepository.On("SaveAll", mock.Anything).Return(nil).Once()
		input := &importusers.Input{JSONBuf: mockJSON}

		_, err := usecase.Execute(input)

		assert.NoError(t, err)
	})

	t.Run("failed import users with invalid JSON", func(t *testing.T) {
		mockJSON := bytes.NewBufferString(`invalid JSON`)
		input := &importusers.Input{JSONBuf: mockJSON}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
	})

	t.Run("failed import users with nil JSON", func(t *testing.T) {
		input := &importusers.Input{JSONBuf: nil}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
	})

	// The rename-on-delete namespace is reserved (same rule as create_user):
	// a roster row named like a renamed dead row would make deleting the
	// colliding user impossible.
	t.Run("rejects a reserved-marker username with the row and reason", func(t *testing.T) {
		// Fresh mocks: no SaveAll/Hash expectation is registered, so either
		// call would fail the test.
		fresh := setupMocks(t)
		freshUsecase := importusers.NewUsecase(fresh.userRepository, fresh.hashService, logger)
		mockJSON := bytes.NewBufferString(`[
			{ "name": "Evil", "username": "evil__deleted_7", "password": "1234", "roles": ["contestant"] }
		]`)
		input := &importusers.Input{JSONBuf: mockJSON}

		_, err := freshUsecase.Execute(input)

		var importErr *importusers.ImportUsersError
		assert.ErrorAs(t, err, &importErr)
		assert.Contains(t, importErr.Details(), "evil__deleted_7 contains a reserved word")
	})

	// Roles are resolved for the whole roster before any password is hashed,
	// so a bad role name in the last row does not cost bcrypt on every row
	// before it.
	t.Run("rejects an unknown role before hashing any password", func(t *testing.T) {
		// Fresh mocks: the shared ones above already recorded Hash calls, so
		// AssertNotCalled would be looking at another subtest's history.
		fresh := setupMocks(t)
		freshUsecase := importusers.NewUsecase(fresh.userRepository, fresh.hashService, logger)
		mockJSON := bytes.NewBufferString(`[
			{ "name": "Player", "username": "player", "password": "1234", "roles": ["contestant"] },
			{ "name": "Broken", "username": "broken", "password": "1234", "roles": ["not_a_role"] }
		]`)
		input := &importusers.Input{JSONBuf: mockJSON}

		_, err := freshUsecase.Execute(input)

		assert.Error(t, err)
		fresh.hashService.AssertNotCalled(t, "Hash", "1234")
	})
}
