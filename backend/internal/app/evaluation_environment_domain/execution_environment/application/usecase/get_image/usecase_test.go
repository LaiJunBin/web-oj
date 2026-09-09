package getimage_test

import (
	"errors"
	getimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/get_image"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	repo := &mocks.ImageRepository{}
	usecase := getimage.NewUsecase(repo)

	t.Run("Execute returns error when tag is missing", func(t *testing.T) {
		input := &getimage.Input{Tag: ""}
		_, err := usecase.Execute(input)
		assert.Equal(t, getimage.ErrorTagIsRequired, err)
	})

	t.Run("Execute returns error when image not found", func(t *testing.T) {
		repo.On("FindByTag", "nginx:latest").Return(nil, nil).Once()
		input := &getimage.Input{Tag: "nginx:latest"}
		_, err := usecase.Execute(input)
		assert.Equal(t, getimage.ErrorImageNotFound, err)
	})

	t.Run("Execute returns error on internal server error", func(t *testing.T) {
		repo.On("FindByTag", "nginx:latest").Return(nil, errors.New("internal error")).Once()
		input := &getimage.Input{Tag: "nginx:latest"}
		_, err := usecase.Execute(input)
		assert.Equal(t, getimage.ErrorInternalServer, err)
	})

	t.Run("Execute returns image info successfully", func(t *testing.T) {
		imageEntity := image.MustNew(&image.Config{
			Tag:         "nginx:latest",
			Description: "description1",
			MountPath:   "/var/www/html",
		})
		repo.On("FindByTag", "nginx:latest").Return(imageEntity, nil).Once()

		input := &getimage.Input{Tag: "nginx:latest"}
		output, err := usecase.Execute(input)
		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, "nginx:latest", output.Tag)
		assert.Equal(t, "description1", output.Description)
		assert.Equal(t, "/var/www/html", output.MountPath)
		assert.Equal(t, imageEntity.CreatedAt(), output.CreatedAt)
		assert.Equal(t, imageEntity.UpdatedAt(), output.UpdatedAt)
	})
}
