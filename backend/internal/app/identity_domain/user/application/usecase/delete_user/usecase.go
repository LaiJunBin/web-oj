package deleteuser

import (
	"errors"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/identity_domain/user/application/service/hashservice"
	"judge-backend/internal/app/identity_domain/user/domain/repository"
	"judge-backend/internal/app/identity_domain/user/domain/service"
)

var (
	ErrorInvalidInput   = errors.New("invalid input")
	ErrorUserNotFound   = errors.New("user not found")
	ErrorInternalServer = errors.New("internal server error")
)

type Input struct {
	ID int32
}

type Output struct {
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	userRepository repository.UserRepository
	hashService    hashservice.HashService
	sessionRevoker service.SessionRevoker
	loginForgetter service.LoginAttemptForgetter
	logger         sharedtypes.Logger
}

func NewUsecase(
	userRepository repository.UserRepository,
	hashService hashservice.HashService,
	sessionRevoker service.SessionRevoker,
	loginForgetter service.LoginAttemptForgetter,
	logger sharedtypes.Logger,
) Usecase {
	return &usecase{
		userRepository: userRepository,
		hashService:    hashService,
		sessionRevoker: sessionRevoker,
		loginForgetter: loginForgetter,
		logger:         logger,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.ID < 1 {
		return ErrorInvalidInput
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	u.logger.Infof("Starting to delete user", "userId", input.ID)

	if err := u.validate(input); err != nil {
		u.logger.Errorf("Failed to validate input", "userId", input.ID, "error", err)
		return nil, err
	}

	user, err := u.userRepository.FindByID(int(input.ID))
	if err != nil {
		u.logger.Errorf("Failed to find user", "userId", input.ID, "error", err)
		return nil, ErrorInternalServer
	}

	if user == nil {
		u.logger.Infof("User not found", "userId", input.ID)
		return nil, ErrorUserNotFound
	}

	// Revocation runs BEFORE the delete: a deleted account's sessions must
	// not outlive it (answering 200 while a refresh token still mints access
	// tokens for it would be saying the opposite of the truth), and this is
	// the only order under which failure handling is honest. Revocation fails
	// here → nothing has changed → the 500 is true and a resend really
	// retries (the scoped FindByID above still resolves the user). The other
	// order deleted first, so a revocation failure 500'd an operation that
	// had already succeeded — and the resend hit the 404 branch above,
	// never reaching the revocation again. If instead the delete below
	// fails after a successful revocation, a live user has been logged out:
	// harmless, they log back in, and the resend runs both steps again.
	if err := u.sessionRevoker.RevokeAllSessions(int(input.ID)); err != nil {
		u.logger.Errorf("Failed to revoke sessions of user being deleted", "userId", input.ID, "error", err)
		return nil, ErrorInternalServer
	}

	if err := u.userRepository.Delete(user); err != nil {
		u.logger.Errorf("Failed to delete user", "userId", input.ID, "error", err)
		return nil, ErrorInternalServer
	}
	// Best-effort: clears the login limiter's per-user record so the freed
	// username starts clean if recreated. The entity predates the delete, so
	// GetUsername() is the original (pre-rename) name — the one the limiter
	// keyed on.
	u.loginForgetter.ForgetUser(user.GetUsername())

	u.logger.Infof("Successfully deleted user", "userId", input.ID)
	return &Output{}, nil
}
