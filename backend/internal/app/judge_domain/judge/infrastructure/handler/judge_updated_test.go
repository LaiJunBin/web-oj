package handler

import (
	"errors"
	"testing"

	"judge-backend/internal/app/judge_domain/events/judgeupdated"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgetype"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/rejudge"
	judgemocks "judge-backend/internal/app/judge_domain/judge/mocks"
	sharedeventbus "judge-backend/internal/app/shared/infrastructure/eventbus"
	sharedlogger "judge-backend/internal/app/shared/infrastructure/logger"

	"github.com/stretchr/testify/assert"
)

// newRejudgeJudgment builds a JUDGE-type judgment attached to a rejudge and
// drives it through the guarded state transitions to the given terminal
// state, mirroring how execute_judgment actually finishes a judgment.
func newRejudgeJudgment(t *testing.T, rejudgeID string, terminal func(*judgment.Judgment)) *judgment.Judgment {
	t.Helper()
	j := judgment.New(&judgment.JudgmentConfig{JudgeType: judgetype.REJUDGE})
	j.SetID("judgment-1")
	j.SetRejudgeID(rejudgeID)
	if terminal != nil {
		terminal(j)
	}
	return j
}

func TestUpdateRejudgeStatus_NoRejudgeID_IsNoOp(t *testing.T) {
	repo := judgemocks.NewRejudgeRepository(t) // no .On(...) registered: any call fails the test
	h := &JudgeUpdatedHandler{rejudgeRepo: repo, logger: sharedlogger.NewMockLogger()}

	j := judgment.New(&judgment.JudgmentConfig{JudgeType: judgetype.JUDGE})
	j.SetID("judgment-1")
	// no SetRejudgeID call: RejudgeID() stays nil

	h.UpdateRejudgeStatus(judgeupdated.NewJudgeUpdatedEvent(j))
}

func TestUpdateRejudgeStatus_MidRunUpdate_AttemptsRunningOnly(t *testing.T) {
	repo := judgemocks.NewRejudgeRepository(t)
	h := &JudgeUpdatedHandler{rejudgeRepo: repo, logger: sharedlogger.NewMockLogger()}

	require := func(err error) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	j := newRejudgeJudgment(t, "rejudge-1", func(j *judgment.Judgment) { require(j.Start()) })

	repo.On("TransitionStatus", "rejudge-1", []*rejudge.Status{rejudge.PENDING}, rejudge.RUNNING).
		Return(true, nil).Once()

	h.UpdateRejudgeStatus(judgeupdated.NewJudgeUpdatedEvent(j))
}

func TestUpdateRejudgeStatus_ErrorTerminal_TransitionsToError(t *testing.T) {
	repo := judgemocks.NewRejudgeRepository(t)
	h := &JudgeUpdatedHandler{rejudgeRepo: repo, logger: sharedlogger.NewMockLogger()}

	j := newRejudgeJudgment(t, "rejudge-1", func(j *judgment.Judgment) {
		_ = j.Start()
		_ = j.Fail()
	})

	notYetTerminal := []*rejudge.Status{rejudge.PENDING, rejudge.RUNNING}
	repo.On("TransitionStatus", "rejudge-1", []*rejudge.Status{rejudge.PENDING}, rejudge.RUNNING).Return(false, nil).Once()
	repo.On("TransitionStatus", "rejudge-1", notYetTerminal, rejudge.ERROR).Return(true, nil).Once()

	h.UpdateRejudgeStatus(judgeupdated.NewJudgeUpdatedEvent(j))
}

func TestUpdateRejudgeStatus_CancelledTerminal_TransitionsToCanceled(t *testing.T) {
	repo := judgemocks.NewRejudgeRepository(t)
	h := &JudgeUpdatedHandler{rejudgeRepo: repo, logger: sharedlogger.NewMockLogger()}

	j := newRejudgeJudgment(t, "rejudge-1", func(j *judgment.Judgment) {
		_ = j.Start()
		_ = j.Cancel("")
	})

	notYetTerminal := []*rejudge.Status{rejudge.PENDING, rejudge.RUNNING}
	repo.On("TransitionStatus", "rejudge-1", []*rejudge.Status{rejudge.PENDING}, rejudge.RUNNING).Return(false, nil).Once()
	repo.On("TransitionStatus", "rejudge-1", notYetTerminal, rejudge.CANCELED).Return(true, nil).Once()

	h.UpdateRejudgeStatus(judgeupdated.NewJudgeUpdatedEvent(j))
}

