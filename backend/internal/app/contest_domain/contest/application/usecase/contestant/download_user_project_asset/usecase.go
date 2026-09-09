package downloaduserprojectasset

import (
	"judge-backend/internal/app/contest_domain/contest/application/query"
	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type Input struct {
	UserID           int32
	ContestProjectID int32
	AssetID          string
}

type Output struct {
	File dto.AssetFile
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	query  query.ProjectQuery
	logger sharedtypes.Logger
}

func NewUsecase(
	query query.ProjectQuery,
	logger sharedtypes.Logger,
) Usecase {
	return &usecase{
		query:  query,
		logger: logger,
	}
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	u.logger.Infof("Starting to download user project asset", "userId", input.UserID, "projectId", input.ContestProjectID, "assetId", input.AssetID)

	file, err := u.query.GetUserProjectAssetFile(input.UserID, input.ContestProjectID, input.AssetID)
	if err != nil {
		u.logger.Errorf("Failed to get user project asset file", "userId", input.UserID, "projectId", input.ContestProjectID, "assetId", input.AssetID, "error", err)
		return nil, err
	}

	u.logger.Infof("Successfully got user project asset file", "userId", input.UserID, "projectId", input.ContestProjectID, "assetId", input.AssetID)
	return &Output{
		File: file,
	}, nil
}
