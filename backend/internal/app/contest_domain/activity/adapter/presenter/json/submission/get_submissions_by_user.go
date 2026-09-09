package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/shared/types/httperror"
	"judge-backend/internal/app/contest_domain/activity/adapter/presenter/types"
	getsubmissionsbyuserid "judge-backend/internal/app/contest_domain/activity/application/usecase/submission/get_submissions_by_user_id"
)

type GetSubmissionsByUserPresenter types.Presenter[getsubmissionsbyuserid.Output, api.GetSubmissionsByUserResponseObject]

type getSubmissionsByUserPresenter struct {
}

func NewGetSubmissionsByUserPresenter() GetSubmissionsByUserPresenter {
	return &getSubmissionsByUserPresenter{}
}

func (p *getSubmissionsByUserPresenter) Output(output *getsubmissionsbyuserid.Output) api.GetSubmissionsByUserResponseObject {
	submissions := []api.AdminSubmissionSpecSubmission{}
	for _, submission := range output.Submissions {
		submissions = append(submissions, api.AdminSubmissionSpecSubmission{
			Id:         submission.ID,
			Score:      submission.Score,
			TotalScore: submission.TotalScore,
			Ignored:    submission.Ignored,
			Contest: api.AdminSubmissionSpecContest{
				Id:   submission.Contest.ID,
				Name: submission.Contest.Name,
			},
			Project: api.AdminSubmissionSpecProjectInfo{
				Id:   submission.Project.ID,
				Name: submission.Project.Name,
			},
			User: api.AdminSubmissionSpecUser{
				Id:       submission.User.ID,
				Username: submission.User.Username,
				Name:     submission.User.Name,
			},
			IsPartial: submission.IsPartial,
			Status:    api.SharedSubmissionStatus(submission.Status),
			Type:      api.SharedSubmissionType(submission.Type),
			CreatedAt: submission.Time,
		})
	}

	return api.GetSubmissionsByUser200JSONResponse{
		Success: true,
		Data:    submissions,
	}
}

func (p *getSubmissionsByUserPresenter) Error(err error) api.GetSubmissionsByUserResponseObject {
	return httperror.From(err).Response()
}
