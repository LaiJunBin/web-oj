package imagemountpath

import "errors"

var (
	ErrorInvalidMountPath = errors.New("invalid mount path")
)

type MountPath struct {
	value string
}

func New(path string) (MountPath, error) {
	m := MountPath{value: path}
	if err := m.validate(); err != nil {
		return MountPath{}, err
	}

	return m, nil
}

func MustNew(path string) MountPath {
	m, err := New(path)
	if err != nil {
		panic(err)
	}

	return m
}

func (m MountPath) validate() error {
	if m.value == "" {
		return ErrorInvalidMountPath
	}

	if m.value[0] != '/' {
		return ErrorInvalidMountPath
	}

	return nil
}

func (m MountPath) String() string {
	return m.value
}

func (m MountPath) Equals(other MountPath) bool {
	return m.value == other.value
}
