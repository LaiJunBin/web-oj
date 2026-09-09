package userid

import "errors"

var (
	ErrInvalidUserID = errors.New("user id must be greater than or equal to 0")
)

type UserID int

func New(id int) (UserID, error) {
	uid := UserID(id)
	if err := uid.Validate(); err != nil {
		return 0, err
	}

	return uid, nil
}

func MustNew(id int) UserID {
	uid, err := New(id)
	if err != nil {
		panic(err)
	}

	return uid
}

func (id UserID) Value() int {
	return int(id)
}

func (id UserID) Equals(other UserID) bool {
	return id == other
}

func (id UserID) Validate() error {
	if id < 0 {
		return ErrInvalidUserID
	}

	return nil
}
