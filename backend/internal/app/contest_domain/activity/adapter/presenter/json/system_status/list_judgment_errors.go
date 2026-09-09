package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/activity/adapter/presenter/types"
	"judge-backend/internal/app/contest_domain/activity/application/query/dto"
	listjudgmenterrors "judge-backend/internal/app/contest_domain/activity/application/usecase/system_status/list_judgment_errors"
	"judge-backend/internal/app/shared/types/httperror"
)

// JudgmentErrorList renders the list. Shared with the SSE presenter, which
// pushes the identical payload — a stream that drifted from the polled response
// would be a trap for a page that uses both.
func JudgmentErrorList(infos []*dto.JudgmentError) []api.AdminJudgmentSpecJudgmentError {
	errors := []api.AdminJudgmentSpecJudgmentError{}
	for _, judgmentError := range infos {
		errors = append(errors, api.AdminJudgmentSpecJudgmentError{
			JudgmentId:       judgmentError.JudgmentID,
			SubmissionId:     judgmentError.SubmissionID,
			ContestProjectId: judgmentError.ContestProjectID,
			ErrorMessage:     judgmentError.ErrorMessage,
			CreatedAt:        judgmentError.CreatedAt,
		})
	}
	return errors
}

type ListJudgmentErrorsPresenter types.Presenter[listjudgmenterrors.Output, api.ListJudgmentErrorsResponseObject]

type listJudgmentErrorsPresenter struct {
}

func NewListJudgmentErrorsPresenter() ListJudgmentErrorsPresenter {
	return &listJudgmentErrorsPresenter{}
}

func (p *listJudgmentErrorsPresenter) Output(output *listjudgmenterrors.Output) api.ListJudgmentErrorsResponseObject {
	return api.ListJudgmentErrors200JSONResponse{
		Success: true,
		Data:    JudgmentErrorList(output.Errors),
	}
}

func (p *listJudgmentErrorsPresenter) Error(err error) api.ListJudgmentErrorsResponseObject {
	return httperror.From(err).Response()
}
