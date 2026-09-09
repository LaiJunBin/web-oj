package rebuild_test

import (
	"errors"
	rebuild "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/rebuild"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/batch"
	buildstatus "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/status"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/repository"
	buildservice "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/service/build_service"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	imageRepo := &mocks.ImageRepository{}
	buildRepo := &mocks.BuildRepository{}
	service := &mocks.BuildService{}
	usecase := rebuild.NewUsecase(imageRepo, buildRepo, service)

	t.Run("Execute returns error when tag is missing", func(t *testing.T) {
		input := &rebuild.Input{Tag: ""}
		_, err := usecase.Execute(input)
		assert.Equal(t, rebuild.ErrorTagIsRequired, err)
	})

	t.Run("Execute returns error when image not found", func(t *testing.T) {
		imageRepo.On("FindByTag", "nginx:latest").Return(nil, nil).Once()
		input := &rebuild.Input{Tag: "nginx:latest"}
		_, err := usecase.Execute(input)
		assert.Equal(t, rebuild.ErrorImageNotFound, err)
	})

	t.Run("Execute returns error on find image internal server error", func(t *testing.T) {
		imageRepo.On("FindByTag", "nginx:latest").Return(nil, errors.New("internal error")).Once()
		input := &rebuild.Input{Tag: "nginx:latest"}
		_, err := usecase.Execute(input)
		assert.Equal(t, rebuild.ErrorInternalServer, err)
	})

	t.Run("Execute returns error on find build internal server error", func(t *testing.T) {
		imageEntity := image.MustNew(&image.Config{
			Tag:       "nginx:latest",
			MountPath: "/var/www/html",
		})
		imageRepo.On("FindByTag", "nginx:latest").Return(imageEntity, nil).Once()
		buildRepo.On("FindLastBuild", "nginx:latest").Return(nil, errors.New("internal error")).Once()
		input := &rebuild.Input{Tag: "nginx:latest"}
		_, err := usecase.Execute(input)
		assert.Equal(t, rebuild.ErrorInternalServer, err)
	})

	t.Run("Execute returns error when build is in progress", func(t *testing.T) {
		imageEntity := image.MustNew(&image.Config{
			Tag:       "nginx:latest",
			MountPath: "/var/www/html",
		})
		buildEntity := build.New(&build.BuildConfig{
			Batch:  batch.MustNew(1),
			Status: buildstatus.BUILDING,
		})
		imageRepo.On("FindByTag", "nginx:latest").Return(imageEntity, nil).Once()
		buildRepo.On("FindLastBuild", "nginx:latest").Return(buildEntity, nil).Once()
		input := &rebuild.Input{Tag: "nginx:latest"}
		_, err := usecase.Execute(input)
		assert.Equal(t, rebuild.ErrorBuildInProgress, err)
	})

	t.Run("Execute return error on image project not set", func(t *testing.T) {
		imageEntity := image.MustNew(&image.Config{
			Tag:       "nginx:latest",
			MountPath: "/var/www/html",
		})
		buildEntity := build.New(&build.BuildConfig{
			Batch:  batch.MustNew(1),
			Status: buildstatus.SUCCESS,
		})
		imageRepo.On("FindByTag", "nginx:latest").Return(imageEntity, nil).Once()
		buildRepo.On("FindLastBuild", "nginx:latest").Return(buildEntity, nil).Once()
		service.On("EnqueueBuildJob", imageEntity, &buildservice.BuildOptions{
			NoCache: true,
		}).Return(buildservice.ErrorImageProjectNotSet).Once()
		input := &rebuild.Input{Tag: "nginx:latest"}
		_, err := usecase.Execute(input)
		assert.Equal(t, rebuild.ErrorImageProjectNotSet, err)
	})

	t.Run("Execute rebuilds build successfully on build not in progress", func(t *testing.T) {
		imageEntity := image.MustNew(&image.Config{
			Tag:       "nginx:latest",
			MountPath: "/var/www/html",
		})
		buildEntity := build.New(&build.BuildConfig{
			Batch:  batch.MustNew(1),
			Status: buildstatus.SUCCESS,
		})
		imageRepo.On("FindByTag", "nginx:latest").Return(imageEntity, nil).Once()
		buildRepo.On("FindLastBuild", "nginx:latest").Return(buildEntity, nil).Once()
		service.On("EnqueueBuildJob", imageEntity, &buildservice.BuildOptions{
			NoCache: true,
		}).Return(nil).Once()

		input := &rebuild.Input{Tag: "nginx:latest"}
		output, err := usecase.Execute(input)
		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	// Regression: the exclusive-enqueue loser must not report success. The
	// old code logged the reservation conflict and returned 200 — the user
	// believed a rebuild was running that never existed.
	t.Run("Execute returns ErrorBuildInProgress when the reservation is already taken", func(t *testing.T) {
		imageEntity := image.MustNew(&image.Config{
			Tag:       "nginx:latest",
			MountPath: "/var/www/html",
		})
		imageRepo.On("FindByTag", "nginx:latest").Return(imageEntity, nil).Once()
		buildRepo.On("FindLastBuild", "nginx:latest").Return(nil, nil).Once()
		service.On("EnqueueBuildJob", imageEntity, &buildservice.BuildOptions{
			NoCache: true,
		}).Return(repository.ErrorBuildAlreadyInProgress).Once()

		input := &rebuild.Input{Tag: "nginx:latest"}
		_, err := usecase.Execute(input)
		assert.Equal(t, rebuild.ErrorBuildInProgress, err)
	})

	t.Run("Execute returns internal error on any other enqueue failure", func(t *testing.T) {
		imageEntity := image.MustNew(&image.Config{
			Tag:       "nginx:latest",
			MountPath: "/var/www/html",
		})
		imageRepo.On("FindByTag", "nginx:latest").Return(imageEntity, nil).Once()
		buildRepo.On("FindLastBuild", "nginx:latest").Return(nil, nil).Once()
		service.On("EnqueueBuildJob", imageEntity, &buildservice.BuildOptions{
			NoCache: true,
		}).Return(errors.New("factory exploded")).Once()

		input := &rebuild.Input{Tag: "nginx:latest"}
		_, err := usecase.Execute(input)
		assert.Equal(t, rebuild.ErrorInternalServer, err)
	})

	t.Run("Execute rebuilds build successfully on build not found", func(t *testing.T) {
		imageEntity := image.MustNew(&image.Config{
			Tag:       "nginx:latest",
			MountPath: "/var/www/html",
		})
		imageRepo.On("FindByTag", "nginx:latest").Return(imageEntity, nil).Once()
		buildRepo.On("FindLastBuild", "nginx:latest").Return(nil, nil).Once()
		service.On("EnqueueBuildJob", imageEntity, &buildservice.BuildOptions{
			NoCache: true,
		}).Return(nil).Once()

		input := &rebuild.Input{Tag: "nginx:latest"}
		output, err := usecase.Execute(input)
		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

}
