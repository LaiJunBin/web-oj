package rolequery

import (
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/identity_domain/user/application/service/query/rolequery"
	"judge-backend/internal/app/identity_domain/user/domain/entity/user/valueobject/role"
)

type RoleQueryService struct {
	logger sharedtypes.Logger
}

func NewRoleQueryService(logger sharedtypes.Logger) rolequery.RoleQueryService {
	return &RoleQueryService{
		logger: logger,
	}
}

func (s *RoleQueryService) ListRole() ([]rolequery.Role, error) {
	s.logger.Infof("Starting to list all roles")

	roles := []rolequery.Role{}
	for _, r := range role.GetRoles() {
		roles = append(roles, rolequery.Role{
			ID:   int32(r.ID()),
			Name: r.Name(),
		})
	}

	s.logger.Infof("Successfully listed all roles", "count", len(roles))
	return roles, nil
}
