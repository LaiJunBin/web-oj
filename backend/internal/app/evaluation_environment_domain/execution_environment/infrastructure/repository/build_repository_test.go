package repository

import (
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/batch"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/buildloglevel"
	buildlog "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/log"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type BuildRepositoryMocks struct {
	sqlMock sqlmock.Sqlmock
}

func newBuildRepository(t *testing.T) (*BuildRepository, *BuildRepositoryMocks) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating sql mock: %v", err)
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("error creating gorm db: %v", err)
	}

	r := NewBuildRepository(gormDB)
	repo, ok := r.(*BuildRepository)
	if !ok {
		t.Fatalf("error creating test repository")
	}

	mocks := &BuildRepositoryMocks{
		sqlMock: sqlMock,
	}

	return repo, mocks
}

func TestBuildRepository_FindByTagAndBatch(t *testing.T) {
	t.Run("non-existing build", func(t *testing.T) {
		repo, _ := newBuildRepository(t)
		mockImage := image.MustNew(&image.Config{
			Tag:       "test-nginx:latest",
			MountPath: "/var/www/html",
		})
		mockBatch := batch.MustNew(1)

		build, err := repo.FindByTagAndBatch(mockImage.Tag().String(), mockBatch.Value())
		if err != nil {
			t.Fatalf("FindByTag(%s) should not return an error, but got %v", mockImage.Tag().String(), err)
		}

		if build != nil {
			t.Fatalf("FindByTag(%s) should return nil, but got a build", mockImage.Tag().String())
		}
	})

	t.Run("existing build", func(t *testing.T) {
		repo, _ := newBuildRepository(t)
		mockImage := image.MustNew(&image.Config{
			Tag:       "test-nginx:latest",
			MountPath: "/var/www/html",
		})
		mockBatch := batch.MustNew(1)
		mockBuildConfig := &build.BuildConfig{
			Tag:   mockImage.Tag(),
			Batch: mockBatch,
		}
		mockBuild := build.New(mockBuildConfig)
		repo.buildMap[mockBuild.Tag()] = mockBuild

		build, err := repo.FindByTagAndBatch(mockImage.Tag().String(), mockBatch.Value())
		if err != nil {
			t.Fatalf("FindByTag(%s) should not return an error, but got %v", mockBuild.Tag(), err)
		}

		if build == nil {
			t.Fatalf("FindByTag(%s) should return a build, but got nil", mockBuild.Tag())
		}

		if build.Tag() != mockBuild.Tag() {
			t.Errorf("FindByTag(%s) should return a build with tag %s, but got %s", mockBuild.Tag(), mockBuild.Tag(), build.Tag())
		}

		if build.Batch().Value() != mockBuild.Batch().Value() {
			t.Errorf("FindByTag(%s) should return a build with batch %d, but got %d", mockBuild.Tag(), mockBuild.Batch().Value(), build.Batch().Value())
		}
	})
}

func TestBuildRepository_SavePendingExclusive_And_FindByTagAndBatch(t *testing.T) {
	t.Run("test new build", func(t *testing.T) {
		repo, _ := newBuildRepository(t)
		mockImage := image.MustNew(&image.Config{
			Tag:       "test-nginx:latest",
			MountPath: "/var/www/html",
		})
		mockBatch := batch.MustNew(1)
		mockBuildConfig := &build.BuildConfig{
			Tag:   mockImage.Tag(),
			Batch: mockBatch,
		}
		mockBuild := build.New(mockBuildConfig)

		err := repo.SavePendingExclusive(mockBuild)
		if err != nil {
			t.Fatalf("SavePendingExclusive(%s) should not return an error, but got %v", mockBuild.Tag(), err)
		}

		build, err := repo.FindByTagAndBatch(mockImage.Tag().String(), mockBatch.Value())
		if err != nil {
			t.Fatalf("FindByTag(%s) should not return an error, but got %v", mockBuild.Tag(), err)
		}

		if build == nil {
			t.Fatalf("FindByTag(%s) should return a build, but got nil", mockBuild.Tag())
		}

		if build.Tag() != mockBuild.Tag() {
			t.Errorf("FindByTag(%s) should return a build with tag %s, but got %s", mockBuild.Tag(), mockBuild.Tag(), build.Tag())
		}

		if build.Batch().Value() != mockBuild.Batch().Value() {
			t.Errorf("FindByTag(%s) should return a build with batch %d, but got %d", mockBuild.Tag(), mockBuild.Batch().Value(), build.Batch().Value())
		}
	})

	t.Run("test existing build with logs", func(t *testing.T) {
		repo, _ := newBuildRepository(t)
		mockImage := image.MustNew(&image.Config{
			Tag:       "test-nginx:latest",
			MountPath: "/var/www/html",
		})
		mockBatch := batch.MustNew(1)
		mockBuildConfig := &build.BuildConfig{
			Tag:   mockImage.Tag(),
			Batch: mockBatch,
		}
		mockBuild := build.New(mockBuildConfig)
		mockBuild.AddLog(buildlog.New(buildloglevel.Info, "Hello, Go!", time.Now()))
		mockBuild.AddLog(buildlog.New(buildloglevel.Info, "Hello, World!", time.Now()))

		err := repo.SavePendingExclusive(mockBuild)
		if err != nil {
			t.Fatalf("SavePendingExclusive(%s) should not return an error, but got %v", mockBuild.Tag(), err)
		}

		build, err := repo.FindByTagAndBatch(mockImage.Tag().String(), mockBatch.Value())
		if err != nil {
			t.Fatalf("FindByTag(%s) should not return an error, but got %v", mockBuild.Tag(), err)
		}

		if build == nil {
			t.Fatalf("FindByTag(%s) should return a build, but got nil", mockBuild.Tag())
		}

		if build.Tag() != mockBuild.Tag() {
			t.Errorf("FindByTag(%s) should return a build with tag %s, but got %s", mockBuild.Tag(), mockBuild.Tag(), build.Tag())
		}

		if build.Batch().Value() != mockBuild.Batch().Value() {
			t.Errorf("FindByTag(%s) should return a build with batch %d, but got %d", mockBuild.Tag(), mockBuild.Batch().Value(), build.Batch().Value())
		}

		if build.Logs() == nil {
			t.Fatalf("FindByTag(%s) should return a build with logs, but got nil", mockBuild.Tag())
		}

		if len(build.Logs()) != 2 {
			t.Fatalf("FindByTag(%s) should return a build with 1 log, but got %d", mockBuild.Tag(), len(build.Logs()))
		}
	})
}

