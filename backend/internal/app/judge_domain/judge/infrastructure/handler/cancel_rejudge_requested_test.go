package handler

import (
	"errors"
	"testing"
	"time"

	"judge-backend/internal/app/judge_domain/events/cancelrejudgerequested"
	"judge-backend/internal/app/judge_domain/events/judgecancelled"
	"judge-backend/internal/app/judge_domain/events/judgeupdated"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgestatus"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgetype"
	judgmentproject "judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/project"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/rejudge"
	judgemocks "judge-backend/internal/app/judge_domain/judge/mocks"
	sharedlogger "judge-backend/internal/app/shared/infrastructure/logger"
	sharedmocks "judge-backend/internal/app/shared/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newCancelRejudgeHandler(t *testing.T) (*CancelRejudgeRequested, *sharedmocks.EventBus, *judgemocks.RejudgeRepository, *judgemocks.JudgmentRepository) {
	eventbus := sharedmocks.NewEventBus(t)
	rejudgeRepo := judgemocks.NewRejudgeRepository(t)
	judgmentRepo := judgemocks.NewJudgmentRepository(t)
	return &CancelRejudgeRequested{
		eventbus:     eventbus,
		rejudgeRepo:  rejudgeRepo,
		judgmentRepo: judgmentRepo,
		logger:       sharedlogger.NewMockLogger(),
	}, eventbus, rejudgeRepo, judgmentRepo
}

// cancelledMember builds the judgment entity FindByID would return for a
// member that was just moved to Cancelled in the database.
func cancelledMember(id string) *judgment.Judgment {
	j := judgment.New(&judgment.JudgmentConfig{
		Project:   judgmentproject.NewProject(&judgmentproject.ProjectConfig{ID: 1}),
		JudgeType: judgetype.REJUDGE,
		CreatedAt: time.Now(),
	})
	j.SetID(id)
	j.SetStatus(judgestatus.CANCELLED)
	return j
}

func expectJudgeCancelledPublish(eventbus *sharedmocks.EventBus, judgmentID string) {
	eventbus.On("Publish", mock.MatchedBy(func(evt *judgecancelled.JudgeCancelledEvent) bool {
		return evt.JudgmentID() == judgmentID
	})).Once()
}

func expectJudgeUpdatedCancelledPublish(eventbus *sharedmocks.EventBus, judgmentID string) {
	eventbus.On("Publish", mock.MatchedBy(func(evt *judgeupdated.JudgeUpdatedEvent) bool {
		return evt.Data().JudgmentID == judgmentID && evt.Data().StatusIsCancelled
	})).Once()
}

// TestCancelRejudge_FindByIDFails_TransitionNeverAttempted is a regression
// test: FindByID is now called BEFORE the atomic status transition, not
// after. If it were the other way around (as it used to be) and FindByID
// failed here, the rejudge would already be sitting in the DB as CANCELED
// with no judge_cancelled events ever published for its judgments — workers
// would keep executing them to completion while the UI claimed the rejudge
// was cancelled. No TransitionStatus expectation is configured below, so an
// unexpected call to it would panic, proving the handler stops before ever
// touching status.
func TestCancelRejudge_FindByIDFails_TransitionNeverAttempted(t *testing.T) {
	h, _, rejudgeRepo, _ := newCancelRejudgeHandler(t)
	e := cancelrejudgerequested.New("evt-1", "rejudge-1")

	rejudgeRepo.On("FindByID", "rejudge-1").Return(nil, errors.New("mongo unavailable")).Once()

	assert.Error(t, h.CancelRejudge(e))
}

func TestCancelRejudge_NotInCancellableState_SkipsWithoutPublishing(t *testing.T) {
	h, _, rejudgeRepo, _ := newCancelRejudgeHandler(t)
	e := cancelrejudgerequested.New("evt-1", "rejudge-1")

	r := rejudge.New(&rejudge.RejudgeConfig{
		ID:        "rejudge-1",
		Status:    rejudge.FINISHED,
		Judgments: []string{"judgment-1"},
	})

	// First read feeds the member list; the second re-reads the CURRENT
	// status after the not-applied transition (FINISHED here, so no
	// reconciliation may run).
	rejudgeRepo.On("FindByID", "rejudge-1").Return(r, nil).Twice()
	rejudgeRepo.On("TransitionStatus", "rejudge-1", []*rejudge.Status{rejudge.PENDING, rejudge.RUNNING}, rejudge.CANCELED).
		Return(false, nil).Once()
	// No eventbus.Publish / TryCancel expectation: an unexpected call would
	// panic, proving the handler stops here when the rejudge was already past
	// the cancellable window (e.g. it just finished). nil: not retryable.

	assert.NoError(t, h.CancelRejudge(e))
}

