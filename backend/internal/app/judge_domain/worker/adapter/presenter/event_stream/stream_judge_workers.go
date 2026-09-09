package espresenter

import (
	"context"

	"judge-backend/api"
	"judge-backend/internal/app/judge_domain/events/judgeworkerstatuschanged"
	jsonpresenter "judge-backend/internal/app/judge_domain/worker/adapter/presenter/json"
	"judge-backend/internal/app/judge_domain/worker/adapter/presenter/types"
	listjudgeworker "judge-backend/internal/app/judge_domain/worker/application/usecase/list_judge_worker"
	"judge-backend/internal/app/shared/adapter/presenter/sse"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/httperror"
)

type StreamJudgeWorkersPresenter types.EsPresenter[struct{}, api.StreamJudgeWorkersResponseObject]
type streamJudgeWorkersPresenter struct {
	usecase   listjudgeworker.Usecase
	presenter jsonpresenter.ListJudgeWorkerPresenter
	eventBus  sharedtypes.EventBus
}

func NewStreamJudgeWorkersPresenter(
	usecase listjudgeworker.Usecase,
	eventBus sharedtypes.EventBus,
) StreamJudgeWorkersPresenter {
	return &streamJudgeWorkersPresenter{
		usecase:   usecase,
		presenter: jsonpresenter.NewListJudgeWorkerPresenter(),
		eventBus:  eventBus,
	}
}

func (p *streamJudgeWorkersPresenter) Output(ctx context.Context, output *struct{}) api.StreamJudgeWorkersResponseObject {
	// Each status-change event is only a trigger; the payload sent to the client
	// is always a fresh snapshot of the whole worker list.
	body := sse.Stream(ctx, p.eventBus, nil, func(w *sse.Writer, _ *judgeworkerstatuschanged.JudgeWorkerStatusChangedEvent) error {
		judgeWorkers, err := p.usecase.Execute(&listjudgeworker.Input{})
		if err != nil {
			return err
		}
		return w.WriteData(p.presenter.Output(judgeWorkers))
	})

	return api.StreamJudgeWorkers200TexteventStreamResponse{
		Body: body,
		Headers: api.StreamJudgeWorkers200ResponseHeaders{
			CacheControl:     "no-cache",
			Connection:       "keep-alive",
			TransferEncoding: "chunked",
		},
	}
}

func (p *streamJudgeWorkersPresenter) Error(err error) api.StreamJudgeWorkersResponseObject {
	return httperror.From(err).Response()
}
