package query

import (
	"judge-backend/internal/app/judge_domain/worker/application/query"
	"judge-backend/internal/app/judge_domain/worker/application/query/dto"
	"judge-backend/internal/app/judge_domain/worker/domain/prioritystrategy"
	"judge-backend/internal/app/judge_domain/worker/domain/provider"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
)

type JudgeWorkerQuery struct {
	workerPool provider.WorkerPool
}

func NewJudgeWorkerQuery(
	workerPool provider.WorkerPool,
) query.JudgeWorkerQuery {
	return &JudgeWorkerQuery{
		workerPool: workerPool,
	}
}

func (j *JudgeWorkerQuery) ListJudgeWorker() ([]*dto.JudgeWorkerInfo, error) {
	handles := j.workerPool.List()
	judgeWorkerInfos := make([]*dto.JudgeWorkerInfo, 0, len(handles))
	for _, handle := range handles {
		judgeWorkerInfos = append(judgeWorkerInfos, &dto.JudgeWorkerInfo{
			ID:              handle.ID(),
			Status:          handle.Status(),
			JudgeOrder:      handle.JudgeOrder(),
			CreatedAt:       handle.CreatedAt(),
			LastExecutionAt: handle.LastExecutionAt(),
		})
	}

	return judgeWorkerInfos, nil
}

func (j *JudgeWorkerQuery) ListJudgePriorityStrategies() ([]string, error) {
	strategies := prioritystrategy.All()
	result := make([]string, 0, len(strategies))
	for _, s := range strategies {
		result = append(result, s.String())
	}
	return result, nil
}

func (j *JudgeWorkerQuery) GetJudgeWorkerByID(judgeWorkerID string) (*dto.JudgeWorker, error) {
	handle := j.workerPool.Get(judgeWorkerID)
	if handle == nil {
		return nil, httperror.New(http.StatusNotFound, "JudgeWorker not found")
	}

	logs := []*dto.Log{}
	for _, l := range handle.Logs() {
		logs = append(logs, &dto.Log{
			Text: l.Text,
			Time: l.Time,
		})
	}

	return &dto.JudgeWorker{
		ID:              handle.ID(),
		Status:          handle.Status(),
		JudgeOrder:      handle.JudgeOrder(),
		CreatedAt:       handle.CreatedAt(),
		LastExecutionAt: handle.LastExecutionAt(),
		Logs:            logs,
	}, nil
}
