package seeders

import (
	"judge-backend/internal/app/identity_domain/user/domain/entity/user/valueobject/role"
	"judge-backend/internal/app/identity_domain/user/infrastructure/service/user/roleconverter"
	"judge-backend/internal/app/shared/infrastructure/logger"
	hashservice "judge-backend/internal/app/shared/infrastructure/services/hash_service"
	"judge-backend/internal/app/shared/model"

	"gorm.io/gorm"
)

type UserSeeder struct{}

func (s UserSeeder) Seed(db *gorm.DB) error {
	data, err := LoadDemoData()
	if err != nil {
		return err
	}

	logger := logger.NewMockLogger()
	hashService := hashservice.NewHashService(logger)
	roleConverter := roleconverter.NewRoleConverter(logger)

	// bcrypt is deliberately slow and every demo account shares one password,
	// so hash it once instead of once per account — the sqlite dev server
	// reseeds on every boot.
	password, err := hashService.Hash(data.Password)
	if err != nil {
		return err
	}

	for _, u := range data.Users() {
		roles, err := role.GetRolesByNames(u.Roles)
		if err != nil {
			return err
		}

		if err := db.Create(&model.User{
			Name:     u.Name,
			Username: u.Username,
			Password: password,
			Roles:    roleConverter.ToBit(roles),
		}).Error; err != nil {
			return err
		}
	}

	return nil
}
