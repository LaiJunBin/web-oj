package deleteimage_test

import (
	"errors"
	deleteimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/delete_image"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExecute(t *testing.T) {
	repo := &mocks.ImageRepository{}
	service := &mocks.ImageService{}
	cancelBuildUsecase := &mocks.Usecase{}
	usecase := deleteimage.NewUsecase(repo, service, cancelBuildUsecase)

	t.Run("Execute returns error when tag is missing", func(t *testing.T) {
		input := &deleteimage.Input{Tag: ""}
		_, err := usecase.Execute(input)
		assert.Equal(t, deleteimage.ErrorTagIsRequired, err)
	})

	t.Run("Execute returns error when image not found", func(t *testing.T) {
		repo.On("FindByTag", "nginx:latest").Return(nil, nil).Once()
		input := &deleteimage.Input{Tag: "nginx:latest"}
		_, err := usecase.Execute(input)
		assert.Equal(t, deleteimage.ErrorImageNotFound, err)
		repo.AssertExpectations(t)
	})

	t.Run("Execute returns error on internal server error", func(t *testing.T) {
		repo.On("FindByTag", "nginx:latest").Return(nil, errors.New("internal error")).Once()
		input := &deleteimage.Input{Tag: "nginx:latest"}
		_, err := usecase.Execute(input)
		assert.Equal(t, deleteimage.ErrorInternalServer, err)
		repo.AssertExpectations(t)
	})

	t.Run("Execute deletes image successfully", func(t *testing.T) {
		imageEntity := image.MustNew(&image.Config{
			Tag:         "nginx:latest",
			Description: "description1",
			MountPath:   "/var/www/html",
		})
		repo.On("FindByTag", "nginx:latest").Return(imageEntity, nil).Once()
		repo.On("Delete", imageEntity).Return(nil).Once()
		service.On("CleanupProject", imageEntity).Return(nil).Once()
		cancelBuildUsecase.On("Execute", mock.Anything).Return(nil, nil).Once()

		input := &deleteimage.Input{Tag: "nginx:latest"}
		output, err := usecase.Execute(input)
		assert.NoError(t, err)
		assert.NotNil(t, output)
		repo.AssertExpectations(t)
		service.AssertExpectations(t)
		cancelBuildUsecase.AssertExpectations(t)
	})

	t.Run("Execute fails when cleanup project fails, and does not touch the row", func(t *testing.T) {
		// Behaviour change with the files-first ordering: a failed
		// cleanup used to be swallowed and the call still answered success
		// with the row already gone. Now the files are removed BEFORE the
		// row, so a cleanup failure aborts the delete: no row delete, no
		// build cancel, and the caller is told the delete did not happen.
		imageEntity := image.MustNew(&image.Config{
			Tag:         "nginx:latest",
			Description: "description1",
			MountPath:   "/var/www/html",
		})
		repo.On("FindByTag", "nginx:latest").Return(imageEntity, nil).Once()
		service.On("CleanupProject", imageEntity).Return(errors.New("cleanup error")).Once()

		input := &deleteimage.Input{Tag: "nginx:latest"}
		output, err := usecase.Execute(input)
		assert.ErrorIs(t, err, deleteimage.ErrorInternalServer)
		assert.Nil(t, output)
		repo.AssertExpectations(t)
		service.AssertExpectations(t)
		cancelBuildUsecase.AssertExpectations(t)
	})

	t.Run("Execute logs error when cancel build fails", func(t *testing.T) {
		imageEntity := image.MustNew(&image.Config{
			Tag:         "nginx:latest",
			Description: "description1",
			MountPath:   "/var/www/html",
		})
		repo.On("FindByTag", "nginx:latest").Return(imageEntity, nil).Once()
		repo.On("Delete", imageEntity).Return(nil).Once()
		service.On("CleanupProject", imageEntity).Return(nil).Once()
		cancelBuildUsecase.On("Execute", mock.Anything).Return(nil, errors.New("cancel build error")).Once()

		input := &deleteimage.Input{Tag: "nginx:latest"}
		output, err := usecase.Execute(input)
		assert.NoError(t, err)
		assert.NotNil(t, output)
		repo.AssertExpectations(t)
		service.AssertExpectations(t)
		cancelBuildUsecase.AssertExpectations(t)
	})
}
