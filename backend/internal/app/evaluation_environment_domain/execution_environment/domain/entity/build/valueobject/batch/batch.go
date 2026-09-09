package batch

import "errors"

var (
	ErrInvalidBatch = errors.New("invalid batch, must be greater than or equal to 0")
)

type Batch struct {
	value int
}

func New(value int) (Batch, error) {
	b := Batch{value: value}
	if err := b.Validate(); err != nil {
		return Batch{}, err
	}

	return b, nil
}

func MustNew(value int) Batch {
	b, err := New(value)
	if err != nil {
		panic(err)
	}

	return b
}

func (b Batch) Value() int {
	return b.value
}

func (b Batch) Validate() error {
	if b.value < 0 {
		return ErrInvalidBatch
	}

	return nil
}
