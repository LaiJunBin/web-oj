// Package migration brings the database up to the shape the models describe.
//
// There is deliberately no version table and no per-version up/down. This
// system is deployed once per contest venue against an empty database and
// torn down afterwards, so the schema only ever has to reach its current
// shape — never to walk an upgrade path across releases of live data. Every
// migration is therefore a plain AutoMigrate off the model, and RunMigrate is
// safe to re-run because AutoMigrate is additive.
//
// The cost of that choice, so nobody discovers it mid-incident: AutoMigrate
// can add tables, columns and indexes, but it cannot rename a column, drop a
// column, or transform existing rows. If this ever has to upgrade a database
// that already holds real data, that is the point to introduce a versioned
// migration tool rather than to extend this.
package migration

import (
	"fmt"
	"judge-backend/database/migration/migrations"

	"gorm.io/gorm"
)

type Migrator interface {
	Up(db *gorm.DB) error
	Down(db *gorm.DB) error
}

var (
	Migrations = []Migrator{
		migrations.UserMigrator{},
		migrations.GroupMigrator{},
		migrations.ImageMigrator{},
		migrations.BuildMigrator{},
		migrations.BuildLogMigrator{},
		migrations.ProjectMigrator{},
		migrations.ProjectAssetMigrator{},
		migrations.ProjectLifecycleMigrator{},
		migrations.ProjectDynamicFieldMigrator{},
		migrations.GradingItemMigrator{},
		migrations.TestcaseMigrator{},
		migrations.GradingItemDependMigrator{},
		migrations.TestcaseDependMigrator{},
		migrations.ContestParticipantMigrator{},
		migrations.ContestMigrator{},
		migrations.ContestProjectMigrator{},
		migrations.ClarificationMigrator{},
	}
)

func RunMigrate(db *gorm.DB) {
	fmt.Println("Migrating...")
	for _, m := range Migrations {
		if err := m.Up(db); err != nil {
			panic(err)
		}
	}
	fmt.Println("Migration done")
}

func RunRollback(db *gorm.DB) {
	fmt.Println("Rolling back...")
	for _, m := range Migrations {
		if err := m.Down(db); err != nil {
			panic(err)
		}
	}
	fmt.Println("Rollback done")
}
