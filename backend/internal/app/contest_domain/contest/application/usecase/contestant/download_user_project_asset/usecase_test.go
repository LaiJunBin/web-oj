package downloaduserprojectasset_test

import (
	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	downloaduserprojectasset "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/download_user_project_asset"
	"judge-backend/internal/app/contest_domain/contest/mocks"
	sharedtypes "judge-backend/internal/app/shared/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

type noopLogger struct{}

func (l *noopLogger) Debugf(msg string, args ...interface{}) {}
func (l *noopLogger) Infof(msg string, args ...interface{})  {}
func (l *noopLogger) Warnf(msg string, args ...interface{})  {}
func (l *noopLogger) Errorf(msg string, args ...interface{}) {}
func (l *noopLogger) Fatalf(msg string, args ...interface{}) {}
func (l *noopLogger) Panicf(msg string, args ...interface{}) {}

var _ sharedtypes.Logger = (*noopLogger)(nil)

type UsecaseMocks struct {
	query *mocks.ProjectQuery
}

func NewUsecaseMocks(t *testing.T) *UsecaseMocks {
	return &UsecaseMocks{
		query: mocks.NewProjectQuery(t),
	}
}

func NewUsecase(t *testing.T) (downloaduserprojectasset.Usecase, *UsecaseMocks) {
	m := NewUsecaseMocks(t)
	return downloaduserprojectasset.NewUsecase(m.query, &noopLogger{}), m
}

func TestUsecase_Execute(t *testing.T) {
	t.Run("should return asset file successfully", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		assetFile := dto.AssetFile{Name: "test.zip"}
		m.query.On("GetUserProjectAssetFile", int32(1), int32(10), "asset-123").Return(assetFile, nil).Once()

		output, err := usecase.Execute(&downloaduserprojectasset.Input{
			UserID:           1,
			ContestProjectID: 10,
			AssetID:          "asset-123",
		})

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("should return error when query fails", func(t *testing.T) {
		usecase, m := NewUsecase(t)
		m.query.On("GetUserProjectAssetFile", int32(1), int32(10), "asset-123").Return(dto.AssetFile{}, assert.AnError).Once()

		_, err := usecase.Execute(&downloaduserprojectasset.Input{
			UserID:           1,
			ContestProjectID: 10,
			AssetID:          "asset-123",
		})

		assert.Error(t, err)
	})
}
