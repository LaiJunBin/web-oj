package downloadasset

import (
	"errors"
	"judge-backend/internal/app/project_domain/project/application/query"
	"judge-backend/internal/app/project_domain/project/application/query/dto"
)

var (
	ErrorInvalidFileName = errors.New("invalid file name, must not be empty")
	ErrorAssetNotFound   = errors.New("asset not found")
	ErrorInternalServer  = errors.New("internal server error")
)

type Input struct {
	FileName string
}

type Output struct {
	File *dto.DownloadFile
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
	if input.FileName == "" {
		return ErrorInvalidFileName
	}

	return nil
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if err := u.validate(input); err != nil {
		return nil, err
	}

	file, err := u.query.DownloadAsset(input.FileName)
	if err != nil {
		if errors.Is(err, query.ErrorAssetNotFound) {
			return nil, ErrorAssetNotFound
		}
		return nil, ErrorInternalServer
	}

	return &Output{
		File: file,
	}, nil
}
