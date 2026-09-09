package judgeworker

import "judge-backend/internal/app/judge_domain/worker/domain/entity/judgeworker/judgeorder"

type JudgeWorker struct {
	id         string
	judgeOrder []judgeorder.JudgeOrder
}

func New(
	judgeOrder []string,
) (*JudgeWorker, error) {
	order := make([]judgeorder.JudgeOrder, 0)
	for _, v := range judgeOrder {
		o, err := judgeorder.New(v)
		if err != nil {
			return nil, err
		}
		order = append(order, *o)
	}

	return &JudgeWorker{
		judgeOrder: order,
	}, nil
}

func (j *JudgeWorker) ID() string {
	return j.id
}

func (j *JudgeWorker) SetID(id string) {
	j.id = id
}

func (j *JudgeWorker) JudgeOrder() []judgeorder.JudgeOrder {
	return j.judgeOrder
}
