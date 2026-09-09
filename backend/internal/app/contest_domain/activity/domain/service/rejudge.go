package service

import (
	"judge-backend/internal/app/contest_domain/activity/domain/entity/rejudge"
)

type RejudgeService interface {
	RequestApproveRejudge(rejudge *rejudge.Rejudge)
	RequestRejectRejudge(rejudge *rejudge.Rejudge)
	RequestCancelRejudge(rejudge *rejudge.Rejudge)
}
