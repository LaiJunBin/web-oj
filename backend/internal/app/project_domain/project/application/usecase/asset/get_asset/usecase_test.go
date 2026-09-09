package getasset_test

import (
	"errors"
	"judge-backend/internal/app/project_domain/project/application/query/dto"
	getasset "judge-backend/internal/app/project_domain/project/application/usecase/asset/get_asset"
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

func NewUsecase(t *testing.T) (getasset.Usecase, *UsecaseMocks) {
	mocks := NewUsecaseMocks(t)
	return getasset.NewUsecase(mocks.query), mocks
}

func TestExecute(t *testing.T) {
	t.Run("should get asset successfully", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &getasset.Input{ID: "1"}
		asset := &dto.Asset{}
		mocks.query.On("GetAsset", input.ID).Return(asset, nil).Once()

		output, err := usecase.Execute(input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, asset, output.Asset)
	})

	t.Run("should return error when id is invalid", func(t *testing.T) {
		usecase, _ := NewUsecase(t)
		input := &getasset.Input{ID: ""}

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, getasset.ErrorInvalidID, err)
	})

	t.Run("should return error when asset not found", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &getasset.Input{ID: "1"}
		mocks.query.On("GetAsset", input.ID).Return(nil, nil).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, getasset.ErrorAssetNotFound, err)
	})

	t.Run("should return internal server error when repository find fails", func(t *testing.T) {
		usecase, mocks := NewUsecase(t)
		input := &getasset.Input{ID: "1"}
		mocks.query.On("GetAsset", input.ID).Return(nil, errors.New("find error")).Once()

		_, err := usecase.Execute(input)

		assert.Error(t, err)
		assert.Equal(t, getasset.ErrorInternalServer, err)
	})
}
