package model

import "gorm.io/gorm"

type Build struct {
	gorm.Model
	ID      int `gorm:"primaryKey;autoIncrement"`
	ImageID int `gorm:"index"`
	Batch   int
	Status  int
	Logs    []BuildLog `gorm:"foreignKey:BuildID"`
}
