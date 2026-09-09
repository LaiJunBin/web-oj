package cancelbuild_test

import (
	"errors"
	cancelbuild "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/cancel_build"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/batch"
	buildstatus "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/status"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	repo := &mocks.BuildRepository{}
	service := &mocks.BuildService{}
	usecase := cancelbuild.NewUsecase(repo, service)

	t.Run("Execute returns error when tag is missing", func(t *testing.T) {
		input := &cancelbuild.Input{Tag: ""}
		_, err := usecase.Execute(input)
		assert.Equal(t, cancelbuild.ErrorInvalidInput, err)
	})

	t.Run("Execute returns error when build not found", func(t *testing.T) {
		repo.On("FindLastBuild", "nginx:latest").Return(nil, nil).Once()
		input := &cancelbuild.Input{Tag: "nginx:latest"}
		_, err := usecase.Execute(input)
		assert.Equal(t, cancelbuild.ErrorBuildNotFound, err)
	})

	t.Run("Execute returns error on internal server error", func(t *testing.T) {
		repo.On("FindLastBuild", "nginx:latest").Return(nil, errors.New("internal error")).Once()
		input := &cancelbuild.Input{Tag: "nginx:latest"}
		_, err := usecase.Execute(input)
		assert.Equal(t, cancelbuild.ErrorInternalServer, err)
	})

	t.Run("Execute returns error when build is not in progress", func(t *testing.T) {
		buildEntity := build.New(&build.BuildConfig{
			Batch:  batch.MustNew(1),
			Status: buildstatus.SUCCESS,
		})
		repo.On("FindLastBuild", "nginx:latest").Return(buildEntity, nil).Once()
		input := &cancelbuild.Input{Tag: "nginx:latest"}
		_, err := usecase.Execute(input)
		assert.Equal(t, cancelbuild.ErrorBuildNotInProgress, err)
	})

	t.Run("Execute cancels build successfully when build is in progress", func(t *testing.T) {
		buildEntity := build.New(&build.BuildConfig{
			Batch:  batch.MustNew(1),
			Status: buildstatus.BUILDING,
		})
		repo.On("FindLastBuild", "nginx:latest").Return(buildEntity, nil).Once()
		service.On("Cancel", buildEntity).Once()

		input := &cancelbuild.Input{Tag: "nginx:latest"}
		output, err := usecase.Execute(input)
		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("Execute cancels build successfully when build is in pending", func(t *testing.T) {
		buildEntity := build.New(&build.BuildConfig{
			Batch:  batch.MustNew(1),
			Status: buildstatus.PENDING,
		})
		repo.On("FindLastBuild", "nginx:latest").Return(buildEntity, nil).Once()
		service.On("Cancel", buildEntity).Once()

		input := &cancelbuild.Input{Tag: "nginx:latest"}
		output, err := usecase.Execute(input)
		assert.NoError(t, err)
		assert.NotNil(t, output)
	})
}
