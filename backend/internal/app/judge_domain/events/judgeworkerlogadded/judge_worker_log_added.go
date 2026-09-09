package judgeworkerlogadded

import "time"

type JudgeWorkerLogAddedEvent struct {
	id        string
	log       string
	timestamp time.Time
}

func (*JudgeWorkerLogAddedEvent) Topic() string {
	return "judge_worker_log_added"
}

func NewJudgeWorkerLogAddedEvent(id, log string, time time.Time) *JudgeWorkerLogAddedEvent {
	return &JudgeWorkerLogAddedEvent{
		id:        id,
		log:       log,
		timestamp: time,
	}
}

func (e *JudgeWorkerLogAddedEvent) ID() string {
	return e.id
}

func (e *JudgeWorkerLogAddedEvent) Log() string {
	return e.log
}

func (e *JudgeWorkerLogAddedEvent) Timestamp() time.Time {
	return e.timestamp
}
