package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/shared/types/httperror"
	"judge-backend/internal/app/contest_domain/activity/adapter/presenter/types"
	canceljudge "judge-backend/internal/app/contest_domain/activity/application/usecase/judgment/cancel_judge"
)

type CancelJudgePresenter types.Presenter[canceljudge.Output, api.CancelJudgeResponseObject]

type cancelJudgePresenter struct {
}

func NewCancelJudgePresenter() CancelJudgePresenter {
	return &cancelJudgePresenter{}
}

func (p *cancelJudgePresenter) Output(output *canceljudge.Output) api.CancelJudgeResponseObject {
	return api.CancelJudge200JSONResponse{
		Success: true,
	}
}

func (p *cancelJudgePresenter) Error(err error) api.CancelJudgeResponseObject {
	return httperror.From(err).Response()
}
