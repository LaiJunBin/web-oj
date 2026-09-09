package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/shared/types/httperror"
	"judge-backend/internal/app/contest_domain/activity/adapter/presenter/types"
	getsubmissionsbycontestproject "judge-backend/internal/app/contest_domain/activity/application/usecase/submission/get_submissions_by_contest_project"
)

type GetSubmissionsByContestProjectPresenter types.Presenter[getsubmissionsbycontestproject.Output, api.GetSubmissionsByContestProjectResponseObject]

type getSubmissionsByContestProjectPresenter struct {
}

func NewGetSubmissionsByContestProjectPresenter() GetSubmissionsByContestProjectPresenter {
	return &getSubmissionsByContestProjectPresenter{}
}

func (p *getSubmissionsByContestProjectPresenter) Output(output *getsubmissionsbycontestproject.Output) api.GetSubmissionsByContestProjectResponseObject {
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

	return api.GetSubmissionsByContestProject200JSONResponse{
		Success: true,
		Data:    submissions,
	}
}

func (p *getSubmissionsByContestProjectPresenter) Error(err error) api.GetSubmissionsByContestProjectResponseObject {
	return httperror.From(err).Response()
}
