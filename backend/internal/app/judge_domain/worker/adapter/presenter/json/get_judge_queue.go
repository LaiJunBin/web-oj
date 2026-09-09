package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/judge_domain/worker/adapter/presenter/types"
	"judge-backend/internal/app/judge_domain/worker/application/query/dto"
	getjudgequeue "judge-backend/internal/app/judge_domain/worker/application/usecase/get_judge_queue"
	"judge-backend/internal/app/shared/types/httperror"
)

// Judge type keys as the pipeline names them internally. The published response
// spells the same three out as named fields rather than as a free-form map, so
// the client gets a compile-time shape instead of having to know these strings —
// which is also why the mapping has to live somewhere, and this is the boundary
// it belongs on.
const (
	judgeKey        = "JUDGE"
	judgePartialKey = "JUDGE_PARTIAL"
	rejudgeKey      = "REJUDGE"
)

// JudgeQueueStatus renders the status snapshot. Shared with the SSE presenter,
// which pushes the identical payload — a stream that drifted from the polled
// response would be a trap for a page that uses both.
func JudgeQueueStatus(status *dto.JudgeQueueStatus) api.AdminJudgeWorkerSpecJudgeQueueStatus {
	return api.AdminJudgeWorkerSpecJudgeQueueStatus{
		QueueDepths: api.AdminJudgeWorkerSpecJudgeQueueDepths{
			Judge:        status.QueueDepths[judgeKey],
			JudgePartial: status.QueueDepths[judgePartialKey],
			Rejudge:      status.QueueDepths[rejudgeKey],
		},
		ActiveJudgments:     status.ActiveJudgments,
		OldestQueuedSeconds: status.OldestQueuedSeconds,
	}
}

type GetJudgeQueuePresenter types.Presenter[getjudgequeue.Output, api.GetJudgeQueueResponseObject]

type getJudgeQueuePresenter struct {
}

func NewGetJudgeQueuePresenter() GetJudgeQueuePresenter {
	return &getJudgeQueuePresenter{}
}

func (p *getJudgeQueuePresenter) Output(output *getjudgequeue.Output) api.GetJudgeQueueResponseObject {
	return api.GetJudgeQueue200JSONResponse{
		Success: true,
		Data:    JudgeQueueStatus(output.Status),
	}
}

func (p *getJudgeQueuePresenter) Error(err error) api.GetJudgeQueueResponseObject {
	return httperror.From(err).Response()
}
