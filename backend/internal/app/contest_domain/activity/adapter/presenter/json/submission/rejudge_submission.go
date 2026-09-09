package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/shared/types/httperror"
	"judge-backend/internal/app/contest_domain/activity/adapter/presenter/types"
	rejudgesubmission "judge-backend/internal/app/contest_domain/activity/application/usecase/submission/rejudge_submission"
)

type RejudgeSubmissionPresenter types.Presenter[rejudgesubmission.Output, api.RejudgeSubmissionResponseObject]

type rejudgeSubmissionPresenter struct {
}

func NewRejudgeSubmissionPresenter() RejudgeSubmissionPresenter {
	return &rejudgeSubmissionPresenter{}
}

func (p *rejudgeSubmissionPresenter) Output(output *rejudgesubmission.Output) api.RejudgeSubmissionResponseObject {
	return api.RejudgeSubmission200JSONResponse{
		Success: true,
	}
}

func (p *rejudgeSubmissionPresenter) Error(err error) api.RejudgeSubmissionResponseObject {
	return httperror.From(err).Response()
}
