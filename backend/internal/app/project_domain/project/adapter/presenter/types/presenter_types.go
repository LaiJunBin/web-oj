package types

type Presenter[I any, O any] interface {
	Output(data *I) O
	Error(err error) O
}
