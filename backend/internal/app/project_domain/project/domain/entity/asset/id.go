package asset

import "errors"

var (
	ErrorInvalidID = errors.New("id is invalid, must not be empty")
)

type ID struct {
	value string
}

func NewID(value string) (ID, error) {
	id := ID{value: value}
	if err := id.validate(); err != nil {
		return ID{}, nil
	}

	return id, nil
}

func MustNewID(value string) ID {
	id, err := NewID(value)
	if err != nil {
		panic(err)
	}

	return id
}

func (i ID) validate() error {
	if i.value == "" {
		return ErrorInvalidID
	}

	return nil
}

func (i ID) String() string {
	return i.value
}

func (i ID) Equals(other ID) bool {
	return i.value == other.value
}
