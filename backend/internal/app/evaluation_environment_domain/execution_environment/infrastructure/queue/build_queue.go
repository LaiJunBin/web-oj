package queue

import (
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/evaluation_environment_domain/events"
	"sync"
	"time"
)

type BuildQueue interface {
	Enqueue(event *event.BuildImageEvent)
}

type buildQueue struct {
	eventbus sharedtypes.EventBus
	events   *sharedtypes.Heap[*event.BuildImageEvent]
}

func NewBuildQueue(
	eventbus sharedtypes.EventBus,
	workerCount int,
) BuildQueue {
	e := &buildQueue{
		eventbus: eventbus,
		events: sharedtypes.NewHeap([]*event.BuildImageEvent{}, func(i, j *event.BuildImageEvent) bool {
			return i.CreatedAt().Before(j.CreatedAt())
		}),
	}

	e.initWorker(workerCount)
	return e
}

func (e *buildQueue) initWorker(workerCount int) {
	lock := sync.Mutex{}
	for i := 0; i < workerCount; i++ {
		go func(i int) {
			for {
				lock.Lock()
				if e.events.Len() == 0 {
					lock.Unlock()
					time.Sleep(1 * time.Second)
					continue
				}

				event := e.events.Pop()
				lock.Unlock()
				e.eventbus.Publish(event)
			}
		}(i)
	}
}

func (e *buildQueue) Enqueue(event *event.BuildImageEvent) {
	e.events.Push(event)
}
