package service

import (
	"judge-backend/internal/app/contest_domain/activity/domain/entity/rejudge"
	"judge-backend/internal/app/contest_domain/activity/domain/service"
	sharedevent "judge-backend/internal/app/shared/events"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type RejudgeService struct {
	eventBus        sharedtypes.EventBus
	judgmentService service.JudgmentService
}

func NewRejudgeService(
	eventBus sharedtypes.EventBus,
	judgmentService service.JudgmentService,
) service.RejudgeService {
	return &RejudgeService{
		eventBus:        eventBus,
		judgmentService: judgmentService,
	}
}

func (s *RejudgeService) RequestApproveRejudge(rejudge *rejudge.Rejudge) {
	s.eventBus.Publish(sharedevent.NewRejudgeApprovalRequested(rejudge.ID()))
}

func (s *RejudgeService) RequestRejectRejudge(rejudge *rejudge.Rejudge) {
	s.eventBus.Publish(sharedevent.NewRejudgeRejectionRequested(rejudge.ID()))
}

func (s *RejudgeService) RequestCancelRejudge(rejudge *rejudge.Rejudge) {
	s.eventBus.Publish(sharedevent.NewCancelRejudgeRequested(rejudge.ID()))
}
