package dynamicfield

import "errors"

var (
	ErrorInvalidLabel = errors.New("label is invalid, must not be empty")
)

type Label struct {
	value string
}

func NewLabel(label string) (Label, error) {
	l := Label{value: label}
	if err := l.validate(); err != nil {
		return Label{}, err
	}

	return l, nil
}

func MustNewLabel(label string) Label {
	l, err := NewLabel(label)
	if err != nil {
		panic(err)
	}

	return l
}

func (l Label) validate() error {
	if l.String() == "" {
		return ErrorInvalidLabel
	}

	return nil
}

func (l Label) String() string {
	return l.value
}

func (l Label) Equals(other Label) bool {
	return l == other
}
