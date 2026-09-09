package query

import (
	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
)

var (
	ErrorContestNotFound = httperror.New(http.StatusNotFound, "contest not found")
)

type ContestQuery interface {
	ListPublicContests() ([]dto.Contest, error)
	ListUserContests(userID int32) ([]dto.Contest, error)

	GetPublicContestByID(contestID int32) (dto.ContestDetail, error)
	GetUserContestByID(userID, contestID int32) (dto.ContestDetail, error)

	GetUserContestScheduleByID(userID, contestID int32) (dto.ContestSchedule, error)
	GetPublicContestScheduleByID(contestID int32) (dto.ContestSchedule, error)
}
