package createuser

import (
	"errors"
	"judge-backend/internal/app/identity_domain/user/application/service/hashservice"
	"judge-backend/internal/app/identity_domain/user/domain/entity/user"
	"judge-backend/internal/app/identity_domain/user/domain/entity/user/valueobject/role"
	"judge-backend/internal/app/identity_domain/user/domain/entity/user/valueobject/username"
	"judge-backend/internal/app/identity_domain/user/domain/repository"
	sharedtypes "judge-backend/internal/app/shared/types"
)

var (
	ErrorNameIsRequired     = errors.New("name is required")
	ErrorUsernameIsRequired = errors.New("username is required")
	// The rename-on-delete namespace is reserved: a live user whose name
	// literally contains the marker could collide with a dead row's renamed
	// username, turning that row's delete into a unique-index failure — a
	// 500, and a user that can never be deleted.
	ErrorUsernameReserved   = errors.New("username contains a reserved word")
	ErrorPasswordIsRequired = errors.New("password is required")
	ErrorInvalidRoles       = errors.New("invalid roles")

	ErrorDuplicateUsername = errors.New("username already exists")
	ErrorInternalServer    = errors.New("internal server error")
)

type Input struct {
	Name     string
	Username string
	Password string
	Roles    []int32
}

type Output struct {
	ID int
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	userRepository repository.UserRepository
	hashService    hashservice.HashService
	logger         sharedtypes.Logger
}

func NewUsecase(
	userRepository repository.UserRepository,
	hashService hashservice.HashService,
	logger sharedtypes.Logger,
) Usecase {
	return &usecase{
		userRepository: userRepository,
		hashService:    hashService,
		logger:         logger,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.Name == "" {
		return ErrorNameIsRequired
	}

	if input.Username == "" {
		return ErrorUsernameIsRequired
	}

	if username.ContainsDeletedMarker(input.Username) {
		return ErrorUsernameReserved
	}

	if input.Password == "" {
		return ErrorPasswordIsRequired
	}

	for _, role := range input.Roles {
		if role < 1 {
			return ErrorInvalidRoles
		}
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	u.logger.Debugf("Starting to create user", "username", input.Username)

	if err := u.validate(input); err != nil {
		u.logger.Debugf("Failed to validate user input", "error", err)
		return nil, err
	}

	if user, err := u.userRepository.FindByUsername(input.Username); user != nil {
		u.logger.Debugf("Username already exists", "username", input.Username)
		return nil, ErrorDuplicateUsername
	} else if err != nil {
		u.logger.Errorf("Failed to check username existence", "error", err)
		return nil, ErrorInternalServer
	}

	hashedPassword, err := u.hashService.Hash(input.Password)
	if err != nil {
		u.logger.Errorf("Failed to hash password", "error", err)
		return nil, ErrorInternalServer
	}

	user, err := user.New(&user.UserConfig{
		Name:           input.Name,
		Username:       input.Username,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		u.logger.Errorf("Failed to create user entity", "error", err)
		return nil, ErrorInternalServer
	}

	roles, err := role.GetRolesByID(input.Roles)
	if err != nil {
		u.logger.Errorf("Failed to get roles by ID", "error", err)
		return nil, ErrorInvalidRoles
	}

	user.SetRoles(roles...)

	if err := u.userRepository.Save(user); err != nil {
		u.logger.Errorf("Failed to save user", "error", err)
		// The rename-on-delete fix is the PRIMARY cure for "recreate after
		// delete" (a renamed dead row no longer occupies the index); this
		// mapping is the backstop for a LIVE row racing the scoped
		// pre-check above — a friendly 400 instead of a bare 500.
		if errors.Is(err, repository.ErrDuplicateUsername) {
			return nil, ErrorDuplicateUsername
		}
		return nil, ErrorInternalServer
	}

	u.logger.Debugf("Successfully created user", "id", user.GetID(), "username", user.GetUsername())
	return &Output{
		ID: user.GetID(),
	}, nil
}
