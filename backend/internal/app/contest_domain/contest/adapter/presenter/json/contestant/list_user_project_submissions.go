package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/contest/adapter/presenter/types"
	listuserprojectsubmissions "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/list_user_project_submissions"
	"judge-backend/internal/app/shared/types/httperror"
)

type ListUserProjectSubmissionsPresenter types.Presenter[listuserprojectsubmissions.Output, api.ListUserProjectSubmissionsResponseObject]

type listUserProjectSubmissionsPresenter struct{}

func NewListUserProjectSubmissionsPresenter() ListUserProjectSubmissionsPresenter {
	return &listUserProjectSubmissionsPresenter{}
}

func (p *listUserProjectSubmissionsPresenter) Output(output *listuserprojectsubmissions.Output) api.ListUserProjectSubmissionsResponseObject {
	submissions := make([]api.UserUserSpecSubmission, 0, len(output.Submissions))
	for _, submission := range output.Submissions {
		submissions = append(submissions, api.UserUserSpecSubmission{
			Id:         submission.ID,
			Status:     api.UserUserSpecSubmissionStatus(submission.Status),
			IsPartial:  submission.IsPartial,
			Score:      submission.Score,
			TotalScore: submission.TotalScore,
			Reason:     submission.Reason,
			StaffNote:  submission.StaffNote,
			CreatedAt:  submission.CreatedAt,
		})
	}

	return api.ListUserProjectSubmissions200JSONResponse{
		Success: true,
		Data:    submissions,
	}
}

func (p *listUserProjectSubmissionsPresenter) Error(err error) api.ListUserProjectSubmissionsResponseObject {
	return httperror.From(err).Response()
}
