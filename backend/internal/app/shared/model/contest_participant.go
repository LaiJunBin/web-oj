package model

type ContestParticipant struct {
	ID        uint  `gorm:"primaryKey;autoIncrement"`
	ContestID int32 `gorm:"not null;uniqueIndex:udx_contest_user;uniqueIndex:udx_contest_group"`
	UserID    *int  `gorm:"uniqueIndex:udx_contest_user"`
	GroupID   *int  `gorm:"uniqueIndex:udx_contest_group"`
}
