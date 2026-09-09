package setup

import (
	"fmt"
	"net/http"
	"time"

	"judge-backend/internal/app/judge_domain/judge/domain/queue"
	"judge-backend/internal/app/judge_domain/judge/infrastructure/worker"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/config"
)

// JudgeSeamPortOffset is added to the test server port to derive the address
// of the test-seam HTTP server. Both processes (the in-process app started by
// tests/test.go and the `go test` process running the BDD scenarios) compute
// it from the same config, so they always agree.
const JudgeSeamPortOffset = 1

// RegisterJudgeTestSeam is an fx Invoke wired ONLY by the test server setup
// (never in production wiring). BDD scenarios run in a separate `go test`
// process and therefore cannot touch the app's in-memory judge queue or
// worker pool directly; this endpoint gives them a reset seam over HTTP:
//
//	POST /judge/reset → drain in-flight event handlers, drain all queues,
//	wait for workers to go idle, restart any worker that is not running.
//
// Responses: 204 when idle was reached, 504 when the drain/idle wait ran past
// budget (the caller should fail its scenario loudly in that case).
func RegisterJudgeTestSeam(
	eventbus sharedtypes.EventBus,
	queueProvider queue.JudgeQueueProvider,
	pool *worker.JudgeWorkerPool,
	cfg *config.Config,
	logger sharedtypes.Logger,
) {
	mux := http.NewServeMux()
	mux.HandleFunc("/judge/reset", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Publish only enqueues an async handler and returns immediately
		// (asaskevich/EventBus spawns a goroutine per handler per Publish);
		// it does NOT wait for that goroutine to run. So a judgment_created
		// event fired by the previous scenario (e.g. from a submit request
		// that already returned its HTTP response) can still be sitting
		// unprocessed at this point. If Reset() ran first, that handler
		// would push its judgment into the FRESH post-reset queue moments
		// later — a ghost judgment bleeding into the next scenario, which
		// Reset() has no way to catch since the push hasn't happened yet.
		// Draining every in-flight async handler first closes that gap: by
		// the time Reset() runs, that handler has either already pushed
		// (into the queue Reset() is about to replace) or was never going
		// to run before another WaitAsync call, and either way Reset() sees
		// a state it can fully own.
		if !waitAsyncWithTimeout(eventbus, 10*time.Second) {
			logger.Errorf("[JudgeTestSeam] In-flight event handlers did not drain within budget")
			w.WriteHeader(http.StatusGatewayTimeout)
			return
		}

		queueProvider.Reset()
		if !pool.WaitForIdle(10 * time.Second) {
			logger.Errorf("[JudgeTestSeam] Workers did not go idle within budget")
			w.WriteHeader(http.StatusGatewayTimeout)
			return
		}

		// Restore every worker to running. Stopping workers is how a scenario
		// makes "queued" an observable state instead of a race it has to catch
		// (see the judge-queue feature), but a stopped worker is process-wide
		// state: one app serves the whole run, packages run sequentially
		// against it, and nothing else ever turns a worker back on. A scenario
		// that stops the pool and then fails before restarting it would
		// otherwise leave every later scenario in the run judging nothing.
		// Doing it here rather than in that suite's cleanup means it holds even
		// when the scenario dies partway.
		//
		// Start is a no-op unless the worker is Stopped or Error, so this
		// neither disturbs a healthy pool nor races the idle wait above — the
		// queues are already empty by this point, so a revived worker has
		// nothing to pick up.
		for _, w := range pool.ListJudgeWorkers() {
			w.Start()
		}

		// A worker that was still executing a (pre-reset) judgment when
		// WaitForIdle started publishes its terminal judged/judgeupdated
		// events synchronously, inside Execute, BEFORE it clears
		// currentJudgeID — so by the time WaitForIdle observes it as idle,
		// those events have already been Published (and are therefore
		// already counted in the eventbus's wait group). Their handlers
		// (rejudge status rollup, contest-side SSE updates, ...) are still
		// async, though, and can still be writing to the DB after this
		// request returns 204 — right as the next scenario's own DB reset
		// runs. Drain them too before declaring the reset done.
		if !waitAsyncWithTimeout(eventbus, 10*time.Second) {
			logger.Errorf("[JudgeTestSeam] Terminal event handlers did not drain within budget")
			w.WriteHeader(http.StatusGatewayTimeout)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	addr := fmt.Sprintf("localhost:%d", cfg.Test.Server.Port+JudgeSeamPortOffset)
	go func() {
		if err := http.ListenAndServe(addr, mux); err != nil {
			logger.Errorf("[JudgeTestSeam] Seam server stopped: %v", err)
		}
	}()
	logger.Infof("[JudgeTestSeam] Listening on %s", addr)
}

// waitAsyncWithTimeout bounds sharedtypes.EventBus.WaitAsync, which has no
// timeout of its own: it blocks until every in-flight async handler (and
// anything they transitively publish) has returned. Returns false if that
// doesn't happen within timeout — a stuck handler shouldn't be able to hang
// this endpoint forever.
func waitAsyncWithTimeout(eventbus sharedtypes.EventBus, timeout time.Duration) bool {
	done := make(chan struct{})
	go func() {
		eventbus.WaitAsync()
		close(done)
	}()
	select {
	case <-done:
		return true
	case <-time.After(timeout):
		return false
	}
}
