package project

import (
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
)

var (
	ErrorInvalidProjectName = httperror.New(http.StatusBadRequest, "invalid project name")
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
		return ErrorInvalidProjectName
	}

	return nil
}

func (n Name) String() string {
	return n.value
}

func (n Name) Equals(other Name) bool {
	return n == other
}
