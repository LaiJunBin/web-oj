package userquery

import (
	"judge-backend/internal/app/shared/model"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/identity_domain/user/application/service/query/userquery"
	"judge-backend/internal/app/identity_domain/user/application/service/user/roleconverter"

	"gorm.io/gorm"
)

type UserQueryService struct {
	db            *gorm.DB
	roleConverter roleconverter.RoleConverter
	logger        sharedtypes.Logger
}

func NewUserQueryService(
	db *gorm.DB,
	roleConverter roleconverter.RoleConverter,
	logger sharedtypes.Logger,
) userquery.UserQueryService {
	return &UserQueryService{
		db:            db,
		roleConverter: roleConverter,
		logger:        logger,
	}
}

func (uqs *UserQueryService) ListUser() ([]userquery.User, error) {
	uqs.logger.Debugf("Starting to list all users")

	users := []model.User{}
	if err := uqs.db.Find(&users).Error; err != nil {
		uqs.logger.Errorf("Error listing users", "error", err)
		return nil, err
	}

	output := make([]userquery.User, 0, len(users))
	for _, u := range users {
		roles := []string{}
		for _, role := range uqs.roleConverter.ToRoles(u.Roles) {
			roles = append(roles, role.Name())
		}
		output = append(output, userquery.User{
			ID:        int32(u.ID),
			Name:      u.Name,
			Username:  u.Username,
			Roles:     roles,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
		})
	}

	uqs.logger.Debugf("Successfully listed all users", "count", len(output))
	return output, nil
}

func (uqs *UserQueryService) GetUserByID(id int32) (*userquery.User, error) {
	uqs.logger.Debugf("Starting to get user by ID", "id", id)

	u := model.User{}
	if err := uqs.db.First(&u, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			uqs.logger.Debugf("User not found", "id", id)
			return nil, userquery.ErrorUserNotFound
		}
		uqs.logger.Errorf("Error getting user by ID", "id", id, "error", err)
		return nil, err
	}

	roles := []string{}
	for _, role := range uqs.roleConverter.ToRoles(u.Roles) {
		roles = append(roles, role.Name())
	}

	uqs.logger.Debugf("Successfully got user by ID", "id", id)
	return &userquery.User{
		ID:        int32(u.ID),
		Name:      u.Name,
		Username:  u.Username,
		Roles:     roles,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}, nil
}
