package service

import "judge-backend/internal/app/judge_domain/worker/domain/entity/judgeworker"

type JudgeWorkerService interface {
	Start(judgeWorker *judgeworker.JudgeWorker) error
	Stop(judgeWorker *judgeworker.JudgeWorker) error
}
