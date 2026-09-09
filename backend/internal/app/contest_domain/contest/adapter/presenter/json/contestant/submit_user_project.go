package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/contest/adapter/presenter/types"
	submituserproject "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/submit_user_project"
	commonpresenter "judge-backend/internal/app/shared/adapter/presenter/common"
)

type SubmitUserProjectPresenter types.Presenter[submituserproject.Output, api.SubmitUserProjectResponseObject]

// This operation used to be served by the shared JSON presenter, and still is
// for everything except the cooldown rejection — 429 is the one response that
// carries a field (retryAfterSeconds) the shared error shape has no room for.
// Delegating the rest keeps the success and error bodies byte-identical to
// what clients already receive; only the new status code is new.
type submitUserProjectPresenter struct {
	shared *commonpresenter.JSONPresenter
}

func NewSubmitUserProjectPresenter() SubmitUserProjectPresenter {
	return &submitUserProjectPresenter{shared: commonpresenter.NewJSONPresenter()}
}

func (p *submitUserProjectPresenter) Output(output *submituserproject.Output) api.SubmitUserProjectResponseObject {
	return p.shared.Output(output)
}

// Error maps the cooldown rejection onto 429 explicitly. Every other error
// goes through the shared error response, which has no room for
// retryAfterSeconds — and that field is the point: the client shows a
// countdown instead of parsing a message it would have to keep in sync.
func (p *submitUserProjectPresenter) Error(err error) api.SubmitUserProjectResponseObject {
	var cooldownErr *submituserproject.SubmissionCooldownError
	if errors.As(err, &cooldownErr) {
		return api.SubmitUserProject429JSONResponse{
			Success:           false,
			Message:           cooldownErr.Error(),
			Details:           []string{},
			RetryAfterSeconds: cooldownErr.RetryAfterSeconds,
		}
	}

	return p.shared.Error(err)
}
