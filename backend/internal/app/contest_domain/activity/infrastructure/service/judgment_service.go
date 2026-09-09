package service

import (
	"judge-backend/internal/app/contest_domain/activity/domain/service"
	sharedevent "judge-backend/internal/app/shared/events"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type JudgmentService struct {
	eventBus sharedtypes.EventBus
}

func NewJudgmentService(eventBus sharedtypes.EventBus) service.JudgmentService {
	return &JudgmentService{
		eventBus: eventBus,
	}
}

func (s *JudgmentService) RequestCancelJudgeByID(id string, reason string) {
	s.eventBus.Publish(sharedevent.NewCancelJudgeRequested(id, reason))
}

func (s *JudgmentService) RequestUpdateStaffNoteByID(id string, staffNote *string) {
	s.eventBus.Publish(sharedevent.NewStaffNoteUpdateRequested(id, staffNote))
}
