package dto

import "time"

var UnknownClarificationUserInfo = ClarificationUserInfo{
	ID:       -1,
	Name:     "Unknown",
	Username: "unknown",
}

type ClarificationUserInfo struct {
	ID       int32
	Name     string
	Username string
}

type Clarification struct {
	ID           int32
	Subject      string
	Content      string
	SenderUser   ClarificationUserInfo
	ReceiverUser *ClarificationUserInfo
	IsAdmin      bool
	SendToAll    bool
	IsAnswered   bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type ClarificationDetail struct {
	*Clarification
	ReplyClarification *ClarificationDetail
}
