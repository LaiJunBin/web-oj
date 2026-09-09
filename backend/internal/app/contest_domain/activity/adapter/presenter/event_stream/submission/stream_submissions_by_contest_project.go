package espresenter

import (
	"context"
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/activity/application/query"
	streamsubmissionsbycontestproject "judge-backend/internal/app/contest_domain/activity/application/usecase/submission/stream_submissions_by_contest_project"
	"judge-backend/internal/app/contest_domain/events/submissionupdated"
	"judge-backend/internal/app/shared/adapter/presenter/sse"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/httperror"
	"time"
)

type StreamSubmissionsByContestProjectPresenter interface {
	Output(ctx context.Context, output *streamsubmissionsbycontestproject.Output) api.StreamSubmissionsByContestProjectResponseObject
	Error(err error) api.StreamSubmissionsByContestProjectResponseObject
}

type streamSubmissionsByContestProjectPresenter struct {
	eventBus sharedtypes.EventBus
	query    query.SubmissionQuery
}

func NewStreamSubmissionsByContestProjectPresenter(
	eventBus sharedtypes.EventBus,
	query query.SubmissionQuery,
) StreamSubmissionsByContestProjectPresenter {
	return &streamSubmissionsByContestProjectPresenter{
		eventBus: eventBus,
		query:    query,
	}
}

type SubmissionUser struct {
	ID       int32  `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
}

type SubmissionContest struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type SubmissionProject struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type SubmissionUpdate struct {
	ID         string            `json:"id"`
	User       SubmissionUser    `json:"user"`
	Contest    SubmissionContest `json:"contest"`
	Project    SubmissionProject `json:"project"`
	IsPartial  bool              `json:"isPartial"`
	Ignored    bool              `json:"ignored"`
	Status     string            `json:"status"`
	Type       string            `json:"type"`
	Score      float64           `json:"score"`
	TotalScore float64           `json:"totalScore"`
	CreatedAt  time.Time         `json:"createdAt"`
}

func (p *streamSubmissionsByContestProjectPresenter) Output(ctx context.Context, output *streamsubmissionsbycontestproject.Output) api.StreamSubmissionsByContestProjectResponseObject {
	body := sse.Stream(ctx, p.eventBus, nil, func(w *sse.Writer, e *submissionupdated.Event) error {
		if e.ContestProjectID() != output.ContestProjectID {
			return nil
		}

		sub, err := p.query.GetSubmissionInfoByID(e.Judgment().SubmissionID)
		if err != nil {
			return nil
		}

		return w.WriteData(SubmissionUpdate{
			ID:         sub.ID,
			User:       SubmissionUser{ID: sub.User.ID, Username: sub.User.Username, Name: sub.User.Name},
			Contest:    SubmissionContest{ID: sub.Contest.ID, Name: sub.Contest.Name},
			Project:    SubmissionProject{ID: sub.Project.ID, Name: sub.Project.Name},
			IsPartial:  sub.IsPartial,
			Ignored:    sub.Ignored,
			Status:     sub.Status,
			Type:       sub.Type,
			Score:      sub.Score,
			TotalScore: sub.TotalScore,
			CreatedAt:  sub.Time,
		})
	})

	return api.StreamSubmissionsByContestProject200TexteventStreamResponse{
		Body: body,
		Headers: api.StreamSubmissionsByContestProject200ResponseHeaders{
			CacheControl:     "no-cache",
			Connection:       "keep-alive",
			TransferEncoding: "chunked",
		},
	}
}

func (p *streamSubmissionsByContestProjectPresenter) Error(err error) api.StreamSubmissionsByContestProjectResponseObject {
	return httperror.From(err).Response()
}
