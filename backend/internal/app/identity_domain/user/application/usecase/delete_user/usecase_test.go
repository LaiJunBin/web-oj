package deleteuser_test

import (
	"judge-backend/internal/app/shared/infrastructure/logger"
	deleteuser "judge-backend/internal/app/identity_domain/user/application/usecase/delete_user"
	"judge-backend/internal/app/identity_domain/user/domain/entity/user"
	"judge-backend/internal/app/identity_domain/user/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type usecaseMocks struct {
	userRepository *mocks.UserRepository
	hashService    *mocks.HashService
	sessionRevoker *mocks.SessionRevoker
	loginForgetter *mocks.LoginAttemptForgetter
}

func setupMocks(t *testing.T) *usecaseMocks {
	return &usecaseMocks{
		userRepository: mocks.NewUserRepository(t),
		hashService:    mocks.NewHashService(t),
		sessionRevoker: mocks.NewSessionRevoker(t),
		loginForgetter: mocks.NewLoginAttemptForgetter(t),
	}
}
func TestExecute(t *testing.T) {
	mocks := setupMocks(t)
	logger := logger.NewMockLogger()
	usecase := deleteuser.NewUsecase(mocks.userRepository, mocks.hashService, mocks.sessionRevoker, mocks.loginForgetter, logger)

	t.Run("successful delete user", func(t *testing.T) {
		mockUser := user.MustNew("John Doe", "johndoe", "hashed_password")
		mocks.userRepository.On("FindByID", 1).Return(mockUser, nil).Once()
		mocks.userRepository.On("Delete", mock.Anything).Return(nil).Once()
		// A deleted account's sessions must not outlive it, and the login
		// limiter forgets the freed (original) username.
		mocks.sessionRevoker.On("RevokeAllSessions", 1).Return(nil).Once()
		mocks.loginForgetter.On("ForgetUser", "johndoe").Return().Once()
		input := &deleteuser.Input{ID: 1}

		_, err := usecase.Execute(input)

		assert.NoError(t, err)
	})

	t.Run("revoke failure leaves the user undeleted", func(t *testing.T) {
		// Revocation runs BEFORE the delete, so a failed revocation must
		// leave everything untouched — that is what makes the 500 honest and
		// the resend effective (a scoped FindByID would 404 an already-deleted
		// user, so revocation could never be retried the other way around).
		// Strict mocks double as the negative assertions: neither Delete nor
		// ForgetUser is registered, so either call would fail the test.
		mockUser := user.MustNew("John Doe", "johndoe", "hashed_password")
		mocks.userRepository.On("FindByID", 1).Return(mockUser, nil).Once()
		mocks.sessionRevoker.On("RevokeAllSessions", 1).Return(assert.AnError).Once()
		input := &deleteuser.Input{ID: 1}

		_, err := usecase.Execute(input)

		assert.ErrorIs(t, err, deleteuser.ErrorInternalServer)
	})

	t.Run("failed delete user with user not found", func(t *testing.T) {
		mocks.userRepository.On("FindByID", 1).Return(nil, nil).Once()
		input := &deleteuser.Input{ID: 1}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
	})

	t.Run("failed delete user with internal error", func(t *testing.T) {
		mocks.userRepository.On("FindByID", 1).Return(nil, assert.AnError).Once()
		input := &deleteuser.Input{ID: 1}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
	})
}
