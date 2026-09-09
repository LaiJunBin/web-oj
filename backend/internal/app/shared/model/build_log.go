package model

import (
	"time"
)

type BuildLog struct {
	ID        int `gorm:"primaryKey;autoIncrement"`
	BuildID   int `gorm:"index"`
	Level     int
	Message   string
	CreatedAt time.Time
}
