package repository

import "judge-backend/internal/app/judge_domain/worker/domain/entity/judgeworker"

type JudgeWorkerRepository interface {
	FindByID(judgeWorkerID string) (*judgeworker.JudgeWorker, error)
	Save(judgeWorker *judgeworker.JudgeWorker) error
	Delete(judgeWorker *judgeworker.JudgeWorker) error
}
