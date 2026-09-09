package handler

import (
	"fmt"
	"testing"

	"judge-backend/internal/app/judge_domain/events/judgeupdated"
	"judge-backend/internal/app/judge_domain/events/judgmentcreated"
	"judge-backend/internal/app/judge_domain/events/rejudgerequested"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/gradingitem"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgestatus"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/lifecycle"
	judgmentproject "judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/project"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/rejudge"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/submission"
	"judge-backend/internal/app/judge_domain/judge/domain/factory"
	"judge-backend/internal/app/judge_domain/judge/domain/repository"
	judgemocks "judge-backend/internal/app/judge_domain/judge/mocks"
	sharedlogger "judge-backend/internal/app/shared/infrastructure/logger"
	sharedmocks "judge-backend/internal/app/shared/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// A rejudge over a mixed batch: one submission resolves, one does not. The
// unresolvable one must still get a judgment — a terminal ERROR one attached
// to the rejudge — or a fully-unresolvable rejudge sits in PENDING forever
// (its status only advances on judge-updated events). The ERROR judgment is
// published as JudgeUpdated only AFTER the rejudge is persisted (the same
// ordering rule the queueing path follows), and never as JudgmentCreated.
func TestCreateRejudgeJudgments_UnresolvableSubmission_GetsErrorJudgment(t *testing.T) {
	eventbus := sharedmocks.NewEventBus(t)
	eventRepo := sharedmocks.NewEventRepository(t)
	submissionRepo := judgemocks.NewSubmissionRepository(t)
	judgmentRepo := judgemocks.NewJudgmentRepository(t)
	rejudgeRepo := judgemocks.NewRejudgeRepository(t)
	projectRepo := judgemocks.NewProjectRepository(t)
	gradingItemRepo := judgemocks.NewGradingItemRepository(t)
	lifecycleRepo := judgemocks.NewLifecycleRepository(t)

	newSub := func(id string) *submission.Submission {
		s, err := submission.New(&submission.SubmissionConfig{UserID: 1, ContestProjectID: 10, IsPartial: true, GradingItemIds: []int32{1}})
		assert.NoError(t, err)
		s.SetID(id)
		return s
	}

	submissionRepo.On("FindByID", "sub-good").Return(newSub("sub-good"), nil).Once()
	submissionRepo.On("FindByID", "sub-bad").Return(newSub("sub-bad"), nil).Twice()
	projectRepo.On("FindBySubmissionID", mock.AnythingOfType("string")).Return(judgmentproject.NewProject(&judgmentproject.ProjectConfig{
		ID:           10,
		ImageTag:     "nginx:latest",
		TestImageTag: "test-nginx:latest",
		MountPath:    "/var/www/html",
	}), nil).Times(3)
	lifecycleRepo.On("FindBySubmissionID", mock.AnythingOfType("string")).Return([]*lifecycle.Lifecycle{}, nil).Times(3)
	gradingItemRepo.On("FindBySubmissionID", "sub-good").Return([]*gradingitem.GradingItem{}, nil).Once()
	gradingItemRepo.On("FindBySubmissionID", "sub-bad").
		Return(nil, fmt.Errorf("submission sub-bad: %w", repository.ErrGradingItemsNotOwned)).Once()

	eventRepo.On("TryClaimEvent", "rj-1").Return(true, nil).Once()

	saved := []*judgment.Judgment{}
	judgmentRepo.On("Save", mock.AnythingOfType("*judgment.Judgment")).Return(nil).
		Run(func(args mock.Arguments) { saved = append(saved, args.Get(0).(*judgment.Judgment)) }).Twice()

	rejudgeSaved := false
	var savedRejudge *rejudge.Rejudge
	rejudgeRepo.On("Save", mock.AnythingOfType("*rejudge.Rejudge")).Return(nil).
		Run(func(args mock.Arguments) {
			rejudgeSaved = true
			savedRejudge = args.Get(0).(*rejudge.Rejudge)
		}).Once()

	published := []interface{}{}
	eventbus.On("Publish", mock.Anything).Run(func(args mock.Arguments) {
		// Every publish — the queueing one included — must see the rejudge
		// already persisted, or a fast judgment's status flip is lost.
		assert.True(t, rejudgeSaved, "published %T before the rejudge was persisted", args.Get(0))
		published = append(published, args.Get(0))
	}).Twice()

	h := &RejudgeRequested{
		eventbus:        eventbus,
		eventRepo:       eventRepo,
		judgmentFactory: factory.NewJudgmentFactory(projectRepo, submissionRepo, gradingItemRepo, lifecycleRepo),
		judgmentRepo:    judgmentRepo,
		rejudgeRepo:     rejudgeRepo,
		logger:          sharedlogger.NewMockLogger(),
	}

	h.CreateRejudgeJudgments(rejudgerequested.New("rj-1", 10, []string{"sub-good", "sub-bad"}, "re-run"))

	// Both submissions ended up with a persisted judgment on the rejudge.
	if assert.Len(t, saved, 2) && assert.NotNil(t, savedRejudge) {
		assert.Len(t, savedRejudge.Judgments(), 2)

		var failed *judgment.Judgment
		for _, j := range saved {
			if j.SubmissionID() == "sub-bad" {
				failed = j
			}
		}
		if assert.NotNil(t, failed) {
			assert.Equal(t, judgestatus.ERROR.String(), failed.Status().String())
			assert.Equal(t, factory.GradingItemMismatchReason, *failed.Reason())
			if assert.NotNil(t, failed.RejudgeID()) {
				assert.Equal(t, "rj-1", *failed.RejudgeID())
			}
		}
	}

	// The good judgment is queued; the failed one is only made visible.
	if assert.Len(t, published, 2) {
		assert.IsType(t, &judgmentcreated.JudgmentCreatedEvent{}, published[0])
		assert.IsType(t, &judgeupdated.JudgeUpdatedEvent{}, published[1])
	}
}
