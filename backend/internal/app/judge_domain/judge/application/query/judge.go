package query

import "judge-backend/internal/app/judge_domain/judge/application/query/dto"

type JudgeQuery interface {
	ListJudges(contestProjectID int32) ([]*dto.Judge, error)
	GetJudgeByID(judgeID string) (*dto.Judge, error)
}
