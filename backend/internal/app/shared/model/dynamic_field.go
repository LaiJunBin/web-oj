package model

import "gorm.io/gorm"

type DynamicField struct {
	gorm.Model
	ID           int32 `gorm:"primaryKey,autoIncrement"`
	ProjectID    int32 `gorm:"index"`
	Type         string
	Name         string
	Label        string
	DefaultValue *string
	IsRequired   bool
}

func (DynamicField) TableName() string {
	return "project_dynamic_fields"
}
