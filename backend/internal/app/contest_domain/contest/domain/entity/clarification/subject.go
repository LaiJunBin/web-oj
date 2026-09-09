package clarification

import (
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
)

var (
	ErrorInvalidClarificationSubject = httperror.New(http.StatusBadRequest, "invalid clarification subject")
)

type Subject struct {
	value string
}

func NewSubject(subject string) (Subject, error) {
	n := Subject{value: subject}
	if err := n.validate(); err != nil {
		return Subject{}, err
	}

	return n, nil
}

func MustNewSubject(subject string) Subject {
	s, err := NewSubject(subject)
	if err != nil {
		panic(err)
	}

	return s
}

func (s Subject) validate() error {
	if s.String() == "" {
		return ErrorInvalidClarificationSubject
	}

	return nil
}

func (s Subject) String() string {
	return s.value
}

func (s Subject) Equals(other Subject) bool {
	return s == other
}
