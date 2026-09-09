package query

import "judge-backend/internal/app/judge_domain/worker/application/query/dto"

type JudgeQueueQuery interface {
	GetJudgeQueueStatus() (*dto.JudgeQueueStatus, error)
}
