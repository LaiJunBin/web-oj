package model

import "gorm.io/gorm"

type GradingItem struct {
	gorm.Model
	ID              int32 `gorm:"primaryKey,autoIncrement"`
	ProjectID       int32 `gorm:"index"`
	Name            string
	Description     string
	Score           float64
	Order           uint32
	ScoringStrategy string
	OutputFormat    string
	ResultParser    string
	Label           string

	// relations
	Testcases []*Testcase `gorm:"ForeignKey:GradingItemID"`
	Depends   []*Testcase `gorm:"many2many:grading_item_depends;"`
}

type GradingItemDepend struct {
	GradingItemID int32 `gorm:"primaryKey"`
	TestcaseID    int32 `gorm:"primaryKey"`
}
