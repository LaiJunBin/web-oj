package getbuild_test

import (
	"errors"
	getbuild "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/get_build"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/batch"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/buildloglevel"
	buildlog "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/log"
	buildstatus "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/status"
	imagetag "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image/valueobject/tag"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/mocks"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	repo := &mocks.BuildRepository{}
	usecase := getbuild.NewUsecase(repo)

	t.Run("Execute returns error when tag is missing", func(t *testing.T) {
		input := &getbuild.Input{Tag: "", Batch: 1}
		_, err := usecase.Execute(input)
		assert.Equal(t, getbuild.ErrorTagIsRequired, err)
	})

	t.Run("Execute returns error when batch is less than or equal to zero", func(t *testing.T) {
		input := &getbuild.Input{Tag: "nginx:latest", Batch: 0}
		_, err := usecase.Execute(input)
		assert.Equal(t, getbuild.ErrorBatchMustBeGreaterThanZero, err)
	})

	t.Run("Execute returns error when build not found", func(t *testing.T) {
		repo.On("FindByTagAndBatch", "nginx:latest", 1).Return(nil, nil).Once()
		input := &getbuild.Input{Tag: "nginx:latest", Batch: 1}
		_, err := usecase.Execute(input)
		assert.Equal(t, getbuild.ErrorBuildNotFound, err)
	})

	t.Run("Execute returns error on internal server error", func(t *testing.T) {
		repo.On("FindByTagAndBatch", "nginx:latest", 1).Return(nil, errors.New("internal error")).Once()
		input := &getbuild.Input{Tag: "nginx:latest", Batch: 1}
		_, err := usecase.Execute(input)
		assert.Equal(t, getbuild.ErrorInternalServer, err)
	})

	t.Run("Execute returns build info successfully", func(t *testing.T) {
		buildLogs := []buildlog.Log{
			buildlog.New(buildloglevel.Info, "Build started", time.Now()),
			buildlog.New(buildloglevel.Info, "Build completed", time.Now()),
		}
		buildEntity := build.New(&build.BuildConfig{
			Batch:  batch.MustNew(1),
			Tag:    imagetag.MustNew("nginx:latest"),
			Status: buildstatus.SUCCESS,
		})
		buildEntity.AddLog(buildLogs[0])
		buildEntity.AddLog(buildLogs[1])

		repo.On("FindByTagAndBatch", "nginx:latest", 1).Return(buildEntity, nil).Once()
		input := &getbuild.Input{Tag: "nginx:latest", Batch: 1}
		output, err := usecase.Execute(input)
		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, int32(1), output.Batch)
		assert.Equal(t, "SUCCESS", output.Status)
		assert.Len(t, output.Logs, 2)
		assert.Equal(t, "Build started", output.Logs[0].Message)
		assert.Equal(t, "Build completed", output.Logs[1].Message)
	})
}
