package factory_test

import (
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/factory"
	imagetag "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image/valueobject/tag"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/mocks"
	"judge-backend/internal/projectpath"
	"testing"
)

func TestImageFactory(t *testing.T) {
	storagePath := "/path/to/storage"
	storagePathProvider := &mocks.StoragePathProvider{}
	storagePathProvider.On("GetStoragePath").Return(storagePath)
	factory := factory.NewImageFactory(storagePathProvider)

	t.Run("CreateImage with valid inputs", func(t *testing.T) {
		t.Run("non-setup project", func(t *testing.T) {
			tag := "nginx:latest"
			mountPath := "/var/www/html"
			description := "Nginx image"
			img, err := factory.CreateImage(tag, mountPath, description)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if img.Tag().String() != tag {
				t.Errorf("Expected tag to be %s, got %s", tag, img.Tag())
			}

			if img.Description() != description {
				t.Errorf("Expected description to be %s, got %s", description, img.Description())
			}

			expectedStoragePath := projectpath.Root + "/path/to/storage/images/nginx:latest"
			if img.Project().StoragePath() != expectedStoragePath {
				t.Errorf("Expected storage path to be %s, got %s", expectedStoragePath, img.Project().StoragePath())
			}

			if img.IsReady() {
				t.Errorf("Expected image to not be ready, got ready")
			}
		})
	})

	t.Run("CreateImage with invalid tag", func(t *testing.T) {
		tag := "invalid_tag"
		mountPath := "/var/www/html"
		description := "Invalid image"
		img, err := factory.CreateImage(tag, mountPath, description)
		if err == nil {
			t.Fatalf("Expected error, got nil")
		}

		if img != nil {
			t.Errorf("Expected image to be nil, got %v", img)
		}
	})

	t.Run("Create Image with invalid mount path", func(t *testing.T) {
		tag := "nginx:latest"
		mountPath := ""
		description := "Nginx image"
		img, err := factory.CreateImage(tag, mountPath, description)
		if err == nil {
			t.Fatalf("Expected error, got nil")
		}

		if img != nil {
			t.Errorf("Expected image to be nil, got %v", img)
		}
	})
}

// The rename-on-delete namespace is reserved at the creation choke point:
// a live image with a marker tag could collide with a dead row's renamed tag
// (blocking that row's delete forever).
func TestImageFactory_RejectsReservedDeletedMarkerTags(t *testing.T) {
	storagePathProvider := &mocks.StoragePathProvider{}
	storagePathProvider.On("GetStoragePath").Return("/path/to/storage")
	f := factory.NewImageFactory(storagePathProvider)

	for _, tag := range []string{"evil__deleted_1:latest", "nginx:latest__deleted_12"} {
		if _, err := f.CreateImage(tag, "/var/www/html", "d"); err != imagetag.ErrInvalidTagFormat {
			t.Errorf("CreateImage(%q) error = %v, want ErrInvalidTagFormat", tag, err)
		}
	}
}
