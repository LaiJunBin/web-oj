package repository

import (
	"judge-backend/internal/app/judge_domain/worker/domain/entity/judgeworker"
	"judge-backend/internal/app/judge_domain/worker/domain/provider"
	"judge-backend/internal/app/judge_domain/worker/domain/repository"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
)

type JudgeWorkerRepository struct {
	workerPool provider.WorkerPool
}

func NewJudgeWorkerRepository(workerPool provider.WorkerPool) repository.JudgeWorkerRepository {
	return &JudgeWorkerRepository{
		workerPool: workerPool,
	}
}

func (r *JudgeWorkerRepository) FindByID(judgeWorkerID string) (*judgeworker.JudgeWorker, error) {
	handle := r.workerPool.Get(judgeWorkerID)
	if handle == nil {
		return nil, httperror.New(http.StatusNotFound, "JudgeWorker not found")
	}

	judgeWorker, err := judgeworker.New(handle.JudgeOrder())
	if err != nil {
		return nil, err
	}

	judgeWorker.SetID(handle.ID())

	return judgeWorker, nil
}

func (r *JudgeWorkerRepository) Save(judgeWorker *judgeworker.JudgeWorker) error {
	order := make([]string, 0, len(judgeWorker.JudgeOrder()))
	for _, o := range judgeWorker.JudgeOrder() {
		order = append(order, o.String())
	}
	if err := r.workerPool.Add(order); err != nil {
		return httperror.New(http.StatusBadRequest, err.Error())
	}
	return nil
}

func (r *JudgeWorkerRepository) Delete(judgeWorker *judgeworker.JudgeWorker) error {
	r.workerPool.Delete(judgeWorker.ID())
	return nil
}
