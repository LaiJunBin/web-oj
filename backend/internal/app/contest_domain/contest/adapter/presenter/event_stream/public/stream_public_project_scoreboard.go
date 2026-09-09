package espresenter

import (
	"context"
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/contest/adapter/presenter/types"
	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	streampublicprojectscoreboard "judge-backend/internal/app/contest_domain/contest/application/usecase/public/project/stream_public_project_scoreboard"
	"judge-backend/internal/app/contest_domain/events/scoreboardupdated"
	"judge-backend/internal/app/shared/adapter/presenter/sse"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/httperror"
)

type StreamPublicProjectScoreboardPresenter types.EsPresenter[streampublicprojectscoreboard.Output, api.StreamPublicProjectScoreboardResponseObject]
type streamPublicProjectScoreboardPresenter struct {
	eventBus sharedtypes.EventBus
}

func NewStreamPublicProjectScoreboardPresenter(
	eventBus sharedtypes.EventBus,
) StreamPublicProjectScoreboardPresenter {
	return &streamPublicProjectScoreboardPresenter{
		eventBus: eventBus,
	}
}

func (p *streamPublicProjectScoreboardPresenter) Output(ctx context.Context, output *streampublicprojectscoreboard.Output) api.StreamPublicProjectScoreboardResponseObject {
	body := sse.Stream(ctx, p.eventBus, nil, func(w *sse.Writer, event *scoreboardupdated.Event[*dto.Scoreboard]) error {
		if event.ContestProjectID() != output.ContestProjectID {
			return nil
		}
		return w.WriteData(event.Data())
	})

	return api.StreamPublicProjectScoreboard200TexteventStreamResponse{
		Body: body,
		Headers: api.StreamPublicProjectScoreboard200ResponseHeaders{
			CacheControl:     "no-cache",
			Connection:       "keep-alive",
			TransferEncoding: "chunked",
		},
	}
}

func (p *streamPublicProjectScoreboardPresenter) Error(err error) api.StreamPublicProjectScoreboardResponseObject {
	return httperror.From(err).Response()
}
