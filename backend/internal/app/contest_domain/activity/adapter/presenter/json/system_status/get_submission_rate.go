package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/activity/adapter/presenter/types"
	"judge-backend/internal/app/contest_domain/activity/application/query/dto"
	getsubmissionrate "judge-backend/internal/app/contest_domain/activity/application/usecase/system_status/get_submission_rate"
	"judge-backend/internal/app/shared/types/httperror"
)

// SubmissionRateSeries renders the series. Shared with the SSE presenter, which
// pushes the identical payload — a stream that drifted from the polled response
// would be a trap for a page that uses both.
func SubmissionRateSeries(infos []*dto.SubmissionRatePoint) []api.AdminSubmissionSpecSubmissionRatePoint {
	points := []api.AdminSubmissionSpecSubmissionRatePoint{}
	for _, point := range infos {
		points = append(points, api.AdminSubmissionSpecSubmissionRatePoint{
			Minute: point.Minute,
			Count:  point.Count,
		})
	}
	return points
}

type GetSubmissionRatePresenter types.Presenter[getsubmissionrate.Output, api.GetSubmissionRateResponseObject]

type getSubmissionRatePresenter struct {
}

func NewGetSubmissionRatePresenter() GetSubmissionRatePresenter {
	return &getSubmissionRatePresenter{}
}

func (p *getSubmissionRatePresenter) Output(output *getsubmissionrate.Output) api.GetSubmissionRateResponseObject {
	return api.GetSubmissionRate200JSONResponse{
		Success: true,
		Data:    SubmissionRateSeries(output.Points),
	}
}

func (p *getSubmissionRatePresenter) Error(err error) api.GetSubmissionRateResponseObject {
	return httperror.From(err).Response()
}
