package common

type Prioritizable[T any] interface {
	Name() string
	Priority(i, j T) bool
	GetExtraFields(T) map[string]interface{}
}
