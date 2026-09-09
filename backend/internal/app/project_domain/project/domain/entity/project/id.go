package project

import "errors"

var (
	ErrorInvalidProjectID = errors.New("invalid project id")
)

type ID int32

func NewID(id int32) (ID, error) {
	i := ID(id)
	if err := i.validate(); err != nil {
		return 0, err
	}

	return i, nil
}

func MustNewID(id int32) ID {
	i, err := NewID(id)
	if err != nil {
		panic(err)
	}

	return i
}

func (i ID) validate() error {
	if i < 0 {
		return ErrorInvalidProjectID
	}

	return nil
}
func (i ID) Value() int32 {
	return int32(i)
}

func (i ID) Equals(other ID) bool {
	return i == other
}
