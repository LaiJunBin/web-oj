package model

import (
	"gorm.io/gorm"
)

type Project struct {
	gorm.Model
	ID           int32 `gorm:"primaryKey;autoIncrement"`
	Name         string
	Description  string
	ImageTag     *string `gorm:"index"`
	Env          string
	TestImageTag *string `gorm:"index"`
	TestEnv      string
	ResultParser string

	// relations
	Image         *Image `gorm:"foreignKey:ImageTag;references:Tag"`
	TestImage     *Image `gorm:"foreignKey:TestImageTag;references:Tag"`
	Assets        []Asset
	DynamicFields []DynamicField
	GradingItems  []GradingItem
	Lifecycles    []Lifecycle
}
