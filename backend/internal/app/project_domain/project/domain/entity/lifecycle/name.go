package lifecycle

import "errors"

var (
	ErrorInvalidLifecycleName = errors.New("invalid lifecycle name")
)

type Name struct {
	value string
}

func NewName(name string) (Name, error) {
	n := Name{value: name}
	if err := n.validate(); err != nil {
		return Name{}, err
	}

	return n, nil
}

func MustNewName(name string) Name {
	n, err := NewName(name)
	if err != nil {
		panic(err)
	}

	return n
}

func (n Name) validate() error {
	if n.String() == "" {
		return ErrorInvalidLifecycleName
	}

	return nil
}

func (n Name) String() string {
	return n.value
}

func (n Name) Equals(other Name) bool {
	return n == other
}
