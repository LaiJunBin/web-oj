package getasset

import (
	"errors"
	"judge-backend/internal/app/project_domain/project/application/query"
	"judge-backend/internal/app/project_domain/project/application/query/dto"
)

var (
	ErrorInvalidID      = errors.New("invalid id, id must not be empty")
	ErrorAssetNotFound  = errors.New("asset not found")
	ErrorInternalServer = errors.New("internal server error")
)

type Input struct {
	ID string
}

type Output struct {
	Asset *dto.Asset
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	query query.AssetQuery
}

func NewUsecase(
	query query.AssetQuery,
) Usecase {
	return &usecase{
		query: query,
	}
}

func (u *usecase) validate(input *Input) error {
	if input.ID == "" {
		return ErrorInvalidID
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	asset, err := u.query.GetAsset(input.ID)
	if err != nil {
		return nil, ErrorInternalServer
	}

	if asset == nil {
		return nil, ErrorAssetNotFound
	}

	return &Output{
		Asset: asset,
	}, nil
}
