package repository

import "judge-backend/internal/app/contest_domain/contest/domain/entity/contest"

type ContestRepository interface {
	FindByID(id int32) (*contest.Contest, error)
	Save(contest *contest.Contest) error
	Delete(contest *contest.Contest) error
}
