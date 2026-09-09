package buildquery_test

import (
	"judge-backend/internal/app/shared/infrastructure/logger"
	"judge-backend/internal/app/shared/model"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/service/query/build_query/models/buildinfo"
	buildstatus "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/status"
	buildquery "judge-backend/internal/app/evaluation_environment_domain/execution_environment/infrastructure/service/query/build_query"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/mocks"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	assert.NoError(t, err)

	return gormDB, mock
}

func TestGetBuilds(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	repo := &mocks.BuildRepository{}
	service := buildquery.NewService(repo, gormDB, logger.NewMockLogger())

	t.Run("GetBuilds returns all builds for a given tag", func(t *testing.T) {
		image := &model.Image{ID: 1, Tag: "nginx:latest"}
		mock.ExpectQuery("SELECT \\* FROM `images` WHERE tag = ? .* LIMIT ?").WithArgs("nginx:latest", 1).WillReturnRows(
			sqlmock.NewRows([]string{"id", "tag"}).AddRow(image.ID, image.Tag),
		)

		builds := []model.Build{
			{Batch: 1, Status: buildstatus.SUCCESS.Value(), ImageID: image.ID},
			{Batch: 2, Status: buildstatus.ERROR.Value(), ImageID: image.ID},
		}
		mock.ExpectQuery("SELECT \\* FROM `builds` WHERE image_id = ?").WithArgs(image.ID).WillReturnRows(
			sqlmock.NewRows([]string{"batch", "status", "image_id", "created_at", "updated_at"}).
				AddRow(builds[0].Batch, builds[0].Status, builds[0].ImageID, time.Now(), time.Now()).
				AddRow(builds[1].Batch, builds[1].Status, builds[1].ImageID, time.Now(), time.Now()),
		)

		repo.On("FindCurrentBuildByTag", "nginx:latest").Return(nil)

		buildInfos, err := service.GetBuilds("nginx:latest")
		assert.NoError(t, err)
		assert.Len(t, buildInfos, 2)

		expected := []buildinfo.BuildInfo{
			{Batch: 1, Status: "SUCCESS"},
			{Batch: 2, Status: "ERROR"},
		}
		assert.Equal(t, expected, buildInfos)
	})

	t.Run("GetBuilds returns error when image not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT \\* FROM `images` WHERE tag = ? .* LIMIT ?").WithArgs("nginx:latest", 1).WillReturnError(gorm.ErrRecordNotFound)

		_, err := service.GetBuilds("nginx:latest")
		assert.Error(t, err)
	})

	t.Run("GetBuilds returns error on DB failure", func(t *testing.T) {
		mock.ExpectQuery("SELECT \\* FROM `images` WHERE tag = ? .* LIMIT ?").WithArgs("nginx:latest", 1).WillReturnRows(
			sqlmock.NewRows([]string{"id", "tag"}),
		)

		mock.ExpectQuery("SELECT \\* FROM `builds` WHERE image_id = ?").WithArgs(1).WillReturnError(gorm.ErrInvalidDB)
		_, err := service.GetBuilds("nginx:latest")
		assert.Error(t, err)
	})
}
