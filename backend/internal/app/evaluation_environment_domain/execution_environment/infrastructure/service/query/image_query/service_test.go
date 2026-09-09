package imagequery_test

import (
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/service/query/image_query/models/imageinfo"
	imagequery "judge-backend/internal/app/evaluation_environment_domain/execution_environment/infrastructure/service/query/image_query"
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

func TestGetImages(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	service := imagequery.NewService(gormDB)

	t.Run("GetImages returns all images", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"tag", "description", "is_ready", "created_at", "updated_at"}).
			AddRow("nginx:latest", "description1", true, time.Now(), time.Now()).
			AddRow("nginx:alpine", "description2", false, time.Now(), time.Now())

		mock.ExpectQuery("SELECT \\* FROM `images`").WillReturnRows(rows)

		images, err := service.GetImages()
		assert.NoError(t, err)
		assert.Len(t, images, 2)

		expected := []imageinfo.ImageInfo{
			{Tag: "nginx:latest", Description: "description1", IsReady: true, CreatedAt: images[0].CreatedAt, UpdatedAt: images[0].UpdatedAt},
			{Tag: "nginx:alpine", Description: "description2", IsReady: false, CreatedAt: images[1].CreatedAt, UpdatedAt: images[1].UpdatedAt},
		}
		assert.Equal(t, expected, images)
	})

	t.Run("GetImages returns error on DB failure", func(t *testing.T) {
		mock.ExpectQuery("SELECT \\* FROM `images`").WillReturnError(gorm.ErrInvalidDB)

		_, err := service.GetImages()
		assert.Error(t, err)
	})
}
