package uploadasset_test

import (
	"bytes"
	"errors"
	uploadasset "judge-backend/internal/app/project_domain/project/application/usecase/asset/upload_asset"
	"judge-backend/internal/app/project_domain/project/domain/entity/asset"
	"judge-backend/internal/app/project_domain/project/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

type UsecaseMocks struct {
	projectRepo *mocks.ProjectRepository
	repository  *mocks.AssetRepository
	service     *mocks.AssetService
}

func NewUsecaseMocks(t *testing.T) *UsecaseMocks {
	return &UsecaseMocks{
		projectRepo: mocks.NewProjectRepository(t),
		repository:  mocks.NewAssetRepository(t),
		service:     mocks.NewAssetService(t),
	}
}

func NewUsecase(t *testing.T) (uploadasset.Usecase, *UsecaseMocks) {
	mocks := NewUsecaseMocks(t)
	return uploadasset.NewUsecase(mocks.repository, mocks.projectRepo, mocks.service), mocks
}

func TestExecute(t *testing.T) {
	t.Run("should upload asset successfully", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &uploadasset.Input{
			ProjectID: 1,
			FileName:  "test.txt",
			File:      bytes.NewBufferString("test content"),
		}
		mocks.projectRepo.On("Exists", int(input.ProjectID)).Return(true, nil).Once()
		mocks.service.On("UploadFile", input.ProjectID, input.FileName, input.File).Return(&asset.Asset{}, nil).Once()
		mocks.repository.On("Save", &asset.Asset{}).Return(nil).Once()

		output, err := usecase.Execute(input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should return error when project does not exist", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &uploadasset.Input{
			ProjectID: 1,
			FileName:  "test.txt",
			File:      bytes.NewBufferString("test content"),
		}
		mocks.projectRepo.On("Exists", int(input.ProjectID)).Return(false, nil).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, uploadasset.ErrorProjectNotFound, err)
	})

	t.Run("should return error when upload service fails", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &uploadasset.Input{
			ProjectID: 1,
			FileName:  "test.txt",
			File:      bytes.NewBufferString("test content"),
		}
		mocks.projectRepo.On("Exists", int(input.ProjectID)).Return(true, nil).Once()
		mocks.service.On("UploadFile", input.ProjectID, input.FileName, input.File).Return(nil, errors.New("upload error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, uploadasset.ErrorInternalServer, err)
	})

	t.Run("should return error when repository save fails", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &uploadasset.Input{
			ProjectID: 1,
			FileName:  "test.txt",
			File:      bytes.NewBufferString("test content"),
		}
		mocks.projectRepo.On("Exists", int(input.ProjectID)).Return(true, nil).Once()
		mocks.service.On("UploadFile", input.ProjectID, input.FileName, input.File).Return(&asset.Asset{}, nil).Once()
		mocks.repository.On("Save", &asset.Asset{}).Return(errors.New("save error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, uploadasset.ErrorInternalServer, err)
	})
}
