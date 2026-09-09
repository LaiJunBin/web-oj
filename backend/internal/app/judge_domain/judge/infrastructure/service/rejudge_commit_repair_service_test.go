package service

import (
	"sync"
	"testing"
	"time"

	judgemocks "judge-backend/internal/app/judge_domain/judge/mocks"
	sharedlogger "judge-backend/internal/app/shared/infrastructure/logger"
	"judge-backend/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestRejudgeCommitRepair_RepairOnce_Serialized pins the reentrancy guard
// added when the sweep became periodic: a tick firing during a slow in-flight
// sweep must wait, never interleave.
func TestRejudgeCommitRepair_RepairOnce_Serialized(t *testing.T) {
	var mu sync.Mutex
	inFlight, peak := 0, 0

	rejudgeRepo := judgemocks.NewRejudgeRepository(t)
	rejudgeRepo.On("RepairUncommittedApprovals").Return(int64(0), nil).Run(func(mock.Arguments) {
		mu.Lock()
		inFlight++
		if inFlight > peak {
			peak = inFlight
		}
		mu.Unlock()

		time.Sleep(20 * time.Millisecond)

		mu.Lock()
		inFlight--
		mu.Unlock()
	}).Twice()

	svc := &RejudgeCommitRepairService{
		rejudgeRepo: rejudgeRepo,
		logger:      sharedlogger.NewMockLogger(),
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

// TestRejudgeCommitRepair_RepairInterval covers the three-state knob.
func TestRejudgeCommitRepair_RepairInterval(t *testing.T) {
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
		{"explicit zero disables", &config.Config{Judge: &config.JudgeConfig{RejudgeCommitRepairInterval: intPtr(0)}}, 0, false},
		{"negative disables", &config.Config{Judge: &config.JudgeConfig{RejudgeCommitRepairInterval: intPtr(-1)}}, 0, false},
		{"positive value is seconds", &config.Config{Judge: &config.JudgeConfig{RejudgeCommitRepairInterval: intPtr(7)}}, 7 * time.Second, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &RejudgeCommitRepairService{config: tc.config, logger: sharedlogger.NewMockLogger()}
			interval, enabled := svc.repairInterval()
			assert.Equal(t, tc.enabled, enabled)
			assert.Equal(t, tc.interval, interval)
		})
	}
}
