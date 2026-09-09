package asset

import (
	"errors"
	"time"
)

var (
	ErrInvalidLastModified = errors.New("invalid last modified timestamp")
)

type LastModified struct {
	value time.Time
}

func NewLastModified(value time.Time) (LastModified, error) {
	lm := LastModified{value: value}
	if err := lm.validate(); err != nil {
		return LastModified{}, ErrInvalidLastModified
	}

	return lm, nil
}

func MustNewLastModified(value time.Time) LastModified {
	lm, err := NewLastModified(value)
	if err != nil {
		panic(err)
	}
	return lm
}

func NowLastModified() LastModified {
	return LastModified{value: time.Now()}
}

func (lm LastModified) validate() error {
	if lm.value.IsZero() {
		return ErrInvalidLastModified
	}
	return nil
}

func (lm LastModified) Value() time.Time {
	return lm.value
}

func (lm LastModified) String() string {
	return lm.value.Format(time.RFC3339)
}

func (lm LastModified) Equals(other LastModified) bool {
	return lm.value.Equal(other.value)
}
