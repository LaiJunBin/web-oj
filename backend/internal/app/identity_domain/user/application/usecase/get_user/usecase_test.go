package getuser_test

import (
	"errors"
	"judge-backend/internal/app/shared/infrastructure/logger"
	"judge-backend/internal/app/identity_domain/user/application/service/query/userquery"
	getuser "judge-backend/internal/app/identity_domain/user/application/usecase/get_user"
	"judge-backend/internal/app/identity_domain/user/mocks"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUsecase_Execute(t *testing.T) {
	mockService := mocks.NewUserQueryService(t)
	logger := logger.NewMockLogger()
	usecase := getuser.NewUsecase(mockService, logger)

	t.Run("success", func(t *testing.T) {
		expectedUser := &userquery.User{
			ID:        1,
			Username:  "user1",
			Name:      "User 1",
			Roles:     []string{"admin"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		mockService.On("GetUserByID", int32(1)).Return(expectedUser, nil).Once()

		output, err := usecase.Execute(&getuser.Input{ID: 1})

		assert.NoError(t, err)
		assert.Equal(t, expectedUser, output.User)
	})

	t.Run("user not found", func(t *testing.T) {
		mockService.On("GetUserByID", int32(1)).Return(nil, userquery.ErrorUserNotFound).Once()

		output, err := usecase.Execute(&getuser.Input{ID: 1})

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.Equal(t, getuser.ErrorUserNotFound, err)
	})

	t.Run("internal server error", func(t *testing.T) {
		mockService.On("GetUserByID", int32(1)).Return(nil, errors.New("some error")).Once()

		output, err := usecase.Execute(&getuser.Input{ID: 1})

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.Equal(t, getuser.ErrorInternalServer, err)
	})

	t.Run("invalid ID", func(t *testing.T) {
		output, err := usecase.Execute(&getuser.Input{ID: 0})

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.Equal(t, getuser.ErrorInvalidID, err)
	})
}
