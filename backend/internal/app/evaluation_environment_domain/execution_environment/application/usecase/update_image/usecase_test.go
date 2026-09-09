package updateimage_test

import (
	"archive/zip"
	"bytes"
	"errors"
	"judge-backend/internal/app/shared/test"
	updateimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/update_image"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/batch"
	buildstatus "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/status"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/repository"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExecute(t *testing.T) {
	imageRepo := &mocks.ImageRepository{}
	buildRepo := &mocks.BuildRepository{}
	imageService := &mocks.ImageService{}
	buildService := &mocks.BuildService{}
	usecase := updateimage.NewUsecase(imageRepo, buildRepo, imageService, buildService)

	t.Run("Execute returns error when tag is missing", func(t *testing.T) {
		input := &updateimage.Input{Tag: ""}
		_, err := usecase.Execute(input)
		assert.Equal(t, updateimage.ErrorTagIsRequired, err)
	})

	t.Run("Execute returns error when mount path is invalid", func(t *testing.T) {
		mockImage := image.MustNew(&image.Config{
			Tag:       "nginx:latest",
			MountPath: "/var/www/html",
		})
		imageRepo.On("FindByTag", "nginx:latest").Return(mockImage, nil).Once()
		input := &updateimage.Input{Tag: "nginx:latest", MountPath: test.StringPtr("invalid path")}
		_, err := usecase.Execute(input)
		assert.Equal(t, updateimage.ErrorInvalidMountPath, err)
	})

	t.Run("Execute returns error when image not found", func(t *testing.T) {
		imageRepo.On("FindByTag", "nginx:latest").Return(nil, nil).Once()
		input := &updateimage.Input{Tag: "nginx:latest"}
		_, err := usecase.Execute(input)
		assert.Equal(t, updateimage.ErrorImageNotFound, err)
	})

	t.Run("Execute returns error on internal server error", func(t *testing.T) {
		imageRepo.On("FindByTag", "nginx:latest").Return(nil, errors.New("internal error")).Once()
		input := &updateimage.Input{Tag: "nginx:latest"}
		_, err := usecase.Execute(input)
		assert.Equal(t, updateimage.ErrorInternalServer, err)
	})

	t.Run("Execute returns error when the reservation fails internally", func(t *testing.T) {
		imageEntity := image.MustNew(&image.Config{
			Tag:       "nginx:latest",
			MountPath: "/var/www/html",
		})
		projectZip := new(bytes.Buffer)
		w := zip.NewWriter(projectZip)
		defer w.Close()

		imageRepo.On("FindByTag", "nginx:latest").Return(imageEntity, nil).Once()
		buildRepo.On("FindLastBuild", "nginx:latest").Return(nil, nil).Once()
		buildService.On("ReserveBuild", imageEntity).Return(nil, errors.New("internal error")).Once()
		// No SetupProject expectation: without a reservation the project
		// files must never be touched.

		input := &updateimage.Input{Tag: "nginx:latest", ProjectZipFile: projectZip}

		_, err := usecase.Execute(input)
		assert.Equal(t, updateimage.ErrorInternalServer, err)
	})

	t.Run("Execute returns error when find build error", func(t *testing.T) {
		imageEntity := image.MustNew(&image.Config{
			Tag:       "nginx:latest",
			MountPath: "/var/www/html",
		})
		projectZip := new(bytes.Buffer)
		w := zip.NewWriter(projectZip)
		defer w.Close()

		imageRepo.On("FindByTag", "nginx:latest").Return(imageEntity, nil).Once()
		buildRepo.On("FindLastBuild", "nginx:latest").Return(nil, errors.New("internal error")).Once()

		input := &updateimage.Input{Tag: "nginx:latest", ProjectZipFile: projectZip}
		_, err := usecase.Execute(input)
		assert.Equal(t, updateimage.ErrorInternalServer, err)
	})

	// The DB-visible precheck: a Building last build (seeded or
	// crash-orphaned state the reservation map never saw) rejects before
	// anything else happens — no reservation, no project writes.
	t.Run("Execute returns error when the last build is in progress in the database", func(t *testing.T) {
		imageEntity := image.MustNew(&image.Config{
			Tag:       "nginx:latest",
			MountPath: "/var/www/html",
		})
		buildEntity := build.New(&build.BuildConfig{
			Tag:    imageEntity.Tag(),
			Batch:  batch.MustNew(1),
			Status: buildstatus.BUILDING,
		})
		projectZip := new(bytes.Buffer)
		w := zip.NewWriter(projectZip)
		defer w.Close()

		imageRepo.On("FindByTag", "nginx:latest").Return(imageEntity, nil).Once()
		buildRepo.On("FindLastBuild", "nginx:latest").Return(buildEntity, nil).Once()
		// No ReserveBuild / SetupProject expectations: a call would panic.

		input := &updateimage.Input{Tag: "nginx:latest", ProjectZipFile: projectZip}
		_, err := usecase.Execute(input)
		assert.Equal(t, updateimage.ErrorCurrentBuildInProgress, err)
	})

	// Decisive tests 1+2 of the exclusive-enqueue contract: the loser of
	// the reservation must NOT report success, and — having no reservation —
	// must NEVER touch the project files (no SetupProject expectation is
	// registered, so a call would panic the strict mock). Overwriting the
	// files while the winner's build may already be reading them is exactly
	// what the reservation-first order exists to prevent.
	t.Run("Execute returns error when losing the reservation race and never touches the project", func(t *testing.T) {
		imageEntity := image.MustNew(&image.Config{
			Tag:       "nginx:latest",
			MountPath: "/var/www/html",
		})
		projectZip := new(bytes.Buffer)
		w := zip.NewWriter(projectZip)
		defer w.Close()

		imageRepo.On("FindByTag", "nginx:latest").Return(imageEntity, nil).Once()
		buildRepo.On("FindLastBuild", "nginx:latest").Return(nil, nil).Once()
		buildService.On("ReserveBuild", imageEntity).Return(nil, repository.ErrorBuildAlreadyInProgress).Once()

		input := &updateimage.Input{Tag: "nginx:latest", ProjectZipFile: projectZip}
		_, err := usecase.Execute(input)
		assert.Equal(t, updateimage.ErrorCurrentBuildInProgress, err)
	})

	// Decisive test 3 (usecase half): a setup failure must RELEASE the
	// reservation — without the release the tag stays reserved forever and
	// no later update or rebuild could ever enqueue again. (The repository
	// half — that a released tag can be reserved again — is
	// TestBuildService_SetupFailureReleasesReservation_NextEnqueueSucceeds.)
	// Decisive test 4: no EnqueueReservedBuild expectation is registered, so
	// a job entering the queue despite the failed setup panics the mock.
	t.Run("Execute returns error when setup project error and releases the reservation", func(t *testing.T) {
		imageEntity := image.MustNew(&image.Config{
			Tag:       "nginx:latest",
			MountPath: "/var/www/html",
		})
		reserved := build.New(&build.BuildConfig{
			Tag:    imageEntity.Tag(),
			Batch:  batch.MustNew(2),
			Status: buildstatus.PENDING,
		})
		projectZip := new(bytes.Buffer)
		w := zip.NewWriter(projectZip)
		defer w.Close()

		imageRepo.On("FindByTag", "nginx:latest").Return(imageEntity, nil).Once()
		buildRepo.On("FindLastBuild", "nginx:latest").Return(nil, nil).Once()
		buildService.On("ReserveBuild", imageEntity).Return(reserved, nil).Once()
		imageService.On("SetupProject", imageEntity, projectZip).Return(errors.New("internal error")).Once()
		buildService.On("ReleaseReservedBuild", reserved).Return().Once()

		input := &updateimage.Input{Tag: "nginx:latest", ProjectZipFile: projectZip}
		_, err := usecase.Execute(input)
		assert.Equal(t, updateimage.ErrorFailedToSetup, err)
		// The shared mocks in this file are not strict (no NewBuildService(t)),
		// so an unfulfilled expectation would pass silently — assert the
		// release explicitly.
		buildService.AssertCalled(t, "ReleaseReservedBuild", reserved)
	})

	t.Run("Execute returns error when save image error", func(t *testing.T) {
		imageEntity := image.MustNew(&image.Config{
			Tag:       "nginx:latest",
			MountPath: "/var/www/html",
		})

		imageRepo.On("FindByTag", "nginx:latest").Return(imageEntity, nil).Once()
		imageRepo.On("Save", imageEntity).Return(errors.New("internal error")).Once()

		input := &updateimage.Input{Tag: "nginx:latest"}
		_, err := usecase.Execute(input)
		assert.Equal(t, updateimage.ErrorInternalServer, err)
	})

	t.Run("Execute updates image successfully with new description", func(t *testing.T) {
		imageEntity := image.MustNew(&image.Config{
			Tag:         "nginx:latest",
			MountPath:   "/var/www/html",
			Description: "old description",
		})
		imageRepo.On("FindByTag", "nginx:latest").Return(imageEntity, nil).Once()
		imageRepo.On("Save", imageEntity).Return(nil).Once()

		input := &updateimage.Input{Tag: "nginx:latest"}
		_, err := usecase.Execute(input)
		assert.NoError(t, err)
	})

	t.Run("Execute updates image successfully with new mount path", func(t *testing.T) {
		imageEntity := image.MustNew(&image.Config{
			Tag:         "nginx:latest",
			MountPath:   "/var/www/html",
			Description: "description",
		})

		imageRepo.On("FindByTag", "nginx:latest").Return(imageEntity, nil).Once()
		imageRepo.On("Save", imageEntity).Return(nil).Once()

		input := &updateimage.Input{Tag: "nginx:latest", MountPath: test.StringPtr("/usr/share/nginx/html")}
		_, err := usecase.Execute(input)

		assert.NoError(t, err)
	})

	// Decisive test 4 (positive half): the job may enter the queue only
	// AFTER setup succeeded — the recorded call order proves
	// reserve → setup → enqueue.
	t.Run("Execute updates image successfully with new project zip", func(t *testing.T) {
		imageEntity := image.MustNew(&image.Config{
			Tag:         "nginx:latest",
			MountPath:   "/var/www/html",
			Description: "old description",
		})
		reserved := build.New(&build.BuildConfig{
			Tag:    imageEntity.Tag(),
			Batch:  batch.MustNew(2),
			Status: buildstatus.PENDING,
		})
		projectZip := new(bytes.Buffer)
		w := zip.NewWriter(projectZip)
		defer w.Close()

		var order []string
		imageRepo.On("FindByTag", "nginx:latest").Return(imageEntity, nil).Once()
		buildRepo.On("FindLastBuild", "nginx:latest").Return(nil, nil).Once()
		buildService.On("ReserveBuild", imageEntity).Return(reserved, nil).Run(func(mock.Arguments) {
			order = append(order, "reserve")
		}).Once()
		imageService.On("SetupProject", imageEntity, projectZip).Return(nil).Run(func(mock.Arguments) {
			order = append(order, "setup")
		}).Once()
		buildService.On("EnqueueReservedBuild", reserved, imageEntity, mock.Anything).Return(nil).Run(func(mock.Arguments) {
			order = append(order, "enqueue")
		}).Once()
		imageRepo.On("Save", imageEntity).Return(nil).Once()

		input := &updateimage.Input{Tag: "nginx:latest", ProjectZipFile: projectZip}
		_, err := usecase.Execute(input)
		assert.NoError(t, err)
		assert.Equal(t, []string{"reserve", "setup", "enqueue"}, order,
			"the reservation must be taken before the project files are touched, and the job may only enter the queue after setup succeeded")
	})
}