func TestUpdateRejudgeStatus_RejudgedAndAllFinished_TransitionsToFinished(t *testing.T) {
	repo := judgemocks.NewRejudgeRepository(t)
	h := &JudgeUpdatedHandler{rejudgeRepo: repo, logger: sharedlogger.NewMockLogger()}

	j := newRejudgeJudgment(t, "rejudge-1", func(j *judgment.Judgment) {
		_ = j.Start()
		_ = j.Finish()
	})

	notYetTerminal := []*rejudge.Status{rejudge.PENDING, rejudge.RUNNING}
	repo.On("TransitionStatus", "rejudge-1", []*rejudge.Status{rejudge.PENDING}, rejudge.RUNNING).Return(false, nil).Once()
	repo.On("AllJudgmentsFinished", "rejudge-1").Return(true, nil).Once()
	repo.On("TransitionStatus", "rejudge-1", notYetTerminal, rejudge.FINISHED).Return(true, nil).Once()

	h.UpdateRejudgeStatus(judgeupdated.NewJudgeUpdatedEvent(j))
}

// TestUpdateRejudgeStatus_RejudgedButNotAllFinished is a regression test for
// the specific race this handler exists to avoid: the terminal REJUDGED
// event of one judgment in a multi-judgment rejudge must NOT flip the batch
// to FINISHED while sibling judgments are still running.
func TestUpdateRejudgeStatus_RejudgedButNotAllFinished_DoesNotTransitionToFinished(t *testing.T) {
	repo := judgemocks.NewRejudgeRepository(t)
	h := &JudgeUpdatedHandler{rejudgeRepo: repo, logger: sharedlogger.NewMockLogger()}

	j := newRejudgeJudgment(t, "rejudge-1", func(j *judgment.Judgment) {
		_ = j.Start()
		_ = j.Finish()
	})

	repo.On("TransitionStatus", "rejudge-1", []*rejudge.Status{rejudge.PENDING}, rejudge.RUNNING).Return(false, nil).Once()
	repo.On("AllJudgmentsFinished", "rejudge-1").Return(false, nil).Once()
	// no TransitionStatus(..., FINISHED) expectation registered: the mock's
	// AssertExpectations would still pass since we never asserted it must be
	// called, but the absence of a `.On` means an unexpected call panics —
	// so calling it here proves the handler correctly does NOT invoke it.

	h.UpdateRejudgeStatus(judgeupdated.NewJudgeUpdatedEvent(j))
}

func TestUpdateRejudgeStatus_TransitionError_StopsProcessing(t *testing.T) {
	repo := judgemocks.NewRejudgeRepository(t)
	h := &JudgeUpdatedHandler{rejudgeRepo: repo, logger: sharedlogger.NewMockLogger()}

	j := newRejudgeJudgment(t, "rejudge-1", func(j *judgment.Judgment) {
		_ = j.Start()
		_ = j.Fail()
	})

	repo.On("TransitionStatus", "rejudge-1", []*rejudge.Status{rejudge.PENDING}, rejudge.RUNNING).
		Return(false, errors.New("mongo unavailable")).Once()
	// No further mock expectations: if the handler kept going after the
	// first TransitionStatus error instead of returning, the ERROR
	// transition call below would panic on an unmocked call. The returned
	// error is what hands the event to the retried subscription.

	assert.Error(t, h.UpdateRejudgeStatus(judgeupdated.NewJudgeUpdatedEvent(j)))
}

