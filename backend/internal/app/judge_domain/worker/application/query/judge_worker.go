package query

import "judge-backend/internal/app/judge_domain/worker/application/query/dto"

type JudgeWorkerQuery interface {
	ListJudgeWorker() ([]*dto.JudgeWorkerInfo, error)
	ListJudgePriorityStrategies() ([]string, error)
	GetJudgeWorkerByID(judgeWorkerID string) (*dto.JudgeWorker, error)
}