func TestBuildRepository_FindLastBuild(t *testing.T) {
	t.Run("non-existing build", func(t *testing.T) {
		repo, mocks := newBuildRepository(t)
		mockImage := image.MustNew(&image.Config{
			Tag:       "test-nginx:latest",
			MountPath: "/var/www/html",
		})
		mocks.sqlMock.ExpectQuery("SELECT \\* FROM `images` WHERE tag = ? .* LIMIT ?").
			WithArgs(mockImage.Tag().String(), 1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		build, err := repo.FindLastBuild(mockImage.Tag().String())
		if err != nil {
			t.Fatalf("FindLastBuild(%s) should not return an error, but got %v", mockImage.Tag().String(), err)
		}

		if build != nil {
			t.Fatalf("FindLastBuild(%s) should return nil, but got a build", mockImage.Tag().String())
		}
	})

	t.Run("existing build", func(t *testing.T) {
		repo, _ := newBuildRepository(t)
		mockImage := image.MustNew(&image.Config{
			Tag:       "test-nginx:latest",
			MountPath: "/var/www/html",
		})
		mockBatch := batch.MustNew(1)
		mockBuildConfig := &build.BuildConfig{
			Tag:   mockImage.Tag(),
			Batch: mockBatch,
		}
		mockBuild := build.New(mockBuildConfig)
		repo.buildMap[mockBuild.Tag()] = mockBuild

		build, err := repo.FindLastBuild(mockImage.Tag().String())
		if err != nil {
			t.Fatalf("FindLastBuild(%s) should not return an error, but got %v", mockBuild.Tag(), err)
		}

		if build == nil {
			t.Fatalf("FindLastBuild(%s) should return a build, but got nil", mockBuild.Tag())
		}

		if build.Tag() != mockBuild.Tag() {
			t.Errorf("FindLastBuild(%s) should return a build with tag %s, but got %s", mockBuild.Tag(), mockBuild.Tag(), build.Tag())
		}

		if build.Batch().Value() != mockBuild.Batch().Value() {
			t.Errorf("FindLastBuild(%s) should return a build with batch %d, but got %d", mockBuild.Tag(), mockBuild.Batch().Value(), build.Batch().Value())
		}
	})
}

func TestBuildRepository_FindCurrentBuildByTag(t *testing.T) {
	t.Run("non-existing build", func(t *testing.T) {
		repo, _ := newBuildRepository(t)
		mockImage := image.MustNew(&image.Config{
			Tag:       "test-nginx:latest",
			MountPath: "/var/www/html",
		})

		build := repo.FindCurrentBuildByTag(mockImage.Tag().String())
		if build != nil {
			t.Fatalf("FindCurrentBuildByTag(%s) should return nil, but got a build", mockImage.Tag().String())
		}
	})

	t.Run("existing build", func(t *testing.T) {
		repo, _ := newBuildRepository(t)
		mockImage := image.MustNew(&image.Config{
			Tag:       "test-nginx:latest",
			MountPath: "/var/www/html",
		})
		mockBatch := batch.MustNew(1)
		mockBuildConfig := &build.BuildConfig{
			Tag:   mockImage.Tag(),
			Batch: mockBatch,
		}
		mockBuild := build.New(mockBuildConfig)
		repo.buildMap[mockBuild.Tag()] = mockBuild

		build := repo.FindCurrentBuildByTag(mockImage.Tag().String())
		if build == nil {
			t.Fatalf("FindCurrentBuildByTag(%s) should return a build, but got nil", mockBuild.Tag())
		}

		if build.Tag() != mockBuild.Tag() {
			t.Errorf("FindCurrentBuildByTag(%s) should return a build with tag %s, but got %s", mockBuild.Tag(), mockBuild.Tag(), build.Tag())
		}

		if build.Batch().Value() != mockBuild.Batch().Value() {
			t.Errorf("FindCurrentBuildByTag(%s) should return a build with batch %d, but got %d", mockBuild.Tag(), mockBuild.Batch().Value(), build.Batch().Value())
		}
	})
}

func TestBuildRepository_Commit(t *testing.T) {
	t.Run("commit new build", func(t *testing.T) {
		repo, mocks := newBuildRepository(t)
		mockImage := image.MustNew(&image.Config{
			Tag:       "test-nginx:latest",
			MountPath: "/var/www/html",
		})
		mockBatch := batch.MustNew(1)
		mockBuildConfig := &build.BuildConfig{
			Tag:   mockImage.Tag(),
			Batch: mockBatch,
		}
		mockBuild := build.New(mockBuildConfig)
		mockBuild.AddLog(buildlog.New(buildloglevel.Info, "Hello, Go!", time.Now()))
		mockBuild.AddLog(buildlog.New(buildloglevel.Info, "Hello, World!", time.Now()))

		mocks.sqlMock.ExpectQuery("SELECT \\* FROM `images` WHERE tag = ? .* LIMIT ?").
			WithArgs(mockImage.Tag().String(), 1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		mocks.sqlMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `builds` WHERE (image_id = ? AND batch = ?) ")).
			WithArgs(1, mockBatch.Value(), 1).
			WillReturnError(gorm.ErrRecordNotFound)

		mocks.sqlMock.ExpectQuery("SELECT \\* FROM `images` WHERE tag = ? .* LIMIT ?").
			WithArgs(mockImage.Tag().String(), 1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		mocks.sqlMock.ExpectBegin()
		mocks.sqlMock.ExpectExec("DELETE").WillReturnResult(sqlmock.NewResult(0, 0))
		mocks.sqlMock.ExpectExec("INSERT INTO `builds`").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mocks.sqlMock.ExpectExec("INSERT INTO `build_logs`").
			WillReturnResult(sqlmock.NewResult(1, 2))
		mocks.sqlMock.ExpectCommit()

		err := repo.Commit(mockBuild)
		if err != nil {
			t.Fatalf("Commit(%s) should not return an error, but got %v", mockBuild.Tag(), err)
		}

		if err := mocks.sqlMock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("commit existing build", func(t *testing.T) {
		repo, mocks := newBuildRepository(t)
		mockImage := image.MustNew(&image.Config{
			Tag:       "test-nginx:latest",
			MountPath: "/var/www/html",
		})
		mockBatch := batch.MustNew(1)
		mockBuildConfig := &build.BuildConfig{
			Tag:   mockImage.Tag(),
			Batch: mockBatch,
		}
		mockBuild := build.New(mockBuildConfig)
		mockBuild.AddLog(buildlog.New(buildloglevel.Info, "Hello, Go!", time.Now()))
		mockBuild.AddLog(buildlog.New(buildloglevel.Info, "Hello, World!", time.Now()))

		mocks.sqlMock.ExpectQuery("SELECT \\* FROM `images` WHERE tag = ? .* LIMIT ?").
			WithArgs(mockImage.Tag().String(), 1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		mocks.sqlMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `builds` WHERE (image_id = ? AND batch = ?) ")).
			WithArgs(1, mockBatch.Value(), 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "image_id", "batch", "status"}).AddRow(1, 1, mockBatch.Value(), mockBuild.Status().Value()))

		mocks.sqlMock.ExpectBegin()
		mocks.sqlMock.ExpectExec("DELETE").WillReturnResult(sqlmock.NewResult(0, 0))
		mocks.sqlMock.ExpectExec("UPDATE `builds`").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mocks.sqlMock.ExpectExec("INSERT INTO `build_logs`").
			WillReturnResult(sqlmock.NewResult(1, 2))
		mocks.sqlMock.ExpectCommit()

		err := repo.Commit(mockBuild)
		if err != nil {
			t.Fatalf("Commit(%s) should not return an error, but got %v", mockBuild.Tag(), err)
		}

		if err := mocks.sqlMock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}