// Every failure branch must surface as an error so the retried subscription
// can re-run it — each of these used to be a logged dead end that could
// permanently strand the rejudge (most visibly in RUNNING).
func TestUpdateRejudgeStatus_TerminalTransitionErrors_ReturnErrors(t *testing.T) {
	t.Run("ERROR transition failure", func(t *testing.T) {
		repo := judgemocks.NewRejudgeRepository(t)
		h := &JudgeUpdatedHandler{rejudgeRepo: repo, logger: sharedlogger.NewMockLogger()}
		j := newRejudgeJudgment(t, "rejudge-1", func(j *judgment.Judgment) {
			_ = j.Start()
			_ = j.Fail()
		})
		repo.On("TransitionStatus", "rejudge-1", []*rejudge.Status{rejudge.PENDING}, rejudge.RUNNING).
			Return(false, nil).Once()
		repo.On("TransitionStatus", "rejudge-1", []*rejudge.Status{rejudge.PENDING, rejudge.RUNNING}, rejudge.ERROR).
			Return(false, errors.New("mongo unavailable")).Once()
		assert.Error(t, h.UpdateRejudgeStatus(judgeupdated.NewJudgeUpdatedEvent(j)))
	})

	t.Run("AllJudgmentsFinished failure", func(t *testing.T) {
		repo := judgemocks.NewRejudgeRepository(t)
		h := &JudgeUpdatedHandler{rejudgeRepo: repo, logger: sharedlogger.NewMockLogger()}
		j := newRejudgeJudgment(t, "rejudge-1", func(j *judgment.Judgment) {
			_ = j.Start()
			_ = j.Finish()
		})
		repo.On("TransitionStatus", "rejudge-1", []*rejudge.Status{rejudge.PENDING}, rejudge.RUNNING).
			Return(false, nil).Once()
		repo.On("AllJudgmentsFinished", "rejudge-1").
			Return(false, errors.New("mongo unavailable")).Once()
		assert.Error(t, h.UpdateRejudgeStatus(judgeupdated.NewJudgeUpdatedEvent(j)))
	})

	t.Run("FINISHED transition failure", func(t *testing.T) {
		repo := judgemocks.NewRejudgeRepository(t)
		h := &JudgeUpdatedHandler{rejudgeRepo: repo, logger: sharedlogger.NewMockLogger()}
		j := newRejudgeJudgment(t, "rejudge-1", func(j *judgment.Judgment) {
			_ = j.Start()
			_ = j.Finish()
		})
		repo.On("TransitionStatus", "rejudge-1", []*rejudge.Status{rejudge.PENDING}, rejudge.RUNNING).
			Return(false, nil).Once()
		repo.On("AllJudgmentsFinished", "rejudge-1").Return(true, nil).Once()
		repo.On("TransitionStatus", "rejudge-1", []*rejudge.Status{rejudge.PENDING, rejudge.RUNNING}, rejudge.FINISHED).
			Return(false, errors.New("mongo unavailable")).Once()
		assert.Error(t, h.UpdateRejudgeStatus(judgeupdated.NewJudgeUpdatedEvent(j)))
	})

	t.Run("happy path returns nil", func(t *testing.T) {
		repo := judgemocks.NewRejudgeRepository(t)
		h := &JudgeUpdatedHandler{rejudgeRepo: repo, logger: sharedlogger.NewMockLogger()}
		j := newRejudgeJudgment(t, "rejudge-1", func(j *judgment.Judgment) {
			_ = j.Start()
			_ = j.Finish()
		})
		repo.On("TransitionStatus", "rejudge-1", []*rejudge.Status{rejudge.PENDING}, rejudge.RUNNING).
			Return(false, nil).Once()
		repo.On("AllJudgmentsFinished", "rejudge-1").Return(true, nil).Once()
		repo.On("TransitionStatus", "rejudge-1", []*rejudge.Status{rejudge.PENDING, rejudge.RUNNING}, rejudge.FINISHED).
			Return(true, nil).Once()
		assert.NoError(t, h.UpdateRejudgeStatus(judgeupdated.NewJudgeUpdatedEvent(j)))
	})
}

// TestJudgeUpdated_WiringPin_RetriesThroughRegistration pins the ROLLUP'S
// registration: RegisterJudgeUpdatedHandler must wire UpdateRejudgeStatus
// through SubscribeRetriedAsync, whose bounded retry is what keeps a
// transient Mongo failure from permanently stranding a rejudge in RUNNING.
// A success-path pin cannot tell the retried wiring from a raw
// SubscribeAsync — both run the handler once — so this test makes the first
// attempt fail and requires a second: were the registration ever downgraded
// to a raw subscription, the error would be silently discarded and the
// second expectation would fail this test.
//
// The one real cost is deliberate: production's RetryOptions are the
// hardcoded zero value (blocking sleep, Backoff(0) = 2s), so this test
// sleeps 2 seconds inside WaitAsync's barrier. That is the price of proving
// the retry at the WIRING level; zero-delay retry mechanics are covered at
// the eventbus layer (claimed_subscription_test's syncOpts).
func TestJudgeUpdated_WiringPin_RetriesThroughRegistration(t *testing.T) {
	logger := sharedlogger.NewMockLogger()
	eb := sharedeventbus.NewEventBus(logger)
	repo := judgemocks.NewRejudgeRepository(t)
	// Mid-run shape (Start, no terminal): the PENDING→RUNNING transition is
	// the handler's only repo call, so one failing and one succeeding
	// expectation cover attempt one and attempt two exactly.
	j := newRejudgeJudgment(t, "rejudge-1", func(j *judgment.Judgment) { _ = j.Start() })
	repo.On("TransitionStatus", "rejudge-1", []*rejudge.Status{rejudge.PENDING}, rejudge.RUNNING).
		Return(false, errors.New("mongo unavailable")).Once()
	repo.On("TransitionStatus", "rejudge-1", []*rejudge.Status{rejudge.PENDING}, rejudge.RUNNING).
		Return(true, nil).Once()

	RegisterJudgeUpdatedHandler(eb, repo, logger)

	eb.Publish(judgeupdated.NewJudgeUpdatedEvent(j))
	eb.WaitAsync()
}
