package query

import "judge-backend/internal/app/judge_domain/judge/application/query/dto"

type JudgehostQuery interface {
	ListJudgehosts() ([]*dto.JudgehostInfo, error)
	GetJudgehostByID(judgehostID string) (*dto.Judgehost, error)
}
