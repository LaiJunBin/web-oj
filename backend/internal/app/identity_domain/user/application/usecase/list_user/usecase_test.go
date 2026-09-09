package listuser_test

import (
	"errors"
	"judge-backend/internal/app/shared/infrastructure/logger"
	"judge-backend/internal/app/identity_domain/user/application/service/query/userquery"
	listuser "judge-backend/internal/app/identity_domain/user/application/usecase/list_user"
	"judge-backend/internal/app/identity_domain/user/mocks"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUsecase_Execute(t *testing.T) {
	mockService := mocks.NewUserQueryService(t)
	logger := logger.NewMockLogger()
	usecase := listuser.NewUsecase(mockService, logger)

	t.Run("success", func(t *testing.T) {
		expectedUsers := []userquery.User{
			{
				ID:        1,
				Username:  "user1",
				Name:      "User 1",
				Roles:     []string{"admin"},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}
		mockService.On("ListUser").Return(expectedUsers, nil).Once()

		output, err := usecase.Execute(&listuser.Input{})

		assert.NoError(t, err)
		assert.Equal(t, expectedUsers, output.Users)
	})

	t.Run("internal server error", func(t *testing.T) {
		mockService.On("ListUser").Return(nil, errors.New("some error")).Once()

		output, err := usecase.Execute(&listuser.Input{})

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.Equal(t, listuser.ErrorInternalServer, err)
		mockService.AssertExpectations(t)
	})
}
