package model

import "gorm.io/gorm"

type Group struct {
	gorm.Model
	ID          int `gorm:"primaryKey;autoIncrement"`
	Name        string
	Description string
}

type UserGroup struct {
	UserID  int `gorm:"primaryKey"`
	GroupID int `gorm:"primaryKey"`
}
