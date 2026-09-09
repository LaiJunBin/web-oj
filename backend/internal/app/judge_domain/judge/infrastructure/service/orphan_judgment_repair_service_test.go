package service

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgestatus"
	judgetype "judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgetype"
	judgmentproject "judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/project"
	"judge-backend/internal/app/judge_domain/judge/infrastructure/queue"
	"judge-backend/internal/app/judge_domain/judge/infrastructure/worker"
	"judge-backend/internal/app/judge_domain/events/judgeupdated"
	judgemocks "judge-backend/internal/app/judge_domain/judge/mocks"
	sharedeventbus "judge-backend/internal/app/shared/infrastructure/eventbus"
	sharedmocks "judge-backend/internal/app/shared/mocks"
	sharedlogger "judge-backend/internal/app/shared/infrastructure/logger"
	"judge-backend/internal/config"
)

func newJudgingJudgment(id string) *judgment.Judgment {
	j := judgment.New(&judgment.JudgmentConfig{
		Project:   judgmentproject.NewProject(&judgmentproject.ProjectConfig{ID: 1}),
		CreatedAt: time.Now(),
	})
	j.SetID(id)
	j.SetStatus(judgestatus.JUDGING)
	return j
}

func newTestWorkerPool(t *testing.T) *worker.JudgeWorkerPool {
	logger := sharedlogger.NewMockLogger()
	eb := sharedeventbus.NewEventBus(logger)
	queueProvider := queue.NewJudgeQueueMap(logger)
	provisioner := judgemocks.NewExecutionEnvironmentProvisioner(t)
	cfg := &config.Config{Judge: &config.JudgeConfig{MaxJudgeDurationMinutes: 30}}
	return worker.NewJudgeWorkerPool(eb, queueProvider, nil, provisioner, cfg, logger)
}

// TestOrphanJudgmentRepairService_RepairOnce_NoOrphans — does nothing when
// the DB holds no Judging judgments.
func TestOrphanJudgmentRepairService_RepairOnce_NoOrphans(t *testing.T) {
	judgmentRepo := judgemocks.NewJudgmentRepository(t)
	judgmentRepo.On("FindJudgings").Return([]*judgment.Judgment{}, nil).Once()
	eventbus := sharedmocks.NewEventBus(t)

	svc := &OrphanJudgmentRepairService{
		judgmentRepo: judgmentRepo,
		workerPool:   newTestWorkerPool(t),
		eventbus:     eventbus,
		logger:       sharedlogger.NewMockLogger(),
	}
	svc.RepairOnce()
}

// TestOrphanJudgmentRepairService_RepairOnce_MarksOrphansAsError
// — a Judging judgment no worker is executing is updated to ERROR with the
// proper reason / errorMessage / finishedAt.
func TestOrphanJudgmentRepairService_RepairOnce_MarksOrphansAsError(t *testing.T) {
	judgmentRepo := judgemocks.NewJudgmentRepository(t)

	orphan1 := newJudgingJudgment("orphan-1")
	orphan2 := newJudgingJudgment("orphan-2")

	judgmentRepo.On("FindJudgings").Return([]*judgment.Judgment{orphan1, orphan2}, nil).Once()

	eventbus := sharedmocks.NewEventBus(t)
	eventbus.On("Publish", mock.MatchedBy(func(evt *judgeupdated.JudgeUpdatedEvent) bool {
		return evt.Data().StatusIsError
	})).Twice()

	var saved []*judgment.Judgment
	judgmentRepo.On("SaveTerminalIfJudging", mock.AnythingOfType("*judgment.Judgment")).Return(true, nil).Run(func(args mock.Arguments) {
		saved = append(saved, args.Get(0).(*judgment.Judgment))
	}).Twice()

	svc := &OrphanJudgmentRepairService{
		judgmentRepo: judgmentRepo,
		workerPool:   newTestWorkerPool(t),
		eventbus:     eventbus,
		logger:       sharedlogger.NewMockLogger(),
	}
	svc.RepairOnce()

	if assert.Len(t, saved, 2) {
		for _, j := range saved {
			assert.True(t, j.Status().IsError(), "expected ERROR status, got %s", j.Status())
			assert.Equal(t, orphanJudgmentRepairReason, *j.Reason())
			assert.Equal(t, orphanJudgmentRepairErrorMessage, *j.ErrorMessage())
			assert.NotNil(t, j.JudgingFinishedAt(), "JudgingFinishedAt should be set")
		}
	}
}

