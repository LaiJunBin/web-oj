package updateuser_test

import (
	"judge-backend/internal/app/shared/infrastructure/logger"
	"judge-backend/internal/app/shared/test"
	updateuser "judge-backend/internal/app/identity_domain/user/application/usecase/update_user"
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
	usecase := updateuser.NewUsecase(mocks.userRepository, mocks.hashService, mocks.sessionRevoker, mocks.loginForgetter, logger)

	t.Run("successful update user", func(t *testing.T) {
		mockUser := user.MustNew("John Doe", "johndoe", "hashed_password")
		mocks.userRepository.On("FindByID", 1).Return(mockUser, nil).Once()
		mocks.hashService.On("Hash", "new_password").Return("new_hashed_password", nil).Once()
		mocks.userRepository.On("Save", mock.Anything).Return(nil).Once()
		mocks.sessionRevoker.On("RevokeAllSessions", 1).Return(nil).Once()
		// The reset clears the contestant's login throttling so they can sign in
		// with the new password immediately.
		mocks.loginForgetter.On("ForgetUser", "johndoe").Return().Once()
		input := &updateuser.Input{
			ID:       1,
			Name:     test.StringPtr("John Doe"),
			Password: test.StringPtr("new_password"),
			Roles:    test.Int32SlicePtr([]int32{1, 3}),
		}

		_, err := usecase.Execute(input)

		assert.NoError(t, err)
	})

	t.Run("failed update user with user not found", func(t *testing.T) {
		mocks.userRepository.On("FindByID", 1).Return(nil, nil).Once()
		input := &updateuser.Input{
			ID:   1,
			Name: test.StringPtr("John Doe"),
		}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
	})

	t.Run("failed update user with invalid role", func(t *testing.T) {
		mockUser := user.MustNew("John Doe", "johndoe", "hashed_password")
		mocks.userRepository.On("FindByID", 1).Return(mockUser, nil).Once()
		input := &updateuser.Input{
			ID:    1,
			Roles: test.Int32SlicePtr([]int32{1, 2, 100}),
		}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
	})

	t.Run("failed update user with internal error", func(t *testing.T) {
		mocks.userRepository.On("FindByID", 1).Return(nil, assert.AnError).Once()
		input := &updateuser.Input{
			ID:   1,
			Name: test.StringPtr("John Doe"),
		}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
	})

	// A password reset whose sessions survive is the failure this exists to
	// prevent, so a revocation error fails the request rather than reporting
	// a success the caller cannot rely on.
	t.Run("reports failure when sessions cannot be revoked", func(t *testing.T) {
		fresh := setupMocks(t)
		freshUsecase := updateuser.NewUsecase(fresh.userRepository, fresh.hashService, fresh.sessionRevoker, fresh.loginForgetter, logger)
		fresh.userRepository.On("FindByID", 5).Return(user.MustNew("Player", "player", "hashed"), nil).Once()
		fresh.hashService.On("Hash", "pw").Return("hashed_pw", nil).Once()
		fresh.userRepository.On("Save", mock.Anything).Return(nil).Once()
		fresh.sessionRevoker.On("RevokeAllSessions", 5).Return(assert.AnError).Once()

		_, err := freshUsecase.Execute(&updateuser.Input{ID: 5, Password: test.StringPtr("pw")})

		assert.Error(t, err)
	})

	// Renaming somebody is not a credential change; signing them out would
	// make routine edits disruptive mid-contest.
	t.Run("does not revoke sessions when the password is untouched", func(t *testing.T) {
		fresh := setupMocks(t)
		freshUsecase := updateuser.NewUsecase(fresh.userRepository, fresh.hashService, fresh.sessionRevoker, fresh.loginForgetter, logger)
		fresh.userRepository.On("FindByID", 6).Return(user.MustNew("Player", "player", "hashed"), nil).Once()
		fresh.userRepository.On("Save", mock.Anything).Return(nil).Once()

		_, err := freshUsecase.Execute(&updateuser.Input{ID: 6, Name: test.StringPtr("Renamed")})

		assert.NoError(t, err)
		fresh.sessionRevoker.AssertNotCalled(t, "RevokeAllSessions", mock.Anything)
	})
}
