package query

import "time"

type JudgmentQuery interface {
	GetPrevSubmissionTimeByJudgmentID(id string) (time.Time, error)
}
