package downloadasset_test

import (
	"errors"
	"judge-backend/internal/app/project_domain/project/application/query"
	"judge-backend/internal/app/project_domain/project/application/query/dto"
	downloadasset "judge-backend/internal/app/project_domain/project/application/usecase/asset/download_asset"
	"judge-backend/internal/app/project_domain/project/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

type UsecaseMocks struct {
	query *mocks.AssetQuery
}

func NewUsecaseMocks(t *testing.T) *UsecaseMocks {
	return &UsecaseMocks{
		query: mocks.NewAssetQuery(t),
	}
}

func NewUsecase(t *testing.T) (downloadasset.Usecase, *UsecaseMocks) {
	mocks := NewUsecaseMocks(t)
	return downloadasset.NewUsecase(mocks.query), mocks
}

func TestExecute(t *testing.T) {
	t.Run("should download asset successfully", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &downloadasset.Input{FileName: "test.txt"}
		file := &dto.DownloadFile{}
		mocks.query.On("DownloadAsset", input.FileName).Return(file, nil).Once()

		output, err := usecase.Execute(input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, file, output.File)
	})

	t.Run("should return error when file name is empty", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &downloadasset.Input{FileName: ""}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, downloadasset.ErrorInvalidFileName, err)
	})

	t.Run("should return error when asset not found", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &downloadasset.Input{FileName: "test.txt"}
		mocks.query.On("DownloadAsset", input.FileName).Return(nil, query.ErrorAssetNotFound).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, downloadasset.ErrorAssetNotFound, err)
	})

	t.Run("should return internal server error when query fails", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &downloadasset.Input{FileName: "test.txt"}
		mocks.query.On("DownloadAsset", input.FileName).Return(nil, errors.New("query error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, downloadasset.ErrorInternalServer, err)
	})
}
