package repository

import (
	"errors"

	"judge-backend/internal/app/identity_domain/user/domain/entity/user"
)

// ErrDuplicateUsername is Save's verdict when the username's unique index
// rejected the write: another LIVE row already holds the name. (Soft-deleted
// rows no longer occupy names — deletion renames them, see
// model.DeletedUsername — so this is the concurrent-create backstop, not the
// after-delete case.)
var ErrDuplicateUsername = errors.New("duplicate username")

type UserRepository interface {
	FindByID(id int) (*user.User, error)
	FindByUsername(username string) (*user.User, error)
	Save(user *user.User) error
	Delete(user *user.User) error

	SaveAll(users []*user.User) error
}

type SaveAllError struct {
	index int
}

func (e *SaveAllError) Error() string {
	return "failed to save all users"
}

func NewSaveAllError(index int) *SaveAllError {
	return &SaveAllError{index: index}
}

func (e *SaveAllError) Index() int {
	return e.index
}

type SaveAllExistingUserError struct {
	index int
}

func (e *SaveAllExistingUserError) Error() string {
	return "failed to save all users because some users already exist"
}

func NewSaveAllExistingUserError(index int) *SaveAllExistingUserError {
	return &SaveAllExistingUserError{index: index}
}

func (e *SaveAllExistingUserError) Index() int {
	return e.index
}
