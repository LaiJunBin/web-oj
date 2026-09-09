package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/shared/types/httperror"
	"judge-backend/internal/app/contest_domain/activity/adapter/presenter/types"
	updateclarification "judge-backend/internal/app/contest_domain/activity/application/usecase/clarification/update_clarification"
)

type MarkClarificationAsAnsweredPresenter types.Presenter[updateclarification.Output, api.MarkClarificationAsAnsweredResponseObject]

type markClarificationAsAnsweredPresenter struct{}

func NewMarkClarificationAsAnsweredPresenter() MarkClarificationAsAnsweredPresenter {
	return &markClarificationAsAnsweredPresenter{}
}

func (p *markClarificationAsAnsweredPresenter) Output(output *updateclarification.Output) api.MarkClarificationAsAnsweredResponseObject {
	return api.MarkClarificationAsAnswered200JSONResponse{
		Success: true,
	}
}

func (p *markClarificationAsAnsweredPresenter) Error(err error) api.MarkClarificationAsAnsweredResponseObject {
	return httperror.From(err).Response()
}
