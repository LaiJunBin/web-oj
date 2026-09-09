package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/contest/adapter/presenter/types"
	getusersubmission "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/get_user_submission"
	"judge-backend/internal/app/shared/types/httperror"
)

type GetUserSubmissionPresenter types.Presenter[getusersubmission.Output, api.GetUserSubmissionResponseObject]

type getUserSubmissionContestPresenter struct{}

func NewGetUserSubmissionPresenter() GetUserSubmissionPresenter {
	return &getUserSubmissionContestPresenter{}
}

func (p *getUserSubmissionContestPresenter) Output(output *getusersubmission.Output) api.GetUserSubmissionResponseObject {
	gradingItems := make([]api.UserUserSpecSubmissionGradingItem, 0, len(output.Submission.GradingItems))
	for _, gradingItem := range output.Submission.GradingItems {
		testcases := make([]api.UserUserSpecSubmissionTestcase, 0, len(gradingItem.Testcases))
		for _, testcase := range gradingItem.Testcases {
			testcases = append(testcases, api.UserUserSpecSubmissionTestcase{
				Id:       testcase.ID,
				Status:   testcase.Status,
				ExitCode: testcase.ExitCode,
				ExecTime: testcase.ExecTime,
				Stdout:   testcase.Stdout,
				Stderr:   testcase.Stderr,
			})
		}

		gradingItems = append(gradingItems, api.UserUserSpecSubmissionGradingItem{
			Id:        gradingItem.ID,
			Status:    api.SharedSubmissionGradingItemStatus(gradingItem.Status),
			Score:     gradingItem.Score,
			Testcases: testcases,
		})
	}

	return api.GetUserSubmission200JSONResponse{
		Success: true,
		Data: api.UserUserSpecSubmissionDetail{
			Id:           output.Submission.ID,
			Status:       api.UserUserSpecSubmissionStatus(output.Submission.Status),
			IsPartial:    output.Submission.IsPartial,
			Score:        output.Submission.Score,
			TotalScore:   output.Submission.TotalScore,
			Reason:       output.Submission.Reason,
			StaffNote:    output.Submission.StaffNote,
			GradingItems: gradingItems,
			CreatedAt:    output.Submission.CreatedAt,
		},
	}
}

func (p *getUserSubmissionContestPresenter) Error(err error) api.GetUserSubmissionResponseObject {
	return httperror.From(err).Response()
}
