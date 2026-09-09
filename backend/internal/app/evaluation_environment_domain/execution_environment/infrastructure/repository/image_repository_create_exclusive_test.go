package repository_test

import (
	"errors"
	"testing"

	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image"
	domainrepo "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/repository"

	"github.com/DATA-DOG/go-sqlmock"
	mysqldriver "github.com/go-sql-driver/mysql"
)

func newImageForCreate(t *testing.T) *image.Image {
	t.Helper()
	img, err := image.New(&image.Config{
		Tag:         "create-exclusive:latest",
		MountPath:   "/var/www/html",
		Description: "create exclusive test",
		StoragePath: "storage",
	})
	if err != nil {
		t.Fatalf("creating image entity: %v", err)
	}
	return img
}

// TestImageRepository_CreateExclusive pins the method's whole point: it
// INSERTS only (never falls back to updating an existing row the way Save's
// read-then-upsert does), and the database's unique-index rejection maps to
// the unambiguous ErrImageTagAlreadyExists — so a concurrent create's loser
// is told it lost instead of silently overwriting the winner.
func TestImageRepository_CreateExclusive(t *testing.T) {
	t.Run("inserts a new row", func(t *testing.T) {
		repo, sqlMock, err := newImageRepository()
		if err != nil {
			t.Fatalf("creating repository: %v", err)
		}

		sqlMock.ExpectBegin()
		sqlMock.ExpectExec("INSERT INTO `images`").
			WillReturnResult(sqlmock.NewResult(1, 1))
		sqlMock.ExpectCommit()

		token, err := repo.CreateExclusive(newImageForCreate(t))
		if err != nil {
			t.Fatalf("expected insert to succeed, got %v", err)
		}
		if token == 0 {
			t.Fatalf("expected a non-zero ownership token for the inserted row")
		}
		if err := sqlMock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet SQL expectations (an UPDATE instead of an INSERT would land here): %v", err)
		}
	})

	t.Run("maps a duplicate-key rejection to ErrImageTagAlreadyExists", func(t *testing.T) {
		repo, sqlMock, err := newImageRepository()
		if err != nil {
			t.Fatalf("creating repository: %v", err)
		}

		sqlMock.ExpectBegin()
		sqlMock.ExpectExec("INSERT INTO `images`").
			WillReturnError(&mysqldriver.MySQLError{Number: 1062, Message: "Duplicate entry 'create-exclusive:latest' for key 'tag'"})
		sqlMock.ExpectRollback()

		_, err = repo.CreateExclusive(newImageForCreate(t))
		if !errors.Is(err, domainrepo.ErrImageTagAlreadyExists) {
			t.Fatalf("expected ErrImageTagAlreadyExists, got %v", err)
		}
	})

	t.Run("any other database error passes through untranslated", func(t *testing.T) {
		repo, sqlMock, err := newImageRepository()
		if err != nil {
			t.Fatalf("creating repository: %v", err)
		}

		sqlMock.ExpectBegin()
		sqlMock.ExpectExec("INSERT INTO `images`").
			WillReturnError(&mysqldriver.MySQLError{Number: 1213, Message: "Deadlock found"})
		sqlMock.ExpectRollback()

		_, err = repo.CreateExclusive(newImageForCreate(t))
		if err == nil || errors.Is(err, domainrepo.ErrImageTagAlreadyExists) {
			t.Fatalf("a non-duplicate failure must not masquerade as a duplicate, got %v", err)
		}
	})
}
