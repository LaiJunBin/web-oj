package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/shared/types/httperror"
	"judge-backend/internal/app/contest_domain/activity/adapter/presenter/types"
	approverejudge "judge-backend/internal/app/contest_domain/activity/application/usecase/rejudge/approve_rejudge"
)

type ApproveRejudgePresenter types.Presenter[approverejudge.Output, api.ApproveRejudgeResponseObject]

type approveRejudgePresenter struct {
}

func NewApproveRejudgePresenter() ApproveRejudgePresenter {
	return &approveRejudgePresenter{}
}

func (p *approveRejudgePresenter) Output(output *approverejudge.Output) api.ApproveRejudgeResponseObject {
	return api.ApproveRejudge200JSONResponse{
		Success: true,
	}
}

func (p *approveRejudgePresenter) Error(err error) api.ApproveRejudgeResponseObject {
	return httperror.From(err).Response()
}
