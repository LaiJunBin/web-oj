package rejudge_test

import (
	"testing"

	"judge-backend/internal/app/contest_domain/activity/domain/entity/rejudge"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Whether a status counts as "awaiting a decision" decides whether a new rejudge
// request is refused, so a status drifting to the wrong side of this line either
// blocks legitimate rejudges forever or lets two of them be decided out of order.
func TestStatus_IsAwaitingDecision(t *testing.T) {
	cases := []struct {
		status   string
		awaiting bool
		why      string
	}{
		{"Pending", true, "queued but not run"},
		{"Running", true, "still judging"},
		{"Finished", true, "judged, but no admin has accepted or discarded it"},
		{"Approved", false, "the result was accepted and is now the score"},
		{"Rejected", false, "the result was discarded"},
		{"Canceled", false, "the run was abandoned"},
		{"Error", false, "the run failed and will never produce a score"},
	}

	for _, c := range cases {
		t.Run(c.status, func(t *testing.T) {
			r, err := rejudge.New(&rejudge.RejudgeConfig{ID: "rejudge-1", Status: c.status})
			require.NoError(t, err)

			assert.Equal(t, c.awaiting, r.Status().IsAwaitingDecision(), c.why)
		})
	}
}

// The repository filters Mongo on these strings, so they have to stay in step with
// IsAwaitingDecision — a mismatch would make the guard silently query for the
// wrong set.
func TestAwaitingDecisionStatuses_MatchesIsAwaitingDecision(t *testing.T) {
	all := []string{"Pending", "Running", "Finished", "Approved", "Rejected", "Canceled", "Error"}
	listed := map[string]struct{}{}
	for _, s := range rejudge.AwaitingDecisionStatuses() {
		listed[s] = struct{}{}
	}

	for _, status := range all {
		r, err := rejudge.New(&rejudge.RejudgeConfig{ID: "rejudge-1", Status: status})
		require.NoError(t, err)

		_, inList := listed[status]
		assert.Equal(t, r.Status().IsAwaitingDecision(), inList,
			"%q must appear in AwaitingDecisionStatuses() exactly when IsAwaitingDecision() reports it", status)
	}

	assert.Len(t, rejudge.AwaitingDecisionStatuses(), 3, "guards against a status being added to the list by accident")
}
