package types

import "context"

type Presenter[I any, O any] interface {
	Output(data *I) O
	Error(err error) O
}

type EsPresenter[I any, O any] interface {
	Output(ctx context.Context, data *I) O
	Error(err error) O
}