// TestOrphanJudgmentRepairService_RepairOnce_JudgeTypePreserved
// — the repair never changes judgeType (a REJUDGE member stays REJUDGE).
func TestOrphanJudgmentRepairService_RepairOnce_JudgeTypePreserved(t *testing.T) {
	judgmentRepo := judgemocks.NewJudgmentRepository(t)

	j := judgment.New(&judgment.JudgmentConfig{
		Project:   judgmentproject.NewProject(&judgmentproject.ProjectConfig{ID: 1}),
		JudgeType: judgetype.REJUDGE,
		CreatedAt: time.Now(),
	})
	j.SetID("rejudge-orphan")
	j.SetStatus(judgestatus.JUDGING)

	judgmentRepo.On("FindJudgings").Return([]*judgment.Judgment{j}, nil).Once()

	eventbus := sharedmocks.NewEventBus(t)
	eventbus.On("Publish", mock.MatchedBy(func(evt *judgeupdated.JudgeUpdatedEvent) bool {
		return evt.Data().JudgmentID == "rejudge-orphan" && evt.Data().StatusIsError
	})).Once()

	var saved []*judgment.Judgment
	judgmentRepo.On("SaveTerminalIfJudging", mock.AnythingOfType("*judgment.Judgment")).Return(true, nil).Run(func(args mock.Arguments) {
		saved = append(saved, args.Get(0).(*judgment.Judgment))
	}).Once()

	svc := &OrphanJudgmentRepairService{
		judgmentRepo: judgmentRepo,
		workerPool:   newTestWorkerPool(t),
		eventbus:     eventbus,
		logger:       sharedlogger.NewMockLogger(),
	}
	svc.RepairOnce()

	if assert.Len(t, saved, 1) {
		assert.True(t, saved[0].Status().IsError())
		assert.Equal(t, judgetype.REJUDGE.String(), saved[0].JudgeType().String(), "judgeType must not be changed by repair")
	}
}

// TestOrphanJudgmentRepairService_RepairOnce_SkipsJudgmentAnotherWriterDecided
// is the regression test for the unconditional-Save version of this service:
// RepairOnce runs on its own goroutine with a snapshot from FindJudgings, and
// an API cancel (or another instance's worker) can move the document out of
// Judging before the repair write lands. The old Save would overwrite that
// accepted outcome with a stale ERROR; SaveTerminalIfJudging reports
// applied=false instead, and the service must treat that as "already
// decided" and move on — processing the remaining orphans, not aborting.
func TestOrphanJudgmentRepairService_RepairOnce_SkipsJudgmentAnotherWriterDecided(t *testing.T) {
	judgmentRepo := judgemocks.NewJudgmentRepository(t)

	lost := newJudgingJudgment("orphan-lost-race")
	won := newJudgingJudgment("orphan-won")

	judgmentRepo.On("FindJudgings").Return([]*judgment.Judgment{lost, won}, nil).Once()
	judgmentRepo.On("SaveTerminalIfJudging", lost).Return(false, nil).Once()
	judgmentRepo.On("SaveTerminalIfJudging", won).Return(true, nil).Once()

	// Exactly one judgeupdated: the judgment whose repair actually applied.
	// Publishing for the lost one would announce an ERROR the database never
	// stored — the strict mock panics on any second Publish.
	eventbus := sharedmocks.NewEventBus(t)
	eventbus.On("Publish", mock.MatchedBy(func(evt *judgeupdated.JudgeUpdatedEvent) bool {
		return evt.Data().JudgmentID == "orphan-won" && evt.Data().StatusIsError
	})).Once()

	svc := &OrphanJudgmentRepairService{
		judgmentRepo: judgmentRepo,
		workerPool:   newTestWorkerPool(t),
		eventbus:     eventbus,
		logger:       sharedlogger.NewMockLogger(),
	}
	svc.RepairOnce()
}

