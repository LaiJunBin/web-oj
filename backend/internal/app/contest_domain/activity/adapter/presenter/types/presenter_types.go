package types

import "context"

type Presenter[I any, O any] interface {
	Output(data *I) O
	Error(err error) O
}

// EsPresenter is Presenter for a server-sent-event response: Output is handed
// the request's context so the presenter can stop writing when the client goes
// away. Mirrors the same alias in the worker subdomain's presenter types.
type EsPresenter[I any, O any] interface {
	Output(ctx context.Context, data *I) O
	Error(err error) O
}
