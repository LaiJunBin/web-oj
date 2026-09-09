package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/shared/types/httperror"
	"judge-backend/internal/app/contest_domain/activity/adapter/presenter/types"
	rejectrejudge "judge-backend/internal/app/contest_domain/activity/application/usecase/rejudge/reject_rejudge"
)

type RejectRejudgePresenter types.Presenter[rejectrejudge.Output, api.RejectRejudgeResponseObject]

type rejectRejudgePresenter struct {
}

func NewRejectRejudgePresenter() RejectRejudgePresenter {
	return &rejectRejudgePresenter{}
}

func (p *rejectRejudgePresenter) Output(output *rejectrejudge.Output) api.RejectRejudgeResponseObject {
	return api.RejectRejudge200JSONResponse{
		Success: true,
	}
}

func (p *rejectRejudgePresenter) Error(err error) api.RejectRejudgeResponseObject {
	return httperror.From(err).Response()
}
