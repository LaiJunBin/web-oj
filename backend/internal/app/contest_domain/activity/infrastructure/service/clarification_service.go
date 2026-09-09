package service

import (
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/contest_domain/activity/domain/entity/clarification"
	"judge-backend/internal/app/contest_domain/events/clarificationnotificationrequested"
	"judge-backend/internal/app/contest_domain/activity/domain/service"
)

type ClarificationService struct {
	eventBus sharedtypes.EventBus
}

func NewClarificationService(eventBus sharedtypes.EventBus) service.ClarificationService {
	return &ClarificationService{
		eventBus: eventBus,
	}
}

func (s *ClarificationService) NotifyClarificationToUser(clarification *clarification.Clarification) {
	s.eventBus.Publish(clarificationnotificationrequested.New(clarification))
}
