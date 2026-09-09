package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/activity/adapter/presenter/types"
	getjudgmentanalytics "judge-backend/internal/app/contest_domain/activity/application/usecase/contest_project/get_judgment_analytics"
	"judge-backend/internal/app/shared/types/httperror"
)

type GetJudgmentAnalyticsPresenter types.Presenter[getjudgmentanalytics.Output, api.GetContestProjectJudgmentAnalyticsResponseObject]

type getJudgmentAnalyticsPresenter struct{}

func NewGetJudgmentAnalyticsPresenter() GetJudgmentAnalyticsPresenter {
	return &getJudgmentAnalyticsPresenter{}
}

func (p *getJudgmentAnalyticsPresenter) Output(output *getjudgmentanalytics.Output) api.GetContestProjectJudgmentAnalyticsResponseObject {
	a := output.Analytics

	statusCounts := make([]api.AdminContestProjectSpecJudgmentStatusCount, 0, len(a.StatusCounts))
	for _, sc := range a.StatusCounts {
		statusCounts = append(statusCounts, api.AdminContestProjectSpecJudgmentStatusCount{
			Status: sc.Status,
			Count:  sc.Count,
		})
	}

	return api.GetContestProjectJudgmentAnalytics200JSONResponse{
		Success: true,
		Data: api.AdminContestProjectSpecJudgmentAnalytics{
			TotalSubmissions:     a.TotalSubmissions,
			TotalJudgments:       a.TotalJudgments,
			StatusCounts:         statusCounts,
			AvgJudgingDurationMs: a.AvgJudgingDurationMs,
			AvgQueueWaitMs:       a.AvgQueueWaitMs,
		},
	}
}

func (p *getJudgmentAnalyticsPresenter) Error(err error) api.GetContestProjectJudgmentAnalyticsResponseObject {
	return httperror.From(err).Response()
}
