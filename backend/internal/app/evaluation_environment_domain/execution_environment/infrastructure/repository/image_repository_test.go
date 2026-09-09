package repository_test

import (
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/infrastructure/repository"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/mocks"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func newImageRepository() (*repository.ImageRepository, sqlmock.Sqlmock, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}

	storagePathProvider := &mocks.StoragePathProvider{}
	storagePathProvider.On("GetStoragePath").Return("storage")
	r := repository.NewImageRepository(gormDB, storagePathProvider)
	repo, ok := r.(*repository.ImageRepository)
	if !ok {
		return nil, nil, err
	}

	return repo, mock, nil
}

type Row struct {
	ID          int
	Tag         string
	MountPath   string
	Description string
	Status      int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func TestImageRepository(t *testing.T) {
	t.Run("GetAll", func(t *testing.T) {
		t.Run("should return all images", func(t *testing.T) {
			repo, mock, err := newImageRepository()
			if err != nil {
				t.Fatal(err)
			}

			mockImages := []*Row{
				{ID: 1, Tag: "nginx:latest", MountPath: "/var/www/html", Description: "description1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
				{ID: 2, Tag: "nginx:alpine", MountPath: "/var/www/html", Description: "description2", CreatedAt: time.Now(), UpdatedAt: time.Time{}},
				{ID: 3, Tag: "mysql:latest", MountPath: "/var/www/html", Description: "description3", CreatedAt: time.Now(), UpdatedAt: time.Time{}},
				{ID: 4, Tag: "mysql:8.0.26", MountPath: "/var/www/html", Description: "description4", CreatedAt: time.Now(), UpdatedAt: time.Now()},
				{ID: 5, Tag: "test-nginx:latest", MountPath: "/var/www/html", Description: "", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			}

			rows := sqlmock.NewRows([]string{"tag", "mount_path", "description", "created_at", "updated_at"})
			for _, image := range mockImages {
				rows.AddRow(image.Tag, image.MountPath, image.Description, image.CreatedAt, image.UpdatedAt)
			}

			mock.ExpectQuery("SELECT").WillReturnRows(rows)

			images, err := repo.GetAll()
			if err != nil {
				t.Fatal(err)
			}

			if len(images) != 5 {
				t.Fatalf("expected 5 images, got %d", len(images))
			}

			for i, image := range images {
				if image.Tag().String() != mockImages[i].Tag {
					t.Fatalf("expected %s, got %s", mockImages[i].Tag, image.Tag().String())
				}
				if image.Description() != mockImages[i].Description {
					t.Fatalf("expected %s, got %s", mockImages[i].Description, image.Description())
				}
				if !image.CreatedAt().Equal(mockImages[i].CreatedAt) {
					t.Fatalf("expected %v, got %v", mockImages[i].CreatedAt, image.CreatedAt())
				}
				if !image.UpdatedAt().Equal(mockImages[i].UpdatedAt) {
					t.Fatalf("expected %v, got %v", mockImages[i].UpdatedAt, image.UpdatedAt())
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})

		t.Run("should return empty list", func(t *testing.T) {
			repo, mock, err := newImageRepository()
			if err != nil {
				t.Fatal(err)
			}

			mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id", "tag", "mount_path", "description", "created_at", "updated_at"}))

			images, err := repo.GetAll()
			if err != nil {
				t.Fatal(err)
			}

			if len(images) != 0 {
				t.Fatalf("expected 0 images, got %d", len(images))
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})

		t.Run("should return error", func(t *testing.T) {
			repo, mock, err := newImageRepository()
			if err != nil {
				t.Fatal(err)
			}

			mock.ExpectQuery("SELECT").WillReturnError(sqlmock.ErrCancelled)

			_, err = repo.GetAll()
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	})

	t.Run("FindByTag", func(t *testing.T) {
		t.Run("should return image", func(t *testing.T) {
			repo, mock, err := newImageRepository()
			if err != nil {
				t.Fatal(err)
			}

			createdAt := time.Now()
			updatedAt := time.Now()
			mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id", "tag", "description", "mount_path", "created_at", "updated_at"}).AddRow(1, "nginx:latest", "description1", "/var/www/html", createdAt, updatedAt))
			image, err := repo.FindByTag("nginx:latest")
			if err != nil {
				t.Fatal(err)
			}

			if image == nil {
				t.Fatal("expected not nil, got nil")
			}

			if image.Tag().String() != "nginx:latest" {
				t.Fatalf("expected nginx:latest, got %s", image.Tag().String())
			}

			if image.Description() != "description1" {
				t.Fatalf("expected description1, got %s", image.Description())
			}

			if !image.CreatedAt().Equal(createdAt) {
				t.Fatalf("expected %v, got %v", createdAt, image.CreatedAt())
			}

			if !image.UpdatedAt().Equal(updatedAt) {
				t.Fatalf("expected %v, got %v", updatedAt, image.UpdatedAt())
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})

		t.Run("should return nil", func(t *testing.T) {
			repo, mock, err := newImageRepository()
			if err != nil {
				t.Fatal(err)
			}

			mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{}))

			image, err := repo.FindByTag("nginx:latest")
			if err != nil {
				t.Fatal(err)
			}

			if image != nil {
				t.Fatalf("expected nil, got %v", image)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})

		t.Run("should return error", func(t *testing.T) {
			repo, mock, err := newImageRepository()
			if err != nil {
				t.Fatal(err)
			}

			mock.ExpectQuery("SELECT").WillReturnError(sqlmock.ErrCancelled)

			_, err = repo.FindByTag("nginx:latest")
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	})

	t.Run("Save", func(t *testing.T) {
		t.Run("should create new image", func(t *testing.T) {
			repo, mock, err := newImageRepository()
			if err != nil {
				t.Fatal(err)
			}

			mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(0))
			mock.ExpectBegin()
			mock.ExpectExec("INSERT").WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectCommit()

			img, _ := image.New(&image.Config{Tag: "nginx:latest", Description: "description1", MountPath: "/var/www/html"})
			err = repo.Save(img)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}

			if img.CreatedAt().IsZero() {
				t.Fatal("expected created_at not zero")
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})

		t.Run("should update existing image", func(t *testing.T) {
			repo, mock, err := newImageRepository()
			if err != nil {
				t.Fatal(err)
			}

			mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(0))
			mock.ExpectBegin()
			mock.ExpectExec("INSERT").WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectCommit()

			img, _ := image.New(&image.Config{Tag: "nginx:latest", Description: "description1", MountPath: "/var/www/html"})
			repo.Save(img)

			mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id", "tag", "description", "mount_path", "created_at", "updated_at"}).AddRow(1, "nginx:latest", "description1", "/var/www/html", time.Now(), time.Now()))
			mock.ExpectBegin()
			mock.ExpectExec("UPDATE").WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectCommit()

			img.SetTag("nginx:alpine")
			img.SetDescription("description2")
			err = repo.Save(img)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}

			if img.Tag().String() != "nginx:alpine" {
				t.Fatalf("expected nginx:alpine, got %s", img.Tag())
			}

			if img.Description() != "description2" {
				t.Fatalf("expected description2, got %s", img.Description())
			}

			if img.UpdatedAt().IsZero() {
				t.Fatal("expected updated_at not zero")
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})

		t.Run("should return error", func(t *testing.T) {
			repo, mock, err := newImageRepository()
			if err != nil {
				t.Fatal(err)
			}

			mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(0))
			mock.ExpectBegin()
			mock.ExpectExec("INSERT").WillReturnError(sqlmock.ErrCancelled)
			mock.ExpectRollback()

			img, _ := image.New(&image.Config{Tag: "nginx:latest", Description: "description1", MountPath: "/var/www/html"})
			err = repo.Save(img)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})

		t.Run("error if have id but missing createdAt", func(t *testing.T) {
			repo, mock, err := newImageRepository()
			if err != nil {
				t.Fatal(err)
			}

			mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			img, _ := image.New(&image.Config{Tag: "nginx:latest", Description: "description1", MountPath: "/var/www/html"})
			err = repo.Save(img)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})

		t.Run("error if missing id but have createdAt", func(t *testing.T) {
			repo, mock, err := newImageRepository()
			if err != nil {
				t.Fatal(err)
			}

			mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(0))
			img, _ := image.New(&image.Config{Tag: "nginx:latest", Description: "description1", MountPath: "/var/www/html"})
			img.SetCreatedAt(time.Now())
			err = repo.Save(img)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("should delete image", func(t *testing.T) {
			repo, mock, err := newImageRepository()
			if err != nil {
				t.Fatal(err)
			}

			mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id", "tag", "description", "mount_path", "created_at", "updated_at"}).AddRow(1, "nginx:latest", "description1", "/var/www/html", time.Now(), time.Now()))
			mock.ExpectBegin()
			mock.ExpectExec("UPDATE").WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectCommit()

			img, _ := image.New(&image.Config{Tag: "nginx:latest", Description: "description1", MountPath: "/var/www/html"})
			err = repo.Delete(img)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})

		t.Run("should return error", func(t *testing.T) {
			repo, mock, err := newImageRepository()
			if err != nil {
				t.Fatal(err)
			}

			mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id", "tag", "description", "mount_path", "created_at", "updated_at"}).AddRow(1, "nginx:latest", "description1", "/var/www/html", time.Now(), time.Now()))

			mock.ExpectBegin()
			mock.ExpectExec("UPDATE").WillReturnError(sqlmock.ErrCancelled)
			mock.ExpectRollback()

			img, _ := image.New(&image.Config{Tag: "nginx:latest", Description: "description1", MountPath: "/var/www/html"})
			err = repo.Delete(img)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	})
}
