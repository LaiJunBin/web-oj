package listrole_test

import (
	"errors"
	"judge-backend/internal/app/shared/infrastructure/logger"
	"judge-backend/internal/app/identity_domain/user/application/service/query/rolequery"
	listrole "judge-backend/internal/app/identity_domain/user/application/usecase/list_role"
	"judge-backend/internal/app/identity_domain/user/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsecase_Execute(t *testing.T) {
	mockService := mocks.NewRoleQueryService(t)
	logger := logger.NewMockLogger()
	usecase := listrole.NewUsecase(mockService, logger)

	t.Run("success", func(t *testing.T) {
		expectedRoles := []rolequery.Role{
			{
				ID:   1,
				Name: "admin",
			},
			{
				ID:   2,
				Name: "contestant",
			},
		}
		mockService.On("ListRole").Return(expectedRoles, nil).Once()

		output, err := usecase.Execute(&listrole.Input{})

		assert.NoError(t, err)
		assert.Equal(t, expectedRoles, output.Roles)
	})

	t.Run("internal server error", func(t *testing.T) {
		mockService.On("ListRole").Return(nil, errors.New("some error")).Once()

		output, err := usecase.Execute(&listrole.Input{})

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.Equal(t, listrole.ErrorInternalServer, err)
	})
}
