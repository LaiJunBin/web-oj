package model

import "gorm.io/gorm"

type Asset struct {
	gorm.Model
	ID          string `gorm:"primaryKey"`
	ProjectID   int32  `gorm:"index"`
	Name        string
	Description string
	Size        int64
	IsEnabled   bool
}

func (Asset) TableName() string {
	return "project_assets"
}
