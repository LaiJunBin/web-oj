package updateasset_test

import (
	"bytes"
	"errors"
	"judge-backend/internal/app/shared/test"
	updateasset "judge-backend/internal/app/project_domain/project/application/usecase/asset/update_asset"
	"judge-backend/internal/app/project_domain/project/domain/entity/asset"
	"judge-backend/internal/app/project_domain/project/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

type UsecaseMocks struct {
	repository *mocks.AssetRepository
	service    *mocks.AssetService
}

func NewUsecaseMocks(t *testing.T) *UsecaseMocks {
	return &UsecaseMocks{
		repository: mocks.NewAssetRepository(t),
		service:    mocks.NewAssetService(t),
	}
}

func NewUsecase(t *testing.T) (updateasset.Usecase, *UsecaseMocks) {
	mocks := NewUsecaseMocks(t)
	return updateasset.NewUsecase(mocks.repository, mocks.service), mocks
}

func TestExecute(t *testing.T) {
	t.Run("should update asset successfully", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &updateasset.Input{
			ID:        "1",
			FileName:  test.StringPtr("updated.txt"),
			File:      bytes.NewBufferString("updated content"),
			IsEnabled: test.BoolPtr(true),
		}
		asset := &asset.Asset{}
		mocks.repository.On("FindByID", input.ID).Return(asset, nil).Once()
		mocks.service.On("UploadFile", asset.ProjectID().Value(), *input.FileName, input.File).Return(asset, nil).Once()
		mocks.repository.On("Delete", asset).Return(nil).Once()
		mocks.service.On("DeleteFile", asset.ID().String()).Return(nil).Once()
		mocks.repository.On("Save", asset).Return(nil).Once()

		output, err := usecase.Execute(input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should return error when asset not found", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &updateasset.Input{ID: "1", FileName: test.StringPtr("updated.txt")}
		mocks.repository.On("FindByID", input.ID).Return(nil, nil).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, updateasset.ErrorAssetNotFound, err)
	})

	t.Run("should return error when upload service fails", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &updateasset.Input{
			ID:       "1",
			FileName: test.StringPtr("updated.txt"),
			File:     bytes.NewBufferString("updated content"),
		}
		asset := &asset.Asset{}
		mocks.repository.On("FindByID", input.ID).Return(asset, nil).Once()
		mocks.service.On("UploadFile", asset.ProjectID().Value(), *input.FileName, input.File).Return(nil, errors.New("upload error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, updateasset.ErrorInternalServer, err)
	})

	t.Run("should return error when repository save fails", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &updateasset.Input{
			ID:       "1",
			FileName: test.StringPtr("updated.txt"),
			File:     bytes.NewBufferString("updated content"),
		}
		asset := &asset.Asset{}
		mocks.repository.On("FindByID", input.ID).Return(asset, nil).Once()
		mocks.service.On("UploadFile", asset.ProjectID().Value(), *input.FileName, input.File).Return(asset, nil).Once()
		mocks.repository.On("Delete", asset).Return(nil).Once()
		mocks.service.On("DeleteFile", asset.ID().String()).Return(nil).Once()
		mocks.repository.On("Save", asset).Return(errors.New("save error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, updateasset.ErrorInternalServer, err)
	})
}
