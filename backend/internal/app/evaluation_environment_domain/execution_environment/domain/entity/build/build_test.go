package build_test

import (
	"sync"
	"testing"
	"time"

	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/batch"
	buildlog "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/log"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/buildloglevel"
	buildstatus "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/status"
	imagetag "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/image/valueobject/tag"

	"github.com/stretchr/testify/assert"
)

func newPendingBuild() *build.Build {
	return build.New(&build.BuildConfig{
		Tag:    imagetag.MustNew("entity-test:latest"),
		Status: buildstatus.PENDING,
		Batch:  batch.MustNew(1),
	})
}

// TestBuild_SetStatus_AtomicCompareAndSet pins the entity's concurrency
// contract directly: BuildRepository's buildMap hands the SAME instance to
// every caller of a tag, and queue workers, the cancel handler, and API
// goroutines all transition it. SetStatus must be a real compare-and-set —
// of N concurrent attempts at the same transition, exactly ONE may succeed —
// and status/log access must be memory-safe (this test is what -race uses
// to catch an unsynchronized entity; the handler-level tests sequence their
// steps too deterministically to overlap here).
func TestBuild_SetStatus_AtomicCompareAndSet(t *testing.T) {
	const goroutines = 8
	const rounds = 50

	for round := 0; round < rounds; round++ {
		b := newPendingBuild()
		var wg sync.WaitGroup
		start := make(chan struct{})
		succeeded := make([]bool, goroutines)

		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				<-start
				// Every goroutine also touches logs and reads status, so the
				// race detector sees genuinely overlapping access on all the
				// fields the mutex must guard.
				b.AddLog(buildlog.New(buildloglevel.Info, "attempting transition", time.Now()))
				succeeded[idx] = b.SetStatus(buildstatus.BUILDING) == nil
				_ = b.Status()
				_ = b.Logs()
			}(i)
		}
		close(start)
		wg.Wait()

		winners := 0
		for _, ok := range succeeded {
			if ok {
				winners++
			}
		}
		assert.Equal(t, 1, winners, "round %d: exactly one goroutine may win the PENDING->BUILDING transition, got %d", round, winners)
		assert.True(t, b.IsBuilding())
		assert.Len(t, b.Logs(), goroutines, "round %d: no log append may be lost", round)
	}
}
