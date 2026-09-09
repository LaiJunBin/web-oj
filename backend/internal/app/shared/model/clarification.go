package model

import (
	"gorm.io/gorm"
)

type Clarification struct {
	gorm.Model
	ID                   int32 `gorm:"primaryKey;autoIncrement"`
	ContestProjectID     int32
	Subject              string
	Content              string
	SenderUserID         *int
	ReceiverUserID       *int
	SendToAll            bool
	IsAnswered           bool
	ReplyClarificationID *int32

	// relations
	ContestProject     *ContestProject `gorm:"foreignKey:ContestProjectID"`
	ReplyClarification *Clarification  `gorm:"foreignKey:ReplyClarificationID"`
	Sender             *User           `gorm:"foreignKey:SenderUserID"`
	Receiver           *User           `gorm:"foreignKey:ReceiverUserID"`
}
