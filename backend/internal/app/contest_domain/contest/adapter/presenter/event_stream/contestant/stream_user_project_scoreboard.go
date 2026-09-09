package espresenter

import (
	"context"
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/contest/adapter/presenter/types"
	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	streamuserprojectscoreboard "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/stream_user_project_scoreboard"
	"judge-backend/internal/app/contest_domain/events/scoreboardupdated"
	"judge-backend/internal/app/shared/adapter/presenter/sse"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/httperror"
)

type StreamUserProjectScoreboardPresenter types.EsPresenter[streamuserprojectscoreboard.Output, api.StreamUserProjectScoreboardResponseObject]
type streamUserProjectScoreboardPresenter struct {
	eventBus sharedtypes.EventBus
}

func NewStreamUserProjectScoreboardPresenter(
	eventBus sharedtypes.EventBus,
) StreamUserProjectScoreboardPresenter {
	return &streamUserProjectScoreboardPresenter{
		eventBus: eventBus,
	}
}

func (p *streamUserProjectScoreboardPresenter) Output(ctx context.Context, output *streamuserprojectscoreboard.Output) api.StreamUserProjectScoreboardResponseObject {
	body := sse.Stream(ctx, p.eventBus, nil, func(w *sse.Writer, event *scoreboardupdated.Event[*dto.Scoreboard]) error {
		if event.ContestProjectID() != output.ContestProjectID {
			return nil
		}
		return w.WriteData(event.Data())
	})

	return api.StreamUserProjectScoreboard200TexteventStreamResponse{
		Body: body,
		Headers: api.StreamUserProjectScoreboard200ResponseHeaders{
			CacheControl:     "no-cache",
			Connection:       "keep-alive",
			TransferEncoding: "chunked",
		},
	}
}

func (p *streamUserProjectScoreboardPresenter) Error(err error) api.StreamUserProjectScoreboardResponseObject {
	return httperror.From(err).Response()
}
