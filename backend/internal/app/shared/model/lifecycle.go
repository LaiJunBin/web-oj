package model

import "gorm.io/gorm"

type Lifecycle struct {
	gorm.Model
	ID          int32 `gorm:"primaryKey;autoIncrement"`
	ProjectID   int32 `gorm:"index"`
	Name        string
	Description string
	Type        string
	RunsOn      string
	Command     string
}

func (Lifecycle) TableName() string {
	return "project_lifecycles"
}
