package deleteasset_test

import (
	"errors"
	deleteasset "judge-backend/internal/app/project_domain/project/application/usecase/asset/delete_asset"
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

func NewUsecase(t *testing.T) (deleteasset.Usecase, *UsecaseMocks) {
	mocks := NewUsecaseMocks(t)
	return deleteasset.NewUsecase(mocks.repository, mocks.service), mocks
}

func TestExecute(t *testing.T) {
	t.Run("should delete asset successfully", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &deleteasset.Input{ID: "1"}
		asset := &asset.Asset{}
		mocks.repository.On("FindByID", input.ID).Return(asset, nil).Once()
		mocks.repository.On("Delete", asset).Return(nil).Once()
		mocks.service.On("DeleteFile", input.ID).Return(nil).Once()

		output, err := usecase.Execute(input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should return error when asset not found", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &deleteasset.Input{ID: "1"}
		mocks.repository.On("FindByID", input.ID).Return(nil, nil).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, deleteasset.ErrorAssetNotFound, err)
	})

	t.Run("should return internal server error when repository delete fails", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &deleteasset.Input{ID: "1"}
		asset := &asset.Asset{}
		mocks.repository.On("FindByID", input.ID).Return(asset, nil).Once()
		mocks.repository.On("Delete", asset).Return(errors.New("delete error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, deleteasset.ErrorInternalServer, err)
	})

	// t.Run("should return internal server error when service delete fails", func(t *testing.T) {
	// 	usecase, mocks := NewUsecase(t)
	// 	input := &deleteasset.Input{ID: "1"}
	// 	asset := &asset.Asset{}
	// 	mocks.repository.On("FindByID", input.ID).Return(asset, nil).Once()
	// 	mocks.repository.On("Delete", asset).Return(nil).Once()
	// 	mocks.service.On("DeleteFile", input.ID).Return(errors.New("service delete error")).Once()

	// 	_, err := usecase.Execute(input)

	// 	assert.Error(t, err)
	// 	assert.Equal(t, deleteasset.ErrorInternalServer, err)
	// })
}
