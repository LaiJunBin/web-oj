package queue_test

import (
	"fmt"
	"judge-backend/internal/app/shared/infrastructure/eventbus"
	"judge-backend/internal/app/shared/infrastructure/logger"
	"judge-backend/internal/app/evaluation_environment_domain/events"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/infrastructure/queue"
	"testing"
	"time"
)

func checkEventHandled(t *testing.T, handled chan bool) {
	select {
	case <-handled:
	case <-time.After(5000 * time.Millisecond):
		t.Fatalf("event was not handled in time")
	}
}

func TestBuildQueue(t *testing.T) {
	t.Run("should subscribe and handle event", func(t *testing.T) {
		logger := logger.NewMockLogger()
		eventBus := eventbus.NewEventBus(logger)
		buildQueue := queue.NewBuildQueue(eventBus, 2)
		buildImageEvent := event.NewBuildImageEvent(&event.BuildImageEventConfig{
			Tag:     "test:latest",
			Path:    "test",
			Options: &event.BuildImageOptions{},
		})
		handled := make(chan bool)

		eventBus.Subscribe(func(event *event.BuildImageEvent) {
			handled <- true
		})

		buildQueue.Enqueue(buildImageEvent)
		checkEventHandled(t, handled)
	})
}

func TestBuildQueue_Loading(t *testing.T) {
	t.Run("should handle massive events", func(t *testing.T) {
		logger := logger.NewMockLogger()
		eventBus := eventbus.NewEventBus(logger)
		buildQueue := queue.NewBuildQueue(eventBus, 10)
		events := make([]*event.BuildImageEvent, 1000)
		handlers := map[string]chan bool{}
		for i := 0; i < 1000; i++ {
			tag := fmt.Sprintf("test-%d:latest", i)
			events[i] = event.NewBuildImageEvent(&event.BuildImageEventConfig{
				Tag:     tag,
				Path:    "test",
				Options: &event.BuildImageOptions{},
			})
			handlers[tag] = make(chan bool, 1)
		}

		eventBus.Subscribe(func(event *event.BuildImageEvent) {
			handlers[event.Tag()] <- true
		})

		for i := 0; i < 1000; i++ {
			buildQueue.Enqueue(events[i])
		}

		for i := 0; i < 1000; i++ {
			checkEventHandled(t, handlers[events[i].Tag()])
		}
	})
}
