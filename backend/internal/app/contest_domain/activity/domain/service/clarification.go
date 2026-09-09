package service

import "judge-backend/internal/app/contest_domain/activity/domain/entity/clarification"

type ClarificationService interface {
	NotifyClarificationToUser(clarification *clarification.Clarification)
}
