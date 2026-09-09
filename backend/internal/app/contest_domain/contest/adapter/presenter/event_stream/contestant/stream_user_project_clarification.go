package espresenter

import (
	"context"
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/contest/adapter/presenter/types"
	streamuserprojectclarification "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/stream_user_project_clarification"
	"judge-backend/internal/app/contest_domain/events/clarificationnotificationrequested"
	"judge-backend/internal/app/shared/adapter/presenter/sse"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/httperror"
)

type StreamUserProjectClarificationPresenter types.EsPresenter[streamuserprojectclarification.Output, api.StreamUserProjectClarificationResponseObject]
type streamUserProjectClarificationPresenter struct {
	eventBus sharedtypes.EventBus
}

func NewStreamUserProjectClarificationPresenter(
	eventBus sharedtypes.EventBus,
) StreamUserProjectClarificationPresenter {
	return &streamUserProjectClarificationPresenter{
		eventBus: eventBus,
	}
}

func (p *streamUserProjectClarificationPresenter) Output(ctx context.Context, output *streamuserprojectclarification.Output) api.StreamUserProjectClarificationResponseObject {
	body := sse.Stream(ctx, p.eventBus, nil, func(w *sse.Writer, event *clarificationnotificationrequested.Event) error {
		if !p.shouldSendClarification(event, output.UserID) {
			return nil
		}
		return w.WriteData(map[string]string{
			"subject": event.Clarification().Subject().String(),
			"content": event.Clarification().Content(),
		})
	})

	return api.StreamUserProjectClarification200TexteventStreamResponse{
		Body: body,
		Headers: api.StreamUserProjectClarification200ResponseHeaders{
			CacheControl:     "no-cache",
			Connection:       "keep-alive",
			TransferEncoding: "chunked",
		},
	}
}

func (p *streamUserProjectClarificationPresenter) Error(err error) api.StreamUserProjectClarificationResponseObject {
	return httperror.From(err).Response()
}

func (p *streamUserProjectClarificationPresenter) shouldSendClarification(event *clarificationnotificationrequested.Event, userID int32) bool {
	clarification := event.Clarification()
	if clarification.ReceiverUserID() != nil && *clarification.ReceiverUserID() != userID {
		if !clarification.SendToAll() {
			return false
		}
	}

	return true
}
