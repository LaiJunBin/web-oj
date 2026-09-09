package importusers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/identity_domain/user/application/service/hashservice"
	"judge-backend/internal/app/identity_domain/user/domain/entity/user"
	"judge-backend/internal/app/identity_domain/user/domain/entity/user/valueobject/role"
	"judge-backend/internal/app/identity_domain/user/domain/entity/user/valueobject/username"
	"judge-backend/internal/app/identity_domain/user/domain/repository"
)

var (
	ErrorInvalidJSON    = errors.New("invalid json")
	ErrorInternalServer = errors.New("internal server error")

	ErrorImportUsers = &ImportUsersError{}
)

type Input struct {
	JSONBuf *bytes.Buffer
}

type Output struct {
	TotalImported int
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

type ImportUser struct {
	Name     string   `json:"name"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	Roles    []string `json:"roles"`
}

func (uc *usecase) Execute(input *Input) (*Output, error) {
	uc.logger.Infof("Starting to import users")

	if input.JSONBuf == nil {
		uc.logger.Errorf("Invalid JSON: buffer is nil")
		return nil, ErrorInvalidJSON
	}

	var importUsers []ImportUser
	if err := json.Unmarshal(input.JSONBuf.Bytes(), &importUsers); err != nil {
		uc.logger.Errorf("Failed to unmarshal JSON", "error", err)
		return nil, ErrorInvalidJSON
	}

	uc.logger.Infof("Found users to import", "count", len(importUsers))

	// Roles are resolved for the whole file before any account is built, so a
	// bad role name in the last row is reported without having spent bcrypt on
	// every row before it.
	rowRoles := make([]role.Roles, len(importUsers))
	for i, u := range importUsers {
		// The rename-on-delete namespace is reserved (same rule as
		// create_user): a roster row named like a renamed dead row would make
		// deleting the colliding user impossible.
		if username.ContainsDeletedMarker(u.Username) {
			uc.logger.Errorf("Username uses the reserved deleted-marker namespace", "username", u.Username)
			return nil, ErrorImportUsers.
				WithFailedRow(u).
				AddDetail(fmt.Sprintf("%s contains a reserved word", u.Username))
		}

		roles, err := role.GetRolesByNames(u.Roles)
		if err != nil {
			uc.logger.Errorf("Failed to get roles by names", "username", u.Username, "error", err)
			return nil, ErrorImportUsers.WithFailedRow(u)
		}

		rowRoles[i] = roles
	}

	users := make([]*user.User, 0, len(importUsers))
	for i, u := range importUsers {
		hashedPassword, err := uc.hashService.Hash(u.Password)
		if err != nil {
			uc.logger.Errorf("Failed to hash password for user", "username", u.Username, "error", err)
			return nil, ErrorImportUsers.WithFailedRow(u)
		}

		user, err := user.New(&user.UserConfig{
			Name:           u.Name,
			Username:       u.Username,
			HashedPassword: hashedPassword,
		})
		if err != nil {
			uc.logger.Errorf("Failed to create user entity", "username", u.Username, "error", err)
			return nil, ErrorImportUsers.WithFailedRow(u)
		}

		user.SetRoles(rowRoles[i]...)

		users = append(users, user)
	}

	if err := uc.userRepository.SaveAll(users); err != nil {
		var saveAllError *repository.SaveAllError
		var saveAllExistingUserError *repository.SaveAllExistingUserError
		switch {
		case errors.As(err, &saveAllError):
			uc.logger.Errorf("Failed to save user", "index", saveAllError.Index(), "error", err)
			return nil, ErrorImportUsers.WithFailedRow(importUsers[saveAllError.Index()])
		case errors.As(err, &saveAllExistingUserError):
			uc.logger.Errorf("User already exists", "username", importUsers[saveAllExistingUserError.Index()].Username)
			return nil, ErrorImportUsers.
				WithFailedRow(importUsers[saveAllExistingUserError.Index()]).
				AddDetail(fmt.Sprintf("%s already exists", importUsers[saveAllExistingUserError.Index()].Username))
		default:
			uc.logger.Errorf("Failed to save users", "error", err)
			return nil, ErrorInternalServer
		}
	}

	uc.logger.Infof("Successfully imported users", "count", len(users))
	return &Output{
		TotalImported: len(users),
	}, nil
}
