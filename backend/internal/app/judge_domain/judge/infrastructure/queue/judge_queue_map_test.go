package queue_test

import (
	"errors"
	"fmt"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgetype"
	"judge-backend/internal/app/judge_domain/judge/domain/prioritystrategies"
	q "judge-backend/internal/app/judge_domain/judge/domain/queue"
	"judge-backend/internal/app/judge_domain/judge/infrastructure/queue"
	judgemocks "judge-backend/internal/app/judge_domain/judge/mocks"
	"judge-backend/internal/app/shared/infrastructure/logger"
	sharedmocks "judge-backend/internal/app/shared/mocks"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
)

func newTestQueueProvider(t *testing.T) q.JudgeQueueProvider {
	l := logger.NewMockLogger()
	return queue.NewJudgeQueueMap(l)
}

// TestJudgeQueueMap_TryPop_DoesNotReturnJudgmentFromQueueDetachedByReset is a
// regression test for the BDD reset-seam ghost-write race: a worker used to
// call GetQueue(judgeType) to get a *reference* to the live queue, then Pop
// from that reference later. Reset() only swaps the map entry to a fresh
// queue — it never touches queues it has already handed out — so a worker
// holding a pre-reset reference kept popping judgments the test believed it
// had cleared. This deterministically demonstrates both halves: the stale
// reference stays poppable forever (proving why GetQueue-then-Pop is unsafe),
// while TryPop, which re-resolves the map entry under the same lock Reset
// uses to swap it, can only ever observe the live post-reset queue.
func TestJudgeQueueMap_TryPop_DoesNotReturnJudgmentFromQueueDetachedByReset(t *testing.T) {
	provider := newTestQueueProvider(t)

	staleQueue, err := provider.GetQueue(judgetype.JUDGE.String())
	if err != nil {
		t.Fatalf("GetQueue failed: %v", err)
	}
	staleQueue.Push(newJudgment("stale-judgment", 1, time.Now()), prioritystrategies.NewDefault())

	provider.Reset()

	if popped := staleQueue.Pop(); popped == nil || popped.ID() != "stale-judgment" {
		t.Fatalf("expected the detached queue reference to still hold its judgment, got %v", popped)
	}

	got, err := provider.TryPop(judgetype.JUDGE.String())
	if err != nil {
		t.Fatalf("TryPop failed: %v", err)
	}
	if got != nil {
		t.Fatalf("expected TryPop to see the reset (empty) queue, got judgment %s", got.ID())
	}
}

// TestJudgeQueueMap_TryPop_ConcurrentWithReset_NoDataRaceOrDuplicatePop runs
// many iterations of concurrent TryPop callers (simulating workers) racing a
// single Reset() (simulating the BDD reset seam) against a seeded queue.
// Meant to be run with `go test -race`: TryPop and Reset share jqm.mu, so
// this proves there is no unsynchronized access left in the queue map, and
// separately asserts every judgment is popped at most once no matter how the
// pops and the reset interleave.
func TestJudgeQueueMap_TryPop_ConcurrentWithReset_NoDataRaceOrDuplicatePop(t *testing.T) {
	const seeded = 20
	const pollers = 4

	for iter := 0; iter < 50; iter++ {
		provider := newTestQueueProvider(t)
		seedQueue, err := provider.GetQueue(judgetype.JUDGE.String())
		if err != nil {
			t.Fatalf("GetQueue failed: %v", err)
		}
		for i := 0; i < seeded; i++ {
			id := fmt.Sprintf("iter-%d-judgment-%d", iter, i)
			seedQueue.Push(newJudgment(id, 1, time.Now().Add(time.Duration(i)*time.Millisecond)), prioritystrategies.NewDefault())
		}

		popped := make(chan *judgment.Judgment, seeded*pollers)
		var wg sync.WaitGroup

		for p := 0; p < pollers; p++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < seeded; i++ {
					j, err := provider.TryPop(judgetype.JUDGE.String())
					if err != nil {
						t.Errorf("TryPop returned error: %v", err)
						return
					}
					if j != nil {
						popped <- j
					}
				}
			}()
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(time.Millisecond)
			provider.Reset()
		}()

		wg.Wait()
		close(popped)

		seen := make(map[string]bool)
		for j := range popped {
			if seen[j.ID()] {
				t.Fatalf("iteration %d: judgment %s popped more than once", iter, j.ID())
			}
			seen[j.ID()] = true
		}
	}
}

// TestLoadQueuedJudgmentsOnStartup_PublishesEventPerQueuedJudgment covers
// LoadQueuedJudgmentsOnStartup directly — something the old fire-and-forget
// goroutine version couldn't be given a deterministic test for at all, since
// nothing signalled when (or whether) it had run.
func TestLoadQueuedJudgmentsOnStartup_PublishesEventPerQueuedJudgment(t *testing.T) {
	l := logger.NewMockLogger()
	repo := judgemocks.NewJudgmentRepository(t)
	eventbus := sharedmocks.NewEventBus(t)

	j1 := newJudgment("queued-1", 1, time.Now())
	j2 := newJudgment("queued-2", 1, time.Now())
	repo.On("LoadQueuedJudgments").Return([]*judgment.Judgment{j1, j2}, nil).Once()

	published := map[string]bool{}
	eventbus.On("Publish", mock.MatchedBy(func(evt interface{ Judgment() *judgment.Judgment }) bool {
		published[evt.Judgment().ID()] = true
		return true
	})).Twice()

	queue.LoadQueuedJudgmentsOnStartup(repo, eventbus, l)

	if !published["queued-1"] || !published["queued-2"] {
		t.Fatalf("expected an event published for both queued judgments, got %v", published)
	}
}

func TestLoadQueuedJudgmentsOnStartup_RepoError_NoPublish(t *testing.T) {
	l := logger.NewMockLogger()
	repo := judgemocks.NewJudgmentRepository(t)
	eventbus := sharedmocks.NewEventBus(t)

	repo.On("LoadQueuedJudgments").Return(nil, errors.New("mongo unavailable")).Once()
	// No Publish expectation: an unexpected call would panic, proving
	// nothing gets published when the load itself fails.

	queue.LoadQueuedJudgmentsOnStartup(repo, eventbus, l)
}

func TestLoadQueuedJudgmentsOnStartup_NoQueuedJudgments_NoPublish(t *testing.T) {
	l := logger.NewMockLogger()
	repo := judgemocks.NewJudgmentRepository(t)
	eventbus := sharedmocks.NewEventBus(t)

	repo.On("LoadQueuedJudgments").Return([]*judgment.Judgment{}, nil).Once()

	queue.LoadQueuedJudgmentsOnStartup(repo, eventbus, l)
}
