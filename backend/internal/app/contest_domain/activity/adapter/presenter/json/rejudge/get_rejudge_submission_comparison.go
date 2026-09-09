package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/activity/adapter/presenter/types"
	"judge-backend/internal/app/contest_domain/activity/application/query/dto"
	getrejudgesubmissioncomparison "judge-backend/internal/app/contest_domain/activity/application/usecase/rejudge/get_rejudge_submission_comparison"
	"judge-backend/internal/app/shared/types/httperror"
)

type GetRejudgeSubmissionComparisonPresenter types.Presenter[getrejudgesubmissioncomparison.Output, api.GetRejudgeSubmissionComparisonResponseObject]

type getRejudgeSubmissionComparisonPresenter struct {
}

func NewGetRejudgeSubmissionComparisonPresenter() GetRejudgeSubmissionComparisonPresenter {
	return &getRejudgeSubmissionComparisonPresenter{}
}

func transformJudgmentOutcomeDetail(detail *dto.JudgmentOutcomeDetail) api.AdminJudgmentSpecJudgmentOutcomeDetail {
	gradingItems := make([]api.AdminJudgmentSpecGradingItemOutcome, 0, len(detail.GradingItems))
	for _, gradingItem := range detail.GradingItems {
		testcases := make([]api.AdminJudgmentSpecTestcaseOutcome, 0, len(gradingItem.Testcases))
		for _, testcase := range gradingItem.Testcases {
			testcases = append(testcases, api.AdminJudgmentSpecTestcaseOutcome{
				Id:          testcase.ID,
				Name:        testcase.Name,
				Status:      testcase.Status,
				Score:       testcase.Score,
				TotalScore:  testcase.TotalScore,
				ParsedScore: testcase.ParsedScore,
			})
		}

		gradingItems = append(gradingItems, api.AdminJudgmentSpecGradingItemOutcome{
			Id:              gradingItem.ID,
			Label:           gradingItem.Label,
			Name:            gradingItem.Name,
			Status:          gradingItem.Status,
			Score:           gradingItem.Score,
			TotalScore:      gradingItem.TotalScore,
			ScoringStrategy: gradingItem.ScoringStrategy,
			Testcases:       testcases,
		})
	}

	return api.AdminJudgmentSpecJudgmentOutcomeDetail{
		JudgmentId:   detail.JudgmentID,
		Status:       detail.Status,
		Score:        detail.Score,
		TotalScore:   detail.TotalScore,
		GradingItems: gradingItems,
	}
}

func (p *getRejudgeSubmissionComparisonPresenter) Output(output *getrejudgesubmissioncomparison.Output) api.GetRejudgeSubmissionComparisonResponseObject {
	comparison := output.Comparison

	// Left nil when the submission had no adopted judgment before the rejudge, so
	// the response carries a null the UI can label "no previous judgment" instead
	// of a zero-scored object it would render as a real score of 0.
	var before *api.AdminJudgmentSpecJudgmentOutcomeDetail
	if comparison.Before != nil {
		detail := transformJudgmentOutcomeDetail(comparison.Before)
		before = &detail
	}

	return api.GetRejudgeSubmissionComparison200JSONResponse{
		Success: true,
		Data: api.AdminJudgmentSpecSubmissionComparisonDetail{
			Id:     comparison.ID,
			Before: before,
			After:  transformJudgmentOutcomeDetail(comparison.After),
		},
	}
}

func (p *getRejudgeSubmissionComparisonPresenter) Error(err error) api.GetRejudgeSubmissionComparisonResponseObject {
	return httperror.From(err).Response()
}
