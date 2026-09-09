package service

import (
	"judge-backend/internal/app/judge_domain/worker/domain/entity/judgeworker"
	"judge-backend/internal/app/judge_domain/worker/domain/provider"
	"judge-backend/internal/app/judge_domain/worker/domain/service"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
)

type JudgeWorkerService struct {
	workerPool provider.WorkerPool
}

func NewJudgeWorkerService(workerPool provider.WorkerPool) service.JudgeWorkerService {
	return &JudgeWorkerService{
		workerPool: workerPool,
	}
}

func (s *JudgeWorkerService) Start(judgeWorker *judgeworker.JudgeWorker) error {
	handle := s.workerPool.Get(judgeWorker.ID())
	if handle == nil {
		return httperror.New(http.StatusNotFound, "JudgeWorker not found")
	}

	handle.Start()
	return nil
}

func (s *JudgeWorkerService) Stop(judgeWorker *judgeworker.JudgeWorker) error {
	handle := s.workerPool.Get(judgeWorker.ID())
	if handle == nil {
		return httperror.New(http.StatusNotFound, "JudgeWorker not found")
	}

	handle.Stop()
	return nil
}
