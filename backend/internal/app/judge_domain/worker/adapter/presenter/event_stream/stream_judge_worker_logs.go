package espresenter

import (
	"context"
	"judge-backend/api"
	"judge-backend/internal/app/judge_domain/events/judgeworkerlogadded"
	"judge-backend/internal/app/judge_domain/worker/adapter/presenter/types"
	streamjudgeworkerlogs "judge-backend/internal/app/judge_domain/worker/application/usecase/stream_judge_worker_logs"
	"judge-backend/internal/app/shared/adapter/presenter/sse"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/httperror"
)

type StreamJudgeWorkerLogsPresenter types.EsPresenter[streamjudgeworkerlogs.Output, api.StreamJudgeWorkerLogsResponseObject]
type streamJudgeWorkerLogsPresenter struct {
	eventBus sharedtypes.EventBus
}

func NewStreamJudgeWorkerLogsPresenter(
	eventBus sharedtypes.EventBus,

) StreamJudgeWorkerLogsPresenter {
	return &streamJudgeWorkerLogsPresenter{
		eventBus: eventBus,
	}
}

func (p *streamJudgeWorkerLogsPresenter) Output(ctx context.Context, output *streamjudgeworkerlogs.Output) api.StreamJudgeWorkerLogsResponseObject {
	body := sse.Stream(ctx, p.eventBus,
		func(w *sse.Writer) error {
			for _, log := range output.InitLog() {
				if err := w.WriteData(map[string]interface{}{"log": log.Text, "timestamp": log.Time}); err != nil {
					return err
				}
			}
			return nil
		},
		func(w *sse.Writer, event *judgeworkerlogadded.JudgeWorkerLogAddedEvent) error {
			if output.ID() != event.ID() {
				return nil
			}
			return w.WriteData(map[string]interface{}{"log": event.Log(), "timestamp": event.Timestamp()})
		},
	)

	return api.StreamJudgeWorkerLogs200TexteventStreamResponse{
		Body: body,
		Headers: api.StreamJudgeWorkerLogs200ResponseHeaders{
			CacheControl:     "no-cache",
			Connection:       "keep-alive",
			TransferEncoding: "chunked",
		},
	}
}

func (p *streamJudgeWorkerLogsPresenter) Error(err error) api.StreamJudgeWorkerLogsResponseObject {
	return httperror.From(err).Response()
}
