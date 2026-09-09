package listbuild_test

import (
	"errors"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/service/query/build_query/models/buildinfo"
	listbuild "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/list_build"
	buildstatus "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/status"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	repo := &mocks.ImageRepository{}
	service := &mocks.BuildQueryService{}
	usecase := listbuild.NewUsecase(repo, service)

	t.Run("Execute returns error when tag is missing", func(t *testing.T) {
		input := &listbuild.Input{Tag: ""}
		_, err := usecase.Execute(input)
		assert.Equal(t, listbuild.ErrorTagIsRequired, err)
	})

	t.Run("Execute returns error when image not found", func(t *testing.T) {
		repo.On("FindByTag", "nginx:latest").Return(nil, nil).Once()
		input := &listbuild.Input{Tag: "nginx:latest"}
		_, err := usecase.Execute(input)
		assert.Equal(t, listbuild.ErrorImageNotFound, err)
	})

	t.Run("Execute returns error on internal server error", func(t *testing.T) {
		repo.On("FindByTag", "nginx:latest").Return(nil, errors.New("internal error")).Once()
		input := &listbuild.Input{Tag: "nginx:latest"}
		_, err := usecase.Execute(input)
		assert.Equal(t, listbuild.ErrorInternalServer, err)
	})

	t.Run("Execute returns build list successfully", func(t *testing.T) {
		image := image.MustNew(&image.Config{
			Tag:       "nginx:latest",
			MountPath: "/var/www/html",
		})
		buildInfos := []buildinfo.BuildInfo{
			{
				Batch:  1,
				Status: buildstatus.SUCCESS.String(),
			},
			{
				Batch:  2,
				Status: buildstatus.ERROR.String(),
			},
		}

		repo.On("FindByTag", "nginx:latest").Return(image, nil).Once()
		service.On("GetBuilds", image.Tag().String()).Return(buildInfos, nil).Once()

		input := &listbuild.Input{Tag: "nginx:latest"}
		output, err := usecase.Execute(input)
		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, buildInfos, output.Builds)
	})
}
