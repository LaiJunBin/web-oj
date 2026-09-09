package dto

import "time"

type Clarification struct {
	ID               int32
	Subject          string
	Content          string
	IsAdminBroadcast bool
	IsAdminResponse  bool
	CreatedAt        time.Time
}

type ClarificationDetail struct {
	Clarification
	ReplyClarification *ClarificationDetail
}
