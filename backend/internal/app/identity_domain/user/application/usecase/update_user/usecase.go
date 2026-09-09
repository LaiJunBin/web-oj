package updateuser

import (
	"errors"
	"judge-backend/internal/app/identity_domain/user/application/service/hashservice"
	"judge-backend/internal/app/identity_domain/user/domain/entity/user/valueobject/role"
	"judge-backend/internal/app/identity_domain/user/domain/repository"
	"judge-backend/internal/app/identity_domain/user/domain/service"
	sharedtypes "judge-backend/internal/app/shared/types"
)

var (
	ErrorInvalidInput   = errors.New("invalid input")
	ErrorUserNotFound   = errors.New("user not found")
	ErrorInvalidRole    = errors.New("invalid role")
	ErrorInternalServer = errors.New("internal server error")
)

type Input struct {
	ID       int32
	Name     *string
	Password *string
	Roles    *[]int32
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
	if input.Name == nil && input.Password == nil && input.Roles == nil {
		return ErrorInvalidInput
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	u.logger.Infof("Starting to update user", "userId", input.ID)

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

	if input.Name != nil {
		user.SetName(*input.Name)
	}

	if input.Password != nil {
		hashedPassword, err := u.hashService.Hash(*input.Password)
		if err != nil {
			u.logger.Errorf("Failed to hash password", "userId", input.ID, "error", err)
			return nil, ErrorInternalServer
		}
		user.SetPassword(hashedPassword)
	}

	if input.Roles != nil {
		roles, err := role.GetRolesByID(*input.Roles)
		if err != nil {
			u.logger.Errorf("Failed to get roles by ID", "userId", input.ID, "error", err)
			return nil, ErrorInvalidRole
		}

		user.SetRoles(roles...)
	}

	if err := u.userRepository.Save(user); err != nil {
		u.logger.Errorf("Failed to save user", "userId", input.ID, "error", err)
		return nil, ErrorInternalServer
	}

	// Revoked after the new password is persisted, not before: revoking first
	// would sign the account out even when the write then fails, for no gain.
	// A failure here is still reported as a failure — the password did change,
	// but the sessions the reset was meant to cut off are still live, and
	// answering 200 would say the opposite. The call is idempotent, so the
	// admin can simply repeat the request.
	if input.Password != nil {
		if err := u.sessionRevoker.RevokeAllSessions(int(input.ID)); err != nil {
			u.logger.Errorf("Password changed but sessions were not revoked", "userId", input.ID, "error", err)
			return nil, ErrorInternalServer
		}
		// Clear any login throttling so the contestant can sign in with the new
		// password right away, rather than being turned away by the failed
		// attempts that led to the reset. Best-effort and in-memory, so a
		// non-issue if it is a no-op.
		u.loginForgetter.ForgetUser(user.GetUsername())
	}

	u.logger.Infof("Successfully updated user", "userId", input.ID)
	return &Output{}, nil
}
