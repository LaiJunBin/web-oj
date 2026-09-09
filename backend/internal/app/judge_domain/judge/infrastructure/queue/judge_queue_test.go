package queue_test

import (
	"fmt"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/project"
	"judge-backend/internal/app/judge_domain/judge/domain/prioritystrategies"
	q "judge-backend/internal/app/judge_domain/judge/domain/queue"
	"judge-backend/internal/app/judge_domain/judge/infrastructure/queue"
	"judge-backend/internal/app/judge_domain/judge/mocks"
	"judge-backend/internal/app/shared/infrastructure/logger"
	"judge-backend/internal/app/shared/infrastructure/metrics"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

type reverseTimeStrategy struct{}

func NewSubmitTime() prioritystrategies.Strategy {
	return &reverseTimeStrategy{}
}

func (s *reverseTimeStrategy) Name() string {
	return "reverse_time"
}

func (s *reverseTimeStrategy) Priority(i, j *judgment.Judgment) bool {
	return i.CreatedAt().After(j.CreatedAt())
}

func (s *reverseTimeStrategy) GetExtraFields(j *judgment.Judgment) map[string]interface{} {
	return nil
}

// blockingStrategy's GetExtraFields signals `started` then blocks on
// `unblock`, simulating a strategy whose extra-fields lookup does slow I/O
// (e.g. TimeSinceLast's DB queries). Used to prove Push does not hold the
// queue lock while this runs.
type blockingStrategy struct {
	started chan struct{}
	unblock chan struct{}
}

func (s *blockingStrategy) Name() string { return "blocking" }
func (s *blockingStrategy) Priority(i, j *judgment.Judgment) bool {
	return true
}
func (s *blockingStrategy) GetExtraFields(j *judgment.Judgment) map[string]interface{} {
	close(s.started)
	<-s.unblock
	return nil
}

func newJudgment(judgmentID string, projectID int32, createdAt time.Time) *judgment.Judgment {
	j := judgment.New(&judgment.JudgmentConfig{
		Project: project.NewProject(&project.ProjectConfig{
			ID: projectID,
		}),
		CreatedAt: createdAt,
	})
	j.SetID(judgmentID)
	return j
}

func pushJudgmentToQueue(jq q.JudgeQueue, j *judgment.Judgment, strategy prioritystrategies.Strategy) {
	j.SetSortingFields(strategy.GetExtraFields(j))
	// fmt.Println("Pushing:", j.ID(), "with strategy:", strategy.Name(), "and fields:", j.SortingFields())
	jq.Push(j, strategy)
}

func TestJudgeQueue_DepthMetric(t *testing.T) {
	// A dedicated label keeps this test's gauge series isolated from every
	// other test that constructs "JUDGE"-named queues.
	const name = "JUDGE-depth-metric-test"
	depth := func() float64 {
		return testutil.ToFloat64(metrics.JudgeQueueDepth.WithLabelValues(name))
	}

	priorityStrategy := prioritystrategies.NewDefault()
	jq := queue.NewJudgeQueue(name, logger.NewMockLogger())
	if depth() != 0 {
		t.Fatalf("fresh queue must report depth 0, got %v", depth())
	}

	// two projects: the gauge must count judgments, not project sub-queues
	jq.Push(newJudgment("judgment 1", 1, time.Now()), priorityStrategy)
	jq.Push(newJudgment("judgment 2", 1, time.Now().Add(time.Second)), priorityStrategy)
	jq.Push(newJudgment("judgment 3", 2, time.Now().Add(2*time.Second)), priorityStrategy)
	if depth() != 3 {
		t.Fatalf("expected depth 3 after three pushes, got %v", depth())
	}

	jq.Pop()
	if depth() != 2 {
		t.Fatalf("expected depth 2 after one pop, got %v", depth())
	}

	jq.Pop()
	jq.Pop()
	if depth() != 0 {
		t.Fatalf("expected depth 0 after draining, got %v", depth())
	}

	// popping empty must not go negative
	jq.Pop()
	if depth() != 0 {
		t.Fatalf("expected depth to stay 0 on empty pop, got %v", depth())
	}
}

func TestJudgeQueue(t *testing.T) {
	t.Run("push and pop", func(t *testing.T) {
		priorityStrategy := prioritystrategies.NewDefault()
		logger := logger.NewMockLogger()
		jq := queue.NewJudgeQueue("JUDGE", logger)
		j := newJudgment("judgment 1", 1, time.Now())
		jq.Push(j, priorityStrategy)
		if j != jq.Pop() {
			t.Error("Judgment not popped correctly")
		}
	})

	t.Run("push and pop with multiple judgments", func(t *testing.T) {
		priorityStrategy := prioritystrategies.NewDefault()
		logger := logger.NewMockLogger()
		jq := queue.NewJudgeQueue("JUDGE", logger)
		jq.Push(newJudgment("judgment 1", 1, time.Now()), priorityStrategy)
		jq.Push(newJudgment("judgment 2", 1, time.Now().Add(time.Second)), priorityStrategy)
		jq.Push(newJudgment("judgment 3", 1, time.Now().Add(2*time.Second)), priorityStrategy)
		jq.Push(newJudgment("judgment 4", 1, time.Now().Add(3*time.Second)), priorityStrategy)
		jq.Push(newJudgment("judgment 5", 1, time.Now().Add(4*time.Second)), priorityStrategy)

		sequence := []string{"judgment 1", "judgment 2", "judgment 3", "judgment 4", "judgment 5"}
		for _, j := range sequence {
			if jq.Pop().ID() != j {
				t.Error("Judgment not popped correctly")
			}
		}
	})

	t.Run("push and pop with multiple projects", func(t *testing.T) {
		priorityStrategy := prioritystrategies.NewDefault()
		logger := logger.NewMockLogger()
		jq := queue.NewJudgeQueue("JUDGE", logger)
		jq.Push(newJudgment("judgment 1", 1, time.Now()), priorityStrategy)
		jq.Push(newJudgment("judgment 2", 2, time.Now().Add(time.Second)), priorityStrategy)
		jq.Push(newJudgment("judgment 3", 1, time.Now().Add(2*time.Second)), priorityStrategy)
		jq.Push(newJudgment("judgment 4", 2, time.Now().Add(3*time.Second)), priorityStrategy)
		jq.Push(newJudgment("judgment 5", 1, time.Now().Add(4*time.Second)), priorityStrategy)

		sequence := []string{"judgment 1", "judgment 2", "judgment 3", "judgment 4", "judgment 5"}
		for _, j := range sequence {
			if jq.Pop().ID() != j {
				t.Error("Judgment not popped correctly")
			}
		}
	})

	t.Run("pop removes queue mapping for another project", func(t *testing.T) {
		priorityStrategy := prioritystrategies.NewDefault()
		logger := logger.NewMockLogger()
		jq := queue.NewJudgeQueue("JUDGE", logger)

		base := time.Now()

		jq.Push(newJudgment("judgment project1 first", 1, base), priorityStrategy)
		jq.Push(newJudgment("judgment project2 only", 2, base.Add(10*time.Second)), priorityStrategy)
		jq.Push(newJudgment("judgment project1 remaining", 1, base.Add(20*time.Second)), priorityStrategy)

		if popped := jq.Pop(); popped.ID() != "judgment project1 first" {
			t.Fatalf("expected first pop to stay in project 1, got %s", popped.ID())
		}

		if popped := jq.Pop(); popped.ID() != "judgment project2 only" {
			t.Fatalf("expected second pop to remove project 2 queue, got %s", popped.ID())
		}

		if jq.Len() != 1 {
			t.Fatalf("expected queue length 1 with remaining project 1 item, got %d", jq.Len())
		}

		jq.Push(newJudgment("judgment project1 new", 1, base.Add(30*time.Second)), priorityStrategy)

		if jq.Len() != 1 {
			t.Fatalf("expected queue length to remain 1 for project 1, got %d", jq.Len())
		}
	})

	t.Run("reinitializes queue with new strategy after empty", func(t *testing.T) {
		defaultStrategy := prioritystrategies.NewDefault()
		reverseTime := &reverseTimeStrategy{}
		logger := logger.NewMockLogger()
		jq := queue.NewJudgeQueue("JUDGE", logger)

		base := time.Now()

		j1 := newJudgment("p1-default-1", 1, base)
		j2 := newJudgment("p1-default-2", 1, base.Add(10*time.Second))

		jq.Push(j1, defaultStrategy)
		jq.Push(j2, defaultStrategy)

		if popped := jq.Pop(); popped.ID() != j1.ID() {
			t.Fatalf("expected first pop to follow default order, got %s", popped.ID())
		}
		if popped := jq.Pop(); popped.ID() != j2.ID() {
			t.Fatalf("expected second pop to complete default order, got %s", popped.ID())
		}

		j3 := newJudgment("p1-reverse-1", 1, base.Add(20*time.Second))
		j4 := newJudgment("p1-reverse-2", 1, base.Add(30*time.Second))

		jq.Push(j3, reverseTime)
		jq.Push(j4, reverseTime)

		if popped := jq.Pop(); popped.ID() != j4.ID() {
			t.Fatalf("expected new queue to respect reverse strategy, got %s", popped.ID())
		}
		if popped := jq.Pop(); popped.ID() != j3.ID() {
			t.Fatalf("expected second reverse pop to be older item, got %s", popped.ID())
		}

		if jq.Len() != 0 {
			t.Fatalf("expected queue to be empty after reinitialized pops, got len=%d", jq.Len())
		}
	})

	t.Run("strategy switch reorders existing queue across projects", func(t *testing.T) {
		defaultStrategy := prioritystrategies.NewDefault()
		reverseTime := &reverseTimeStrategy{}
		logger := logger.NewMockLogger()
		jq := queue.NewJudgeQueue("JUDGE", logger)

		base := time.Now()

		p1a := newJudgment("project1-initial-1", 1, base)
		p1b := newJudgment("project1-initial-2", 1, base.Add(10*time.Second))
		p2a := newJudgment("project2-initial-1", 2, base.Add(5*time.Second))
		p2b := newJudgment("project2-initial-2", 2, base.Add(15*time.Second))

		jq.Push(p1a, defaultStrategy)
		jq.Push(p1b, defaultStrategy)
		jq.Push(p2a, defaultStrategy)
		jq.Push(p2b, defaultStrategy)

		if popped := jq.Pop(); popped.ID() != p1a.ID() {
			t.Fatalf("expected first pop to be earliest default item, got %s", popped.ID())
		}

		p1c := newJudgment("project1-reverse-1", 1, base.Add(20*time.Second))
		jq.Push(p1c, reverseTime)

		if popped := jq.Pop(); popped.ID() != p2a.ID() {
			t.Fatalf("expected next pop to honor project2 earliest item, got %s", popped.ID())
		}

		if popped := jq.Pop(); popped.ID() != p2b.ID() {
			t.Fatalf("expected project2 to finish before reverse project1 items, got %s", popped.ID())
		}

		if popped := jq.Pop(); popped.ID() != p1c.ID() {
			t.Fatalf("expected reverse strategy to pop newest item first, got %s", popped.ID())
		}

		if popped := jq.Pop(); popped.ID() != p1b.ID() {
			t.Fatalf("expected remaining project1 item to be oldest, got %s", popped.ID())
		}

		if jq.Len() != 0 {
			t.Fatalf("expected queue to be empty after pops, got len=%d", jq.Len())
		}
	})

	t.Run("push and pop with timeSince priority in same project", func(t *testing.T) {
		query := mocks.NewJudgmentQuery(t)
		timeSinceStrategy := prioritystrategies.NewTimeSinceLast(query)
		logger := logger.NewMockLogger()
		jq := queue.NewJudgeQueue("JUDGE", logger)

		judgments := []struct {
			id                 string
			projectID          int32
			createdAt          time.Time
			strategy           prioritystrategies.Strategy
			prevSubmissionTime time.Time
		}{
			{"judgment 1", 1, time.Now(), timeSinceStrategy, time.Time{}},
			{"judgment 2", 1, time.Now().Add(time.Second), timeSinceStrategy, time.Now()},
			{"judgment 3", 1, time.Now().Add(2 * time.Second), timeSinceStrategy, time.Time{}},
			{"judgment 4", 1, time.Now().Add(3 * time.Second), timeSinceStrategy, time.Now().Add(2 * time.Second)},
			{"judgment 5", 1, time.Now().Add(4 * time.Second), timeSinceStrategy, time.Now().Add(3 * time.Second)},
			{"judgment 6", 1, time.Now().Add(5 * time.Second), timeSinceStrategy, time.Now().Add(2 * time.Second)},
			{"judgment 7", 1, time.Now().Add(6 * time.Second), timeSinceStrategy, time.Now().Add(5 * time.Second)},
		}

		for _, j := range judgments {
			query.On("GetPrevSubmissionTimeByJudgmentID", j.id).Return(j.prevSubmissionTime, nil)
			pushJudgmentToQueue(jq, newJudgment(j.id, j.projectID, j.createdAt), j.strategy)
		}

		sequence := []string{"judgment 1", "judgment 3", "judgment 2", "judgment 4", "judgment 6", "judgment 5", "judgment 7"}
		for _, j := range sequence {
			if jq.Pop().ID() != j {
				t.Error("Judgment not popped correctly")
			}
		}
	})

	t.Run("push and pop with timeSince priority in difference project", func(t *testing.T) {
		query := mocks.NewJudgmentQuery(t)
		timeSinceStrategy := prioritystrategies.NewTimeSinceLast(query)
		logger := logger.NewMockLogger()
		jq := queue.NewJudgeQueue("JUDGE", logger)

		judgments := []struct {
			id                 string
			projectID          int32
			createdAt          time.Time
			strategy           prioritystrategies.Strategy
			prevSubmissionTime time.Time
		}{
			{"judgment 1", 1, time.Now(), timeSinceStrategy, time.Time{}},
			{"judgment 2", 1, time.Now().Add(time.Second), timeSinceStrategy, time.Now()},
			{"judgment 3", 2, time.Now().Add(2 * time.Second), timeSinceStrategy, time.Time{}},
			{"judgment 4", 1, time.Now().Add(3 * time.Second), timeSinceStrategy, time.Now().Add(2 * time.Second)},
			{"judgment 5", 2, time.Now().Add(4 * time.Second), timeSinceStrategy, time.Now().Add(3 * time.Second)},
			{"judgment 6", 2, time.Now().Add(5 * time.Second), timeSinceStrategy, time.Now().Add(2 * time.Second)},
			{"judgment 7", 1, time.Now().Add(6 * time.Second), timeSinceStrategy, time.Now().Add(5 * time.Second)},
		}

		for _, j := range judgments {
			query.On("GetPrevSubmissionTimeByJudgmentID", j.id).Return(j.prevSubmissionTime, nil)
			pushJudgmentToQueue(jq, newJudgment(j.id, j.projectID, j.createdAt), j.strategy)
		}

		sequence := []string{"judgment 1", "judgment 2", "judgment 3", "judgment 4", "judgment 6", "judgment 5", "judgment 7"}
		for _, j := range sequence {
			if jq.Pop().ID() != j {
				t.Error("Judgment not popped correctly")
			}
		}
	})

	t.Run("timeSince prioritizes first submissions over repeated submissions in same project when previous submission is missing", func(t *testing.T) {
		query := mocks.NewJudgmentQuery(t)
		timeSinceStrategy := prioritystrategies.NewTimeSinceLast(query)
		logger := logger.NewMockLogger()
		jq := queue.NewJudgeQueue("JUDGE", logger)

		base := time.Now()
		judgments := []struct {
			id        string
			createdAt time.Time
			prevTime  time.Time
			err       error
		}{
			{"repeat-1", base, base.Add(-2 * time.Minute), nil},
			{"repeat-2", base.Add(time.Second), base.Add(-time.Minute), nil},
			{"first-1", base.Add(2 * time.Second), time.Time{}, nil},
			{"first-2", base.Add(3 * time.Second), time.Time{}, nil},
		}

		for _, j := range judgments {
			query.On("GetPrevSubmissionTimeByJudgmentID", j.id).Return(j.prevTime, j.err)
			pushJudgmentToQueue(jq, newJudgment(j.id, 1, j.createdAt), timeSinceStrategy)
		}

		sequence := []string{"first-1", "first-2", "repeat-1", "repeat-2"}
		for _, id := range sequence {
			if jq.Pop().ID() != id {
				t.Fatalf("expected %s to pop next", id)
			}
		}
	})

	t.Run("push and pop with difference priority and project", func(t *testing.T) {
		query := mocks.NewJudgmentQuery(t)
		defaultStrategy := prioritystrategies.NewDefault()
		timeSinceStrategy := prioritystrategies.NewTimeSinceLast(query)
		logger := logger.NewMockLogger()
		jq := queue.NewJudgeQueue("JUDGE", logger)

		judgments := []struct {
			id                 string
			projectID          int32
			createdAt          time.Time
			strategy           prioritystrategies.Strategy
			prevSubmissionTime time.Time
		}{
			{"judgment 1", 1, time.Now(), defaultStrategy, time.Time{}},
			{"judgment 2", 1, time.Now().Add(time.Second), defaultStrategy, time.Now()},
			{"judgment 3", 2, time.Now().Add(2 * time.Second), timeSinceStrategy, time.Now().Add(1 * time.Second)},
			{"judgment 4", 1, time.Now().Add(3 * time.Second), defaultStrategy, time.Now().Add(2 * time.Second)},
			{"judgment 5", 2, time.Now().Add(4 * time.Second), timeSinceStrategy, time.Time{}},
			{"judgment 6", 2, time.Now().Add(5 * time.Second), timeSinceStrategy, time.Now().Add(2 * time.Second)},
			{"judgment 7", 2, time.Now().Add(6 * time.Second), timeSinceStrategy, time.Now().Add(1 * time.Second)},
		}

		for _, j := range judgments {
			if j.strategy == timeSinceStrategy {
				query.On("GetPrevSubmissionTimeByJudgmentID", j.id).Return(j.prevSubmissionTime, nil)
			}
			pushJudgmentToQueue(jq, newJudgment(j.id, j.projectID, j.createdAt), j.strategy)
		}

		sequence := []string{"judgment 1", "judgment 2", "judgment 4", "judgment 5", "judgment 3", "judgment 7", "judgment 6"}
		for _, j := range sequence {
			if jq.Pop().ID() != j {
				t.Error("Judgment not popped correctly")
			}
		}
	})

	t.Run("push and pop with difference priority and project - 2", func(t *testing.T) {
		query := mocks.NewJudgmentQuery(t)
		defaultStrategy := prioritystrategies.NewDefault()
		timeSinceStrategy := prioritystrategies.NewTimeSinceLast(query)
		logger := logger.NewMockLogger()
		jq := queue.NewJudgeQueue("JUDGE", logger)

		judgments := []struct {
			id                 string
			projectID          int32
			createdAt          time.Time
			strategy           prioritystrategies.Strategy
			prevSubmissionTime time.Time
		}{
			{"judgment 1", 1, time.Now(), defaultStrategy, time.Time{}},
			{"judgment 2", 2, time.Now(), &reverseTimeStrategy{}, time.Time{}},
			{"judgment 3", 3, time.Now(), defaultStrategy, time.Time{}},
			{"judgment 4", 1, time.Now(), defaultStrategy, time.Time{}},
			{"judgment 5", 2, time.Now(), &reverseTimeStrategy{}, time.Time{}},
			{"judgment 6", 3, time.Now(), defaultStrategy, time.Time{}},
			{"judgment 7", 1, time.Now(), defaultStrategy, time.Time{}},
			{"judgment 8", 2, time.Now(), &reverseTimeStrategy{}, time.Time{}},
			{"judgment 9", 3, time.Now(), defaultStrategy, time.Time{}},
		}

		for _, j := range judgments {
			if j.strategy == timeSinceStrategy {
				query.On("GetPrevSubmissionTimeByJudgmentID", j.id).Return(j.prevSubmissionTime, nil)
			}
			pushJudgmentToQueue(jq, newJudgment(j.id, j.projectID, j.createdAt), j.strategy)
		}

		sequence := []string{"judgment 1", "judgment 3", "judgment 4", "judgment 6", "judgment 7", "judgment 8", "judgment 5", "judgment 2", "judgment 9"}
		for _, j := range sequence {
			if jq.Pop().ID() != j {
				t.Error("Judgment not popped correctly")
			}
		}
	})

	t.Run("push and pop with lazy-update priority strategy", func(t *testing.T) {
		defaultStrategy := prioritystrategies.NewDefault()
		reverseTimeStrategy := &reverseTimeStrategy{}
		logger := logger.NewMockLogger()
		jq := queue.NewJudgeQueue("JUDGE", logger)

		judgments := []struct {
			id                 string
			projectID          int32
			createdAt          time.Time
			strategy           prioritystrategies.Strategy
			prevSubmissionTime time.Time
		}{
			{"judgment 1", 1, time.Now(), defaultStrategy, time.Time{}},
			{"judgment 2", 2, time.Now(), reverseTimeStrategy, time.Time{}},
			{"judgment 3", 3, time.Now(), defaultStrategy, time.Time{}},
			{"judgment 4", 1, time.Now(), defaultStrategy, time.Time{}},
			{"judgment 5", 2, time.Now(), reverseTimeStrategy, time.Time{}},
			{"judgment 6", 3, time.Now(), defaultStrategy, time.Time{}},
			{"judgment 7", 1, time.Now(), defaultStrategy, time.Time{}},
			{"judgment 8", 2, time.Now(), reverseTimeStrategy, time.Time{}},
			{"judgment 9", 3, time.Now(), defaultStrategy, time.Time{}},
		}

		fmt.Println("--------------------")
		for _, j := range judgments {
			pushJudgmentToQueue(jq, newJudgment(j.id, j.projectID, j.createdAt), j.strategy)
		}

		if jq.Pop().ID() != "judgment 1" {
			t.Error("Judgment not popped correctly")
		}

		pushJudgmentToQueue(jq, newJudgment("judgment 10", 2, time.Now()), defaultStrategy)
		if jq.Pop().ID() != "judgment 2" {
			t.Error("Judgment not popped correctly")
		}

		pushJudgmentToQueue(jq, newJudgment("judgment 11", 2, time.Now()), reverseTimeStrategy)

		sequence := []string{"judgment 3", "judgment 4", "judgment 6", "judgment 7", "judgment 9", "judgment 11", "judgment 10", "judgment 8", "judgment 5"}
		for _, j := range sequence {
			if jq.Pop().ID() != j {
				t.Error("Judgment not popped correctly")
			}
		}
	})

	t.Run("massive push and pop with same strategy", func(t *testing.T) {
		defaultStrategy := prioritystrategies.NewDefault()
		reverseTimeStrategy := &reverseTimeStrategy{}
		projectStrategies := []prioritystrategies.Strategy{
			defaultStrategy,
			reverseTimeStrategy,
			defaultStrategy,
			defaultStrategy,
			reverseTimeStrategy,
		}

		logger := logger.NewMockLogger()
		jq := queue.NewJudgeQueue("JUDGE", logger)
		id := 0
		for projectID := 0; projectID < len(projectStrategies); projectID++ {
			for i := 0; i < 10000; i++ {
				id++
				j := newJudgment(fmt.Sprintf("judgment %d", id), int32(projectID+1), time.Now().Add(time.Duration(id)*time.Second))
				pushJudgmentToQueue(jq, j, projectStrategies[projectID])
			}
		}

		sequence := []string{}
		for jq.Len() > 0 {
			id := jq.Pop().ID()
			sequence = append(sequence, id)
		}

		expected := []string{}
		for i := 1; i <= 10000; i++ {
			expected = append(expected, fmt.Sprintf("judgment %d", i))
		}
		for i := 20000; i > 10000; i-- {
			expected = append(expected, fmt.Sprintf("judgment %d", i))
		}
		for i := 20001; i <= 30000; i++ {
			expected = append(expected, fmt.Sprintf("judgment %d", i))
		}
		for i := 30001; i <= 40000; i++ {
			expected = append(expected, fmt.Sprintf("judgment %d", i))
		}
		for i := 50000; i > 40000; i-- {
			expected = append(expected, fmt.Sprintf("judgment %d", i))
		}

		if len(sequence) != len(expected) {
			t.Errorf("Expected %d judgments, got %d", len(expected), len(sequence))
		}

		for i := 0; i < len(sequence); i++ {
			if sequence[i] != expected[i] {
				t.Errorf("Expected %s, got %s", expected[i], sequence[i])
			}
		}
	})

	t.Run("concurrent push and pop does not race or deadlock", func(t *testing.T) {
		defaultStrategy := prioritystrategies.NewDefault()
		logger := logger.NewMockLogger()
		jq := queue.NewJudgeQueue("JUDGE", logger)

		const goroutines = 10
		const perGoroutine = 100

		var wg sync.WaitGroup
		for g := 0; g < goroutines; g++ {
			wg.Add(1)
			go func(g int) {
				defer wg.Done()
				for i := 0; i < perGoroutine; i++ {
					id := fmt.Sprintf("j-g%d-i%d", g, i)
					projectID := int32(g%3 + 1)
					jq.Push(newJudgment(id, projectID, time.Now()), defaultStrategy)
				}
			}(g)
		}
		wg.Wait()

		total := 0
		for jq.Len() > 0 {
			j := jq.Pop()
			if j != nil {
				total++
			}
		}

		if total != goroutines*perGoroutine {
			t.Errorf("expected %d judgments, popped %d", goroutines*perGoroutine, total)
		}
	})

	// Regression test for A5: Push used to call priorityStrategy.GetExtraFields
	// (which can do DB I/O, e.g. TimeSinceLast) while holding jq.mu, so a slow
	// lookup for one judgment serialized every other Push/Pop/Len call on the
	// entire queue behind that latency.
	t.Run("push does not hold the lock during GetExtraFields I/O", func(t *testing.T) {
		logger := logger.NewMockLogger()
		jq := queue.NewJudgeQueue("JUDGE", logger)

		slow := &blockingStrategy{
			started: make(chan struct{}),
			unblock: make(chan struct{}),
		}

		pushDone := make(chan struct{})
		go func() {
			defer close(pushDone)
			jq.Push(newJudgment("slow-judgment", 1, time.Now()), slow)
		}()

		select {
		case <-slow.started:
		case <-time.After(2 * time.Second):
			t.Fatal("Push never reached GetExtraFields")
		}

		// Push is now blocked inside GetExtraFields. If it were holding
		// jq.mu across that call, Len() would hang until unblock is closed.
		lenDone := make(chan int, 1)
		go func() {
			lenDone <- jq.Len()
		}()
		select {
		case n := <-lenDone:
			if n != 0 {
				t.Errorf("expected queue to still be empty, got Len()=%d", n)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("Len() was blocked by Push's in-flight GetExtraFields call — the lock is held during I/O")
		}

		close(slow.unblock)
		select {
		case <-pushDone:
		case <-time.After(2 * time.Second):
			t.Fatal("Push did not complete after GetExtraFields unblocked")
		}

		if got := jq.Len(); got != 1 {
			t.Errorf("expected 1 judgment queued after Push completed, got %d", got)
		}
	})

	t.Run("push and pop effectively with different strategies", func(t *testing.T) {
		defaultStrategy := prioritystrategies.NewDefault()
		reverseTimeStrategy := &reverseTimeStrategy{}

		logger := logger.NewMockLogger()
		jq := queue.NewJudgeQueue("JUDGE", logger)
		id := 0
		for projectID := 0; projectID < 100; projectID++ {
			for i := 0; i < 10000; i++ {
				id++
				j := newJudgment(fmt.Sprintf("judgment %d", id), int32(projectID+1), time.Now().Add(time.Duration(id)*time.Second))
				pushJudgmentToQueue(jq, j, defaultStrategy)
			}
		}

		start := time.Now()
		pushJudgmentToQueue(jq, newJudgment("judgment", 4, time.Now().Add(12345*time.Second)), reverseTimeStrategy)
		end := time.Now()
		t.Logf("Push took %v\n", end.Sub(start))
	})
}
