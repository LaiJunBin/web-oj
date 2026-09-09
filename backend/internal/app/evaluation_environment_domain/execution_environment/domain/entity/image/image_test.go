package image_test

import (
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image"
	"testing"
)

func TestImage(t *testing.T) {
	t.Run("Test Image Creation", func(t *testing.T) {
		t.Run("New Image", func(t *testing.T) {
			config := image.Config{
				Tag:         "nginx:latest",
				Description: "Nginx image",
				MountPath:   "/var/www/html",
			}
			image, err := image.New(&config)
			if err != nil {
				t.Errorf("Error creating image: %s", err)
			}

			if image.Tag().String() != "nginx:latest" {
				t.Errorf("Expected tag to be 'nginx:latest', got %s", image.Tag())
			}

			if image.Description() != "Nginx image" {
				t.Errorf("Expected description to be 'Nginx image', got %s", image.Description())
			}

			if image.MountPath().String() != "/var/www/html" {
				t.Errorf("Expected mount path to be '/var/www/html', got %s", image.MountPath())
			}
		})

		t.Run("Error when creating image with invalid tag", func(t *testing.T) {
			config := image.Config{
				Tag:         "nginx",
				Description: "Nginx image",
				MountPath:   "/var/www/html",
			}

			image, err := image.New(&config)
			if err == nil {
				t.Errorf("Expected error, got nil")
			}

			if image != nil {
				t.Errorf("Expected image to be nil, got %v", image)
			}
		})

		t.Run("Error when creating image with empty tag", func(t *testing.T) {
			config := image.Config{
				Tag:         "",
				Description: "Nginx image",
				MountPath:   "/var/www/html",
			}

			image, err := image.New(&config)
			if err == nil {
				t.Errorf("Expected error, got nil")
			}

			if image != nil {
				t.Errorf("Expected image to be nil, got %v", image)
			}
		})

		t.Run("Error when creating image with invalid mount path", func(t *testing.T) {
			config := image.Config{
				Tag:         "nginx:latest",
				Description: "Nginx image",
				MountPath:   "invalid path",
			}

			image, err := image.New(&config)
			if err == nil {
				t.Errorf("Expected error, got nil")
			}

			if image != nil {
				t.Errorf("Expected image to be nil, got %v", image)
			}
		})
	})
}
