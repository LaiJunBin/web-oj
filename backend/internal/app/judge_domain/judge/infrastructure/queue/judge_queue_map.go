package queue

import (
	"fmt"
	"judge-backend/internal/app/judge_domain/events/judgmentcreated"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgetype"
	"judge-backend/internal/app/judge_domain/judge/domain/queue"
	"judge-backend/internal/app/judge_domain/judge/domain/repository"
	sharedtypes "judge-backend/internal/app/shared/types"
	"sync"
)

type JudgeQueueMap struct {
	mu       sync.RWMutex
	queueMap map[string]queue.JudgeQueue
	logger   sharedtypes.Logger
}

func NewJudgeQueueMap(
	logger sharedtypes.Logger,
) queue.JudgeQueueProvider {
	queueMap := make(map[string]queue.JudgeQueue)
	for _, judgeType := range judgetype.JudgeTypeMap {
		queueMap[judgeType.String()] = NewJudgeQueue(judgeType.String(), logger)
		logger.Infof("[JudgeQueueMap] Created queue for judge type: %s", judgeType.String())
	}

	return &JudgeQueueMap{
		queueMap: queueMap,
		logger:   logger,
	}
}

// LoadQueuedJudgmentsOnStartup recovers judgments still QUEUED from before a
// restart (see JudgeQueueProvider's durability note) by republishing a
// judgment_created event for each. Wired as an fx Invoke — deliberately NOT
// done inside NewJudgeQueueMap's constructor, and deliberately NOT a
// fire-and-forget goroutine:
//
//   - Ordering: fx constructs a Provide function's dependencies before
//     running the Invoke that first needed them, so if this ran inside
//     NewJudgeQueueMap it would race RegisterJudgmentCreatedHandler's own
//     SubscribeAsync call — Publish is a no-op when nothing is subscribed
//     yet, silently dropping every recovered judgment. Handlers are wired
//     as Invokes ahead of this domain's own Invokes list (see
//     app/provider/provider.go's GetProvides), so by the time this runs,
//     the subscription is guaranteed to already exist.
//   - Synchronicity: the old fire-and-forget goroutine wasn't tracked by
//     anything — not the eventbus wait group, nothing — so a caller with no
//     way to know it was even running could observe a "ready" server that
//     still had judgments about to be pushed into a queue moments later. In
//     the BDD reset seam that meant a scenario's first reset (or even its
//     assertions) could race this recovery load with no way to wait it out.
//     Running synchronously here means it's done by the time fx finishes
//     starting the app; the actual queue pushes are still async (through
//     the eventbus like any other judgment_created event), but that part is
//     now covered by the seam's existing WaitAsync call like everything
//     else.
func LoadQueuedJudgmentsOnStartup(
	repo repository.JudgmentRepository,
	eventbus sharedtypes.EventBus,
	logger sharedtypes.Logger,
) {
	logger.Infof("[JudgeQueueMap] Starting to load queued judgments")
	judgments, err := repo.LoadQueuedJudgments()
	if err != nil {
		logger.Errorf("[JudgeQueueMap] Failed to load queued judgments: %v", err)
		return
	}

	logger.Infof("[JudgeQueueMap] Successfully loaded %d queued judgments", len(judgments))
	for _, judgment := range judgments {
		event := judgmentcreated.NewJudgmentCreatedEvent(judgment)
		eventbus.Publish(event)
		logger.Debugf("[JudgeQueueMap] Published judgment created event for judgment: %s", judgment.ID())
	}
	logger.Infof("[JudgeQueueMap] Finished processing queued judgments")
}

func (jqm *JudgeQueueMap) GetQueue(key string) (queue.JudgeQueue, error) {
	jqm.mu.RLock()
	defer jqm.mu.RUnlock()
	if _, ok := jqm.queueMap[key]; !ok {
		jqm.logger.Errorf("[JudgeQueueMap] Queue not found for key: %s", key)
		return nil, fmt.Errorf("queue not found: %s", key)
	}
	return jqm.queueMap[key], nil
}

// TryPop looks up the queue for judgeType and pops from it while holding the
// SAME lock Reset() takes to swap the queue map — see the interface doc
// comment for why the lookup and the pop cannot be two separate critical
// sections.
func (jqm *JudgeQueueMap) TryPop(judgeType string) (*judgment.Judgment, error) {
	jqm.mu.RLock()
	defer jqm.mu.RUnlock()
	q, ok := jqm.queueMap[judgeType]
	if !ok {
		jqm.logger.Errorf("[JudgeQueueMap] Queue not found for key: %s", judgeType)
		return nil, fmt.Errorf("queue not found: %s", judgeType)
	}
	return q.Pop(), nil
}

// Reset replaces every queue with a fresh empty one, dropping all queued
// judgments. Test seam; see queue.JudgeQueueProvider.
func (jqm *JudgeQueueMap) Reset() {
	jqm.mu.Lock()
	defer jqm.mu.Unlock()
	for _, judgeType := range judgetype.JudgeTypeMap {
		jqm.queueMap[judgeType.String()] = NewJudgeQueue(judgeType.String(), jqm.logger)
	}
	jqm.logger.Infof("[JudgeQueueMap] Reset all queues")
}
