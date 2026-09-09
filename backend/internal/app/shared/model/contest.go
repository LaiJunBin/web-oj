package model

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Contest struct {
	gorm.Model
	ID             int32 `gorm:"primaryKey;autoIncrement"`
	Name           string
	Description    string
	ActivateTime   time.Time
	DeactivateTime time.Time
	IsPublic       bool

	// relations
	Projects []ContestProject
}

func (c *Contest) AfterSave(tx *gorm.DB) (err error) {
	if len(c.Projects) > 0 {
		ids := make([]int32, 0)
		for _, p := range c.Projects {
			ids = append(ids, p.ID)
		}
		x := tx.Delete(&ContestProject{}, "contest_id = ? AND id NOT IN ?", c.ID, ids)
		if err := x.Error; err != nil {
			tx.Rollback()
			return err
		}

		for i := range c.Projects {
			c.Projects[i].ContestID = c.ID
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"contest_id", "project_id", "name", "activate_time", "deactivate_time", "start_time", "end_time", "is_public", "judge_priority_strategy", "max_judge_duration_minutes", "submission_cooldown_seconds", "pids_limit", "memory_limit_mb", "scoreboard_compact", "scoreboard_mask_names"}),
		}).Save(&c.Projects).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		if err := tx.Delete(&ContestProject{}, "contest_id = ?", c.ID).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return nil
}

func (c *Contest) IsActivated() bool {
	return c.ActivateTime.Before(time.Now()) && c.DeactivateTime.After(time.Now())
}
