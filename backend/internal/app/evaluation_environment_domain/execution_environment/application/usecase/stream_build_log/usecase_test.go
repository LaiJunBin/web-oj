package streambuildlog_test

import (
	"errors"
	streambuildlog "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/stream_build_log"
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
	imageRepo := &mocks.ImageRepository{}
	buildRepo := &mocks.BuildRepository{}
	usecase := streambuildlog.NewUsecase(imageRepo, buildRepo)

	t.Run("Execute returns error when tag is missing", func(t *testing.T) {
		input := &streambuildlog.Input{Tag: "", Batch: 1}
		_, err := usecase.Execute(input)
		assert.Equal(t, streambuildlog.ErrorTagIsRequired, err)
	})

	t.Run("Execute returns error when batch is less than or equal to zero", func(t *testing.T) {
		input := &streambuildlog.Input{Tag: "nginx:latest", Batch: 0}
		_, err := usecase.Execute(input)
		assert.Equal(t, streambuildlog.ErrorBatchMustBeGreaterThanZero, err)
	})

	t.Run("Execute returns error when build not found", func(t *testing.T) {
		buildRepo.On("FindByTagAndBatch", "nginx:latest", 1).Return(nil, nil).Once()
		input := &streambuildlog.Input{Tag: "nginx:latest", Batch: 1}
		_, err := usecase.Execute(input)
		assert.Equal(t, streambuildlog.ErrorBuildNotFound, err)
	})

	t.Run("Execute returns error on internal server error", func(t *testing.T) {
		buildRepo.On("FindByTagAndBatch", "nginx:latest", 1).Return(nil, errors.New("internal error")).Once()
		input := &streambuildlog.Input{Tag: "nginx:latest", Batch: 1}
		_, err := usecase.Execute(input)
		assert.Equal(t, streambuildlog.ErrorInternalServer, err)
	})

	t.Run("Execute returns build logs successfully", func(t *testing.T) {
		t.Run("static logs", func(t *testing.T) {
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

			buildRepo.On("FindByTagAndBatch", "nginx:latest", 1).Return(buildEntity, nil).Once()
			input := &streambuildlog.Input{Tag: "nginx:latest", Batch: 1}

			output, err := usecase.Execute(input)
			assert.NoError(t, err)
			assert.NotNil(t, output)
			assert.Len(t, output.Logs(), 2)
			assert.Equal(t, "Build started", output.Logs()[0].Message())
			assert.Equal(t, "Build completed", output.Logs()[1].Message())
		})

		t.Run("streaming logs", func(t *testing.T) {
			buildEntity := build.New(&build.BuildConfig{
				Batch:  batch.MustNew(1),
				Tag:    imagetag.MustNew("nginx:latest"),
				Status: buildstatus.BUILDING,
			})

			buildRepo.On("FindByTagAndBatch", "nginx:latest", 1).Return(buildEntity, nil).Once()
			input := &streambuildlog.Input{Tag: "nginx:latest", Batch: 1}

			output, err := usecase.Execute(input)
			assert.NoError(t, err)
			assert.NotNil(t, output)
			assert.Len(t, output.Logs(), 0)

			buildEntity.AddLog(buildlog.New(buildloglevel.Info, "Build started", time.Now()))
			assert.Len(t, output.Logs(), 1)
			assert.Equal(t, "Build started", output.Logs()[0].Message())

			buildEntity.AddLog(buildlog.New(buildloglevel.Info, "Build completed", time.Now()))
			assert.Len(t, output.Logs(), 2)
			assert.Equal(t, "Build completed", output.Logs()[1].Message())
		})
	})
}
