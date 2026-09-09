package lifecycle

import "errors"

var (
	ErrorInvalidLifecycleID = errors.New("invalid lifecycle id")
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
		return ErrorInvalidLifecycleID
	}

	return nil
}
func (i ID) Value() int32 {
	return int32(i)
}

func (i ID) Equals(other ID) bool {
	return i == other
}
