package model

import "gorm.io/gorm"

type Testcase struct {
	gorm.Model
	ID            int32 `gorm:"primaryKey,autoIncrement"`
	GradingItemID int32
	Name          string
	Command       string
	Score        float64
	IsSample bool

	// relations
	Depends     []*Testcase  `gorm:"many2many:testcase_depends;"`
	GradingItem *GradingItem `gorm:"foreignKey:GradingItemID"`
}

type TestcaseDepend struct {
	TestcaseID int32 `gorm:"primaryKey"`
	DependID   int32 `gorm:"primaryKey"`
}
