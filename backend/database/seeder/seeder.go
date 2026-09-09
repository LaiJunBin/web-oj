package seeder

import (
	"fmt"
	"judge-backend/database/seeder/seeders"
	"judge-backend/internal/config"

	"gorm.io/gorm"
)

type Seeder interface {
	Seed(db *gorm.DB) error
}

// RunSeed writes the built-in demonstration dataset — one admin, a handful of
// demo contestants, and one live contest with a gradeable nginx exercise. Its
// content lives in database/seeder/seeders/demo_data.json; the seeders here
// only decide the order it is written in.
//
// cfg is accepted for venue-local seeders that need application config
// (e.g. grading env composed from docker/sonar settings); none of the
// built-in seeders reads it — the demo dataset is self-contained.
func RunSeed(db *gorm.DB, cfg *config.Config) {
	fmt.Println("Seeding...")
	fmt.Println("[seed] WARNING: seeding the built-in DEMO dataset with well-known credentials (admin, demo01–05). For demonstration/development only — never expose this to an untrusted network.")
	for _, s := range []Seeder{
		seeders.UserSeeder{},
		seeders.ImageSeeder{},
		seeders.ProjectSeeder{},
		seeders.ContestSeeder{},
	} {
		if err := s.Seed(db); err != nil {
			panic(err)
		}
	}
	fmt.Println("Seed done")
}
