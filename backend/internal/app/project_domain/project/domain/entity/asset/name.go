package asset

import "errors"

var (
	ErrorInvalidName = errors.New("name is invalid, must not be empty")
)

type Name struct {
	value string
}

func NewName(value string) (Name, error) {
	name := Name{value: value}
	if err := name.validate(); err != nil {
		return Name{}, err
	}

	return name, nil
}

func MustNewName(value string) Name {
	name, err := NewName(value)
	if err != nil {
		panic(err)
	}

	return name
}

func (n Name) validate() error {
	if n.value == "" {
		return ErrorInvalidName
	}

	return nil
}

func (n Name) String() string {
	return n.value
}

func (n Name) Equals(other Name) bool {
	return n.value == other.value
}
