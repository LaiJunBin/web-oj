package rolequery

import (
	"testing"

	"judge-backend/internal/app/shared/infrastructure/logger"
	"judge-backend/internal/app/identity_domain/user/domain/entity/user/valueobject/role"

	"github.com/stretchr/testify/assert"
)

func TestRoleQueryService_ListRole(t *testing.T) {
	logger := logger.NewMockLogger()
	service := NewRoleQueryService(logger)

	t.Run("ListRole returns all roles", func(t *testing.T) {
		roles, err := service.ListRole()
		assert.NoError(t, err)
		assert.NotNil(t, roles)

		expectedRoles := role.GetRoles()
		assert.Len(t, roles, len(expectedRoles))

		for i, expectedRole := range expectedRoles {
			assert.Equal(t, int32(expectedRole.ID()), roles[i].ID)
			assert.Equal(t, expectedRole.Name(), roles[i].Name)
		}
	})
}