// TestOrphanJudgmentRepairService_ActiveWorkerJudgmentIDs_EmptyWhenIdlePool
// — an idle pool (every currentJudgeID empty) yields an empty activeIDs set.
func TestOrphanJudgmentRepairService_ActiveWorkerJudgmentIDs_EmptyWhenIdlePool(t *testing.T) {
	judgmentRepo := judgemocks.NewJudgmentRepository(t)
	eventbus := sharedmocks.NewEventBus(t)

	svc := &OrphanJudgmentRepairService{
		judgmentRepo: judgmentRepo,
		workerPool:   newTestWorkerPool(t),
		eventbus:     eventbus,
		logger:       sharedlogger.NewMockLogger(),
	}

	activeIDs := svc.activeWorkerJudgmentIDs()
	assert.Empty(t, activeIDs, "idle worker pool should have no active judgment IDs")
}

// TestOrphanJudgmentRepairService_RepairOnce_OrphanStatusDiffersFromSubmissionRepair
// — the two sweeps stay distinguishable for the operator through their error
// messages (admin API). Their REASONS are deliberately the same sentence: the
// reason is what a contestant sees, and from their side both events are "the
// judging was interrupted", so it must not carry the internal distinction.
func TestOrphanJudgmentRepairService_RepairOnce_OrphanStatusDiffersFromSubmissionRepair(t *testing.T) {
	assert.NotEqual(t, orphanJudgmentRepairErrorMessage, orphanRepairErrorMessage,
		"orphan judgment error message must be distinct from orphan submission error message")
	assert.Equal(t, orphanRepairReason, orphanJudgmentRepairReason,
		"both sweeps share the contestant-facing reason by design")
	assert.NotContains(t, orphanJudgmentRepairReason, "worker",
		"the contestant-facing reason must not name internals")
}

// TestOrphanJudgmentRepairService_RepairOnce_Serialized pins the reentrancy
// guard added when the sweep became periodic: a tick that fires while a slow
// sweep is still running must wait, never interleave — the in-flight count
// across concurrent RepairOnce calls stays at one.
func TestOrphanJudgmentRepairService_RepairOnce_Serialized(t *testing.T) {
	var mu sync.Mutex
	inFlight, peak := 0, 0

	judgmentRepo := judgemocks.NewJudgmentRepository(t)
	judgmentRepo.On("FindJudgings").Return([]*judgment.Judgment{}, nil).Run(func(mock.Arguments) {
		mu.Lock()
		inFlight++
		if inFlight > peak {
			peak = inFlight
		}
		mu.Unlock()

		time.Sleep(20 * time.Millisecond) // long enough for the other goroutine to try to enter

		mu.Lock()
		inFlight--
		mu.Unlock()
	}).Twice()

	svc := &OrphanJudgmentRepairService{
		judgmentRepo: judgmentRepo,
		workerPool:   newTestWorkerPool(t),
		eventbus:     sharedmocks.NewEventBus(t),
		logger:       sharedlogger.NewMockLogger(),
	}

	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			svc.RepairOnce()
		}()
	}
	wg.Wait()

	assert.Equal(t, 1, peak, "concurrent RepairOnce calls must serialize, not interleave")
}

// TestOrphanJudgmentRepairService_RepairInterval covers the three-state knob:
// nil-safety falls back to 5 minutes, an explicit non-positive value disables
// the periodic sweep, a positive value is seconds.
func TestOrphanJudgmentRepairService_RepairInterval(t *testing.T) {
	intPtr := func(v int) *int { return &v }

	cases := []struct {
		name     string
		config   *config.Config
		interval time.Duration
		enabled  bool
	}{
		{"nil config falls back to 5 minutes", nil, 5 * time.Minute, true},
		{"nil judge section falls back to 5 minutes", &config.Config{}, 5 * time.Minute, true},
		{"nil field falls back to 5 minutes", &config.Config{Judge: &config.JudgeConfig{}}, 5 * time.Minute, true},
		{"explicit zero disables", &config.Config{Judge: &config.JudgeConfig{OrphanJudgmentRepairInterval: intPtr(0)}}, 0, false},
		{"negative disables", &config.Config{Judge: &config.JudgeConfig{OrphanJudgmentRepairInterval: intPtr(-1)}}, 0, false},
		{"positive value is seconds", &config.Config{Judge: &config.JudgeConfig{OrphanJudgmentRepairInterval: intPtr(7)}}, 7 * time.Second, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &OrphanJudgmentRepairService{config: tc.config, logger: sharedlogger.NewMockLogger()}
			interval, enabled := svc.repairInterval()
			assert.Equal(t, tc.enabled, enabled)
			assert.Equal(t, tc.interval, interval)
		})
	}
}