func TestCancelRejudge_TransitionError_StopsProcessing(t *testing.T) {
	h, _, rejudgeRepo, _ := newCancelRejudgeHandler(t)
	e := cancelrejudgerequested.New("evt-1", "rejudge-1")

	r := rejudge.New(&rejudge.RejudgeConfig{
		ID:        "rejudge-1",
		Status:    rejudge.RUNNING,
		Judgments: []string{"judgment-1"},
	})

	rejudgeRepo.On("FindByID", "rejudge-1").Return(r, nil).Once()
	rejudgeRepo.On("TransitionStatus", "rejudge-1", []*rejudge.Status{rejudge.PENDING, rejudge.RUNNING}, rejudge.CANCELED).
		Return(false, errors.New("mongo unavailable")).Once()

	assert.Error(t, h.CancelRejudge(e))
}

// TestCancelRejudge_Applied_CancelsEveryMemberInDBAndPublishesForWinners is
// the regression test for the "queued members keep executing" bug: the old
// handler only published judge_cancelled per member, which interrupts a
// worker currently executing that member but does nothing for members still
// sitting in the queue — their documents stayed Queued and they later ran to
// completion under a rejudge the user had already cancelled. The handler
// must move every member to Cancelled in the DATABASE (TryCancel), and only
// then publish judge_cancelled (interrupt) plus judgeupdated (rollup and
// contest-stream visibility) for the members it actually transitioned.
func TestCancelRejudge_Applied_CancelsEveryMemberInDBAndPublishesForWinners(t *testing.T) {
	h, eventbus, rejudgeRepo, judgmentRepo := newCancelRejudgeHandler(t)
	e := cancelrejudgerequested.New("evt-1", "rejudge-1")

	r := rejudge.New(&rejudge.RejudgeConfig{
		ID:        "rejudge-1",
		Status:    rejudge.RUNNING,
		Judgments: []string{"judgment-1", "judgment-2"},
	})

	rejudgeRepo.On("FindByID", "rejudge-1").Return(r, nil).Once()
	rejudgeRepo.On("TransitionStatus", "rejudge-1", []*rejudge.Status{rejudge.PENDING, rejudge.RUNNING}, rejudge.CANCELED).
		Return(true, nil).Once()

	for _, id := range []string{"judgment-1", "judgment-2"} {
		// The entity comes from TryCancel's own atomic command — no FindByID
		// expectation is registered, so any re-read would panic.
		judgmentRepo.On("TryCancel", id, "", mock.AnythingOfType("time.Time")).Return(cancelledMember(id), nil).Once()
		expectJudgeCancelledPublish(eventbus, id)
		expectJudgeUpdatedCancelledPublish(eventbus, id)
	}

	assert.NoError(t, h.CancelRejudge(e))
}

// TestCancelRejudge_MemberAlreadyTerminal_NoEventsForThatMember: a member
// that already finished (or was individually cancelled) reports
// applied=false from TryCancel. Publishing judge_cancelled for it anyway
// would burn the judgment-ID-keyed claim in JudgeCancelledHandler for
// nothing, swallowing a later legitimate cancellation of the same judgment
// within the claim TTL — so members this handler did not transition get no
// events at all, while remaining members still get theirs.
func TestCancelRejudge_MemberAlreadyTerminal_NoEventsForThatMember(t *testing.T) {
	h, eventbus, rejudgeRepo, judgmentRepo := newCancelRejudgeHandler(t)
	e := cancelrejudgerequested.New("evt-1", "rejudge-1")

	r := rejudge.New(&rejudge.RejudgeConfig{
		ID:        "rejudge-1",
		Status:    rejudge.RUNNING,
		Judgments: []string{"judgment-done", "judgment-queued"},
	})

	rejudgeRepo.On("FindByID", "rejudge-1").Return(r, nil).Once()
	rejudgeRepo.On("TransitionStatus", "rejudge-1", []*rejudge.Status{rejudge.PENDING, rejudge.RUNNING}, rejudge.CANCELED).
		Return(true, nil).Once()

	judgmentRepo.On("TryCancel", "judgment-done", "", mock.AnythingOfType("time.Time")).Return(nil, nil).Once()
	// No Publish for judgment-done: an unexpected call would panic.

	judgmentRepo.On("TryCancel", "judgment-queued", "", mock.AnythingOfType("time.Time")).Return(cancelledMember("judgment-queued"), nil).Once()
	expectJudgeCancelledPublish(eventbus, "judgment-queued")
	expectJudgeUpdatedCancelledPublish(eventbus, "judgment-queued")

	assert.NoError(t, h.CancelRejudge(e))
}

