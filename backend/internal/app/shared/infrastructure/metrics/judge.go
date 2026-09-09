package metrics

import (
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Judge-pipeline metrics. Instrumented from the judge domain's
// infrastructure layer only (queue, worker, docker runtime) — the
// application layer stays free of metrics concerns.
var (
	// JudgeQueueDepth is the number of judgments currently waiting in each
	// judge-type queue (all per-project sub-queues summed). The queue itself
	// keeps this in sync on every push/pop/reset.
	JudgeQueueDepth = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "judge_queue_depth",
		Help: "Number of judgments currently queued, by judge type.",
	}, []string{"judge_type"})

	// ActiveJudgments is how many judgments are being executed right now —
	// equivalently, how many workers are busy.
	ActiveJudgments = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "judge_active_judgments",
		Help: "Number of judgments currently executing (busy workers).",
	})

	// JudgmentsFinishedTotal counts judgments by terminal status
	// (Judged/Rejudged/Error/Cancelled).
	JudgmentsFinishedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "judge_judgments_finished_total",
		Help: "Total judgments finished, by terminal status.",
	}, []string{"result"})

	// JudgmentDuration observes wall-clock execution time per judgment.
	// Buckets stretch to the configured judgment budget scale (minutes).
	JudgmentDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "judge_judgment_duration_seconds",
		Help:    "Judgment wall-clock duration in seconds, by terminal status.",
		Buckets: []float64{10, 30, 60, 120, 300, 600, 900, 1800, 2700},
	}, []string{"result"})

	// ProvisionAttemptsTotal counts environment-provisioning attempts. A
	// rising failure rate is the early-warning signal for host overload
	// (containers missing their HEALTHCHECK window).
	ProvisionAttemptsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "judge_provision_attempts_total",
		Help: "Total environment provisioning attempts, by result.",
	}, []string{"result"})
)

// activeJudgments mirrors the ActiveJudgments gauge in a form the application
// can read back. A Prometheus gauge is write-only outside of test helpers, and
// the admin queue-status endpoint has to report the same number Grafana plots —
// so both are moved by the same pair of functions rather than being counted a
// second way (e.g. by tallying busy workers, which also counts the brief
// pop-then-claim window and would disagree).
var activeJudgments atomic.Int64

func SetJudgeQueueDepth(judgeType string, depth int) {
	JudgeQueueDepth.WithLabelValues(judgeType).Set(float64(depth))
}

// IncActiveJudgments marks one judgment as executing.
func IncActiveJudgments() {
	ActiveJudgments.Inc()
	activeJudgments.Add(1)
}

// DecActiveJudgments marks one judgment as no longer executing.
func DecActiveJudgments() {
	ActiveJudgments.Dec()
	activeJudgments.Add(-1)
}

// ActiveJudgmentCount is how many judgments are executing right now.
func ActiveJudgmentCount() int {
	count := activeJudgments.Load()
	if count < 0 {
		// Defensive: a negative count would surface as a nonsensical reading
		// on the status dashboard rather than as the accounting bug it is.
		return 0
	}
	return int(count)
}

func ObserveJudgmentFinished(result string, seconds float64) {
	JudgmentsFinishedTotal.WithLabelValues(result).Inc()
	JudgmentDuration.WithLabelValues(result).Observe(seconds)
}

func ObserveProvisionAttempt(success bool) {
	result := "success"
	if !success {
		result = "failure"
	}
	ProvisionAttemptsTotal.WithLabelValues(result).Inc()
}
