package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/shared/types/httperror"
	"judge-backend/internal/app/contest_domain/activity/adapter/presenter/types"
	updateclarification "judge-backend/internal/app/contest_domain/activity/application/usecase/clarification/update_clarification"
)

type MarkClarificationAsUnansweredPresenter types.Presenter[updateclarification.Output, api.MarkClarificationAsUnansweredResponseObject]

type markClarificationAsUnansweredPresenter struct{}

func NewMarkClarificationAsUnansweredPresenter() MarkClarificationAsUnansweredPresenter {
	return &markClarificationAsUnansweredPresenter{}
}

func (p *markClarificationAsUnansweredPresenter) Output(output *updateclarification.Output) api.MarkClarificationAsUnansweredResponseObject {
	return api.MarkClarificationAsUnanswered200JSONResponse{
		Success: true,
	}
}

func (p *markClarificationAsUnansweredPresenter) Error(err error) api.MarkClarificationAsUnansweredResponseObject {
	return httperror.From(err).Response()
}
