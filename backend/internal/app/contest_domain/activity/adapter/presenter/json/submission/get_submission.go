package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/activity/adapter/presenter/types"
	"judge-backend/internal/app/contest_domain/activity/application/query/dto"
	getsubmission "judge-backend/internal/app/contest_domain/activity/application/usecase/submission/get_submission"
	"judge-backend/internal/app/shared/types/httperror"
	"strings"
)

type GetSubmissionPresenter types.Presenter[getsubmission.Output, api.GetSubmissionResponseObject]

type getSubmissionPresenter struct {
}

func NewGetSubmissionPresenter() GetSubmissionPresenter {
	return &getSubmissionPresenter{}
}

func transformLifecycle(lifecycles []*dto.Lifecycle) []api.AdminJudgmentSpecLifecycle {
	result := make([]api.AdminJudgmentSpecLifecycle, 0)
	for _, lifecycle := range lifecycles {
		result = append(result, api.AdminJudgmentSpecLifecycle{
			Id:                   lifecycle.ID,
			Command:              lifecycle.Command,
			ExecutionEnvironment: lifecycle.ExecutionEnvironment,
			Result: api.AdminJudgmentSpecResult{
				ExitCode: lifecycle.Result.ExitCode,
				ExecTime: lifecycle.Result.ExecTime,
				Stdout:   lifecycle.Result.Stdout,
				Stderr:   lifecycle.Result.Stderr,
			},
		})
	}
	return result
}

func (p *getSubmissionPresenter) Output(output *getsubmission.Output) api.GetSubmissionResponseObject {
	gradingItems := make([]api.AdminSubmissionSpecGradingItemInfo, 0)
	for _, gradingItem := range output.Submission.GradingItems {
		testcases := make([]api.AdminSubmissionSpecTestcaseInfo, 0)
		for _, testcase := range gradingItem.Testcases {
			testcases = append(testcases, api.AdminSubmissionSpecTestcaseInfo{
				Id:    testcase.ID,
				Name:  testcase.Name,
				Score: testcase.Score,
			})
		}

		gradingItems = append(gradingItems, api.AdminSubmissionSpecGradingItemInfo{
			Id:        gradingItem.ID,
			Name:      gradingItem.Name,
			Score:     gradingItem.Score,
			Testcases: testcases,
		})
	}

	judgments := make([]api.AdminJudgmentSpecJudgment, 0)
	for _, judgment := range output.Submission.Judgments {
		gradingItems := make([]api.AdminJudgmentSpecGradingItem, 0)
		for _, gradingItem := range judgment.GradingItems {
			testcases := make([]api.AdminJudgmentSpecTestcase, 0)
			for _, testcase := range gradingItem.Testcases {
				testcases = append(testcases, api.AdminJudgmentSpecTestcase{
					Id:          testcase.ID,
					Name:        testcase.Name,
					Score:       testcase.Score,
					TotalScore:  testcase.TotalScore,
					ParsedScore: testcase.ParsedScore,
					Status:      testcase.Status,
					Command:     testcase.Command,
					Result: api.AdminJudgmentSpecResult{
						ExitCode: testcase.Result.ExitCode,
						ExecTime: testcase.Result.ExecTime,
						Stdout:   testcase.Result.Stdout,
						Stderr:   testcase.Result.Stderr,
					},
					Lifecycle: api.AdminJudgmentSpecTestcaseLifecycle{
						BeforeEach: transformLifecycle(testcase.Lifecycle.BeforeEach),
						AfterEach:  transformLifecycle(testcase.Lifecycle.AfterEach),
					},
				})
			}

			gradingItems = append(gradingItems, api.AdminJudgmentSpecGradingItem{
				Id:              gradingItem.ID,
				Label:           gradingItem.Label,
				Name:            gradingItem.Name,
				Score:           gradingItem.Score,
				TotalScore:      gradingItem.TotalScore,
				Status:          gradingItem.Status,
				ScoringStrategy: gradingItem.ScoringStrategy,
				ResultParser:    gradingItem.ResultParser,
				Testcases:       testcases,
			})
		}
		judgments = append(judgments, api.AdminJudgmentSpecJudgment{
			Id:                judgment.ID,
			JudgeType:         judgment.JudgeType,
			Status:            judgment.Status,
			Score:             judgment.Score,
			TotalScore:        judgment.TotalScore,
			RejudgeId:         judgment.RejudgeID,
			JudgingStartedAt:  judgment.JudgingStartedAt,
			JudgingFinishedAt: judgment.JudgingFinishedAt,
			ErrorMessage:      judgment.ErrorMessage,
			StaffNote:         judgment.StaffNote,
			Project: api.AdminJudgmentSpecProject{
				Id:           judgment.Project.ID,
				ImageTag:     judgment.Project.ImageTag,
				TestImageTag: judgment.Project.TestImageTag,
				Env:          strings.Join(judgment.Project.Env, "\n"),
				TestEnv:      strings.Join(judgment.Project.TestEnv, "\n"),
				MountPath:    judgment.Project.MountPath,
				ResultParser: judgment.Project.ResultParser,
			},
			CreatedAt:    judgment.CreatedAt,
			GradingItems: gradingItems,
			JudgmentLifecycle: api.AdminJudgmentSpecJudgmentLifeycle{
				BeforeAll: transformLifecycle(judgment.Lifecycle.BeforeAll),
				AfterAll:  transformLifecycle(judgment.Lifecycle.AfterAll),
			},
		})
	}

	return api.GetSubmission200JSONResponse{
		Success: true,
		Data: api.AdminSubmissionSpecSubmissionDetail{
			Id:        output.Submission.ID,
			IsPartial: output.Submission.IsPartial,
			Ignored:   output.Submission.Ignored,
			CreatedAt: output.Submission.CreatedAt,
			Contest: api.AdminSubmissionSpecContest{
				Id:   output.Submission.Contest.ID,
				Name: output.Submission.Contest.Name,
			},
			Project: api.AdminSubmissionSpecProjectInfo{
				Id:   output.Submission.Project.ID,
				Name: output.Submission.Project.Name,
			},
			User: api.AdminSubmissionSpecUser{
				Id:       output.Submission.User.ID,
				Username: output.Submission.User.Username,
				Name:     output.Submission.User.Name,
			},
			GradingItems: gradingItems,
			Judgments:    judgments,
		},
	}
}

func (p *getSubmissionPresenter) Error(err error) api.GetSubmissionResponseObject {
	return httperror.From(err).Response()
}
