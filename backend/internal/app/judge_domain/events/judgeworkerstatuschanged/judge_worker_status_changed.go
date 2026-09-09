package judgeworkerstatuschanged

type JudgeWorkerStatusChangedEvent struct {
	id     string
	status string
}

func (*JudgeWorkerStatusChangedEvent) Topic() string {
	return "judge_worker_status_changed"
}

func NewJudgeWorkerStatusChangedEvent(id, status string) *JudgeWorkerStatusChangedEvent {
	return &JudgeWorkerStatusChangedEvent{
		id:     id,
		status: status,
	}
}

func (e *JudgeWorkerStatusChangedEvent) ID() string {
	return e.id
}

func (e *JudgeWorkerStatusChangedEvent) Status() string {
	return e.status
}