// TestCancelRejudge_MemberCancelError_ContinuesWithRemainingMembers: one
// member's TryCancel failing (transient Mongo error) must not abandon the
// rest of the batch — and the overall verdict is an ERROR, so the claimed
// subscription retries the whole (idempotent) reconciliation for the
// straggler.
func TestCancelRejudge_MemberCancelError_ContinuesWithRemainingMembers(t *testing.T) {
	h, eventbus, rejudgeRepo, judgmentRepo := newCancelRejudgeHandler(t)
	e := cancelrejudgerequested.New("evt-1", "rejudge-1")

	r := rejudge.New(&rejudge.RejudgeConfig{
		ID:        "rejudge-1",
		Status:    rejudge.RUNNING,
		Judgments: []string{"judgment-broken", "judgment-ok"},
	})

	rejudgeRepo.On("FindByID", "rejudge-1").Return(r, nil).Once()
	rejudgeRepo.On("TransitionStatus", "rejudge-1", []*rejudge.Status{rejudge.PENDING, rejudge.RUNNING}, rejudge.CANCELED).
		Return(true, nil).Once()

	judgmentRepo.On("TryCancel", "judgment-broken", "", mock.AnythingOfType("time.Time")).Return(nil, errors.New("mongo unavailable")).Once()

	judgmentRepo.On("TryCancel", "judgment-ok", "", mock.AnythingOfType("time.Time")).Return(cancelledMember("judgment-ok"), nil).Once()
	expectJudgeCancelledPublish(eventbus, "judgment-ok")
	expectJudgeUpdatedCancelledPublish(eventbus, "judgment-ok")

	assert.Error(t, h.CancelRejudge(e), "a failed member must surface as an error so the retry heals it")
}

// TestCancelRejudge_ParentAlreadyCanceled_StillReconcilesMembers is the P1-2
// regression test: a previous cancellation attempt can flip the parent to
// CANCELED and then die partway through its members, leaving Queued children
// under a CANCELED parent. The old handler treated a not-applied parent
// transition as "nothing to do" and returned — so a re-received cancel
// request could never heal the stragglers. It must instead re-read the
// parent and, when it is CANCELED, run member reconciliation anyway.
func TestCancelRejudge_ParentAlreadyCanceled_StillReconcilesMembers(t *testing.T) {
	h, eventbus, rejudgeRepo, judgmentRepo := newCancelRejudgeHandler(t)
	e := cancelrejudgerequested.New("evt-1", "rejudge-1")

	// The stale first read still shows RUNNING; the transition then reports
	// not-applied because a previous attempt already flipped it.
	stale := rejudge.New(&rejudge.RejudgeConfig{
		ID:        "rejudge-1",
		Status:    rejudge.RUNNING,
		Judgments: []string{"judgment-straggler", "judgment-done"},
	})
	current := rejudge.New(&rejudge.RejudgeConfig{
		ID:        "rejudge-1",
		Status:    rejudge.CANCELED,
		Judgments: []string{"judgment-straggler", "judgment-done"},
	})

	rejudgeRepo.On("FindByID", "rejudge-1").Return(stale, nil).Once()
	rejudgeRepo.On("TransitionStatus", "rejudge-1", []*rejudge.Status{rejudge.PENDING, rejudge.RUNNING}, rejudge.CANCELED).
		Return(false, nil).Once()
	rejudgeRepo.On("FindByID", "rejudge-1").Return(current, nil).Once()

	// The straggler still gets cancelled and announced; the member a prior
	// attempt (or the worker) already finished reports (nil, nil) and stays
	// silent.
	judgmentRepo.On("TryCancel", "judgment-straggler", "", mock.AnythingOfType("time.Time")).Return(cancelledMember("judgment-straggler"), nil).Once()
	expectJudgeCancelledPublish(eventbus, "judgment-straggler")
	expectJudgeUpdatedCancelledPublish(eventbus, "judgment-straggler")
	judgmentRepo.On("TryCancel", "judgment-done", "", mock.AnythingOfType("time.Time")).Return(nil, nil).Once()

	assert.NoError(t, h.CancelRejudge(e))
}
