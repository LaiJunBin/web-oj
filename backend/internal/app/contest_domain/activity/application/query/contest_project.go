package query

import "judge-backend/internal/app/contest_domain/activity/application/query/dto"

type ContestProjectQuery interface {
	GetContestProjectByID(id int32) (*dto.ContestProject, error)
	GetUserScores(contestProjectID int32) ([]*dto.UserScore, error)
	GetContestantScoreDetails(contestProjectID int32) ([]*dto.ContestantScoreDetail, error)
	GetJudgmentAnalytics(contestProjectID int32) (*dto.JudgmentAnalytics, error)
}
