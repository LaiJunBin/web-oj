package query

import (
	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
)

var (
	ErrorContestProjectNotFound = httperror.New(http.StatusNotFound, "contest project not found")
	ErrorAssetNotFound          = httperror.New(http.StatusNotFound, "asset not found")
)

type ProjectQuery interface {
	GetPublicProjectByID(contestProjectID int32) (dto.ProjectDetail, error)
	GetUserProjectByID(userID, contestProjectID int32) (dto.UserProjectDetail, error)

	GetProjectScoreboardByID(contestProjectID int32) (*dto.Scoreboard, error)

	GetUserProjectAssetFile(userID, contestProjectID int32, assetID string) (dto.AssetFile, error)
}
