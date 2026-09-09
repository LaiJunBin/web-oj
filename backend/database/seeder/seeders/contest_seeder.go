package seeders

import (
	"fmt"
	"time"

	"judge-backend/internal/app/shared/model"

	"gorm.io/gorm"
)

type ContestSeeder struct{}

func (s ContestSeeder) Seed(db *gorm.DB) error {
	data, err := LoadDemoData()
	if err != nil {
		return err
	}

	// Look the project up by name instead of assuming it landed on ID 1 —
	// the demo dataset should survive someone adding another seeder before
	// this one.
	var project model.Project
	if err := db.Where("name = ?", data.Project.Name).First(&project).Error; err != nil {
		return fmt.Errorf("demo project %q not seeded: %w", data.Project.Name, err)
	}

	// One base timestamp for the whole window: the contest must not activate
	// a second after the project it contains.
	now := time.Now()
	contestData := data.Contest

	contest := &model.Contest{
		Name:           contestData.Name,
		Description:    contestData.Description,
		ActivateTime:   After(now, contestData.ActivateAfterMinutes),
		DeactivateTime: After(now, contestData.DeactivateAfterMinutes),
		IsPublic:       contestData.IsPublic,
		Projects: []model.ContestProject{
			{
				ProjectID:             project.ID,
				Name:                  contestData.Project.Name,
				ActivateTime:          After(now, contestData.Project.ActivateAfterMinutes),
				DeactivateTime:        After(now, contestData.Project.DeactivateAfterMinutes),
				StartTime:             After(now, contestData.Project.StartAfterMinutes),
				EndTime:               After(now, contestData.Project.EndAfterMinutes),
				IsPublic:              contestData.Project.IsPublic,
				JudgePriorityStrategy: contestData.Project.JudgePriorityStrategy,
			},
		},
	}

	if err := db.Create(contest).Error; err != nil {
		return err
	}

	return s.enrollDemoUsers(db, data, contest.ID)
}

// enrollDemoUsers signs every seeded account up for the demo contest —
// including the admin, so whoever installs the system can submit and watch a
// judgment run without first creating a second account.
func (s ContestSeeder) enrollDemoUsers(db *gorm.DB, data *DemoData, contestID int32) error {
	usernames := data.Usernames()

	var users []model.User
	if err := db.Where("username IN ?", usernames).Order("id").Find(&users).Error; err != nil {
		return err
	}

	if len(users) != len(usernames) {
		return fmt.Errorf("expected %d demo users to enroll, found %d", len(usernames), len(users))
	}

	for i := range users {
		userID := users[i].ID
		if err := db.Create(&model.ContestParticipant{
			ContestID: contestID,
			UserID:    &userID,
		}).Error; err != nil {
			return err
		}
	}

	return nil
}
