package query

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The comparisons this pairs up are what an admin reads before approving or
// rejecting a rejudge, so a "before" attached to the wrong submission is a
// wrong-score bug, not a cosmetic one. These cases are deliberately
// order-independent: they feed the two inputs in orders that do not correspond
// to each other, which is exactly the situation the two source queries produce
// and which a positional pairing cannot survive.

func at(day int) time.Time {
	return time.Date(2026, 7, day, 0, 0, 0, 0, time.UTC)
}

func TestPairSubmissionComparisons(t *testing.T) {
	t.Run("pairs each submission with its own baseline regardless of input order", func(t *testing.T) {
		afters := []judgmentOutcomeRow{
			{ID: "after-a", SubmissionID: "sub-a", Score: 50, TotalScore: 100, Status: "Rejudged", CreatedAt: at(2)},
			{ID: "after-b", SubmissionID: "sub-b", Score: 80, TotalScore: 100, Status: "Rejudged", CreatedAt: at(2)},
		}
		// Reversed relative to afters: the orderings of the two queries are
		// unrelated, so this is a legitimate input.
		priors := []judgmentOutcomeRow{
			{ID: "prior-b", SubmissionID: "sub-b", Score: 20, TotalScore: 100, Status: "Judged", CreatedAt: at(1)},
			{ID: "prior-a", SubmissionID: "sub-a", Score: 100, TotalScore: 100, Status: "Judged", CreatedAt: at(1)},
		}

		got := pairSubmissionComparisons(afters, priors)

		require.Len(t, got, 2)
		assert.Equal(t, "sub-a", got[0].ID)
		require.NotNil(t, got[0].Before)
		assert.Equal(t, float64(100), got[0].Before.Score, "sub-a must be compared against its own previous score")
		assert.Equal(t, float64(50), got[0].After.Score)

		assert.Equal(t, "sub-b", got[1].ID)
		require.NotNil(t, got[1].Before)
		assert.Equal(t, float64(20), got[1].Before.Score, "sub-b must be compared against its own previous score")
		assert.Equal(t, float64(80), got[1].After.Score)
	})

	t.Run("a submission with no baseline gets nil, and no other submission's baseline", func(t *testing.T) {
		// Only one of the two submissions has a previous judgment, so the two
		// inputs have different lengths. A positional pairing hands the second
		// submission either nothing or the first submission's score.
		afters := []judgmentOutcomeRow{
			{ID: "after-a", SubmissionID: "sub-a", Score: 50, TotalScore: 100, Status: "Rejudged", CreatedAt: at(2)},
			{ID: "after-b", SubmissionID: "sub-b", Score: 80, TotalScore: 100, Status: "Rejudged", CreatedAt: at(2)},
		}
		priors := []judgmentOutcomeRow{
			{ID: "prior-a", SubmissionID: "sub-a", Score: 100, TotalScore: 100, Status: "Judged", CreatedAt: at(1)},
		}

		got := pairSubmissionComparisons(afters, priors)

		require.Len(t, got, 2)
		require.NotNil(t, got[0].Before)
		assert.Equal(t, float64(100), got[0].Before.Score)
		assert.Nil(t, got[1].Before, "sub-b never had an adopted judgment, so it has no baseline to show")
	})

	t.Run("the submission without a baseline is not given its neighbour's score", func(t *testing.T) {
		// Same as above with the roles swapped, so a positional pairing lands the
		// other contestant's 100 on the submission that has no history at all.
		afters := []judgmentOutcomeRow{
			{ID: "after-c", SubmissionID: "sub-c", Score: 30, TotalScore: 100, Status: "Rejudged", CreatedAt: at(2)},
			{ID: "after-d", SubmissionID: "sub-d", Score: 70, TotalScore: 100, Status: "Rejudged", CreatedAt: at(2)},
		}
		priors := []judgmentOutcomeRow{
			{ID: "prior-d", SubmissionID: "sub-d", Score: 100, TotalScore: 100, Status: "Judged", CreatedAt: at(1)},
		}

		got := pairSubmissionComparisons(afters, priors)

		require.Len(t, got, 2)
		assert.Equal(t, "sub-c", got[0].ID)
		assert.Nil(t, got[0].Before, "sub-c has no history, so it must not inherit sub-d's previous score")
		assert.Equal(t, "sub-d", got[1].ID)
		require.NotNil(t, got[1].Before)
		assert.Equal(t, float64(100), got[1].Before.Score)
	})

	t.Run("picks the most recent baseline when a submission has several", func(t *testing.T) {
		afters := []judgmentOutcomeRow{
			{ID: "after-a", SubmissionID: "sub-a", Score: 50, TotalScore: 100, Status: "Rejudged", CreatedAt: at(9)},
		}
		priors := []judgmentOutcomeRow{
			{ID: "prior-old", SubmissionID: "sub-a", Score: 10, TotalScore: 100, Status: "Judged", CreatedAt: at(1)},
			{ID: "prior-new", SubmissionID: "sub-a", Score: 40, TotalScore: 100, Status: "Judged", CreatedAt: at(5)},
			{ID: "prior-mid", SubmissionID: "sub-a", Score: 25, TotalScore: 100, Status: "Judged", CreatedAt: at(3)},
		}

		got := pairSubmissionComparisons(afters, priors)

		require.Len(t, got, 1)
		require.NotNil(t, got[0].Before)
		assert.Equal(t, float64(40), got[0].Before.Score)
	})

	t.Run("ignores baselines belonging to submissions outside this rejudge", func(t *testing.T) {
		afters := []judgmentOutcomeRow{
			{ID: "after-a", SubmissionID: "sub-a", Score: 50, TotalScore: 100, Status: "Rejudged", CreatedAt: at(2)},
		}
		priors := []judgmentOutcomeRow{
			{ID: "prior-z", SubmissionID: "sub-z", Score: 99, TotalScore: 100, Status: "Judged", CreatedAt: at(1)},
		}

		got := pairSubmissionComparisons(afters, priors)

		require.Len(t, got, 1)
		assert.Equal(t, "sub-a", got[0].ID)
		assert.Nil(t, got[0].Before)
	})

	t.Run("orders by when the rejudge produced each judgment", func(t *testing.T) {
		afters := []judgmentOutcomeRow{
			{ID: "after-late", SubmissionID: "sub-late", Score: 1, Status: "Rejudged", CreatedAt: at(3)},
			{ID: "after-early", SubmissionID: "sub-early", Score: 2, Status: "Rejudged", CreatedAt: at(1)},
			{ID: "after-mid", SubmissionID: "sub-mid", Score: 3, Status: "Rejudged", CreatedAt: at(2)},
		}

		got := pairSubmissionComparisons(afters, nil)

		require.Len(t, got, 3)
		assert.Equal(t, []string{"sub-early", "sub-mid", "sub-late"},
			[]string{got[0].ID, got[1].ID, got[2].ID})
	})

	t.Run("falls back to submission id when timestamps tie, so the list is stable", func(t *testing.T) {
		afters := []judgmentOutcomeRow{
			{ID: "after-c", SubmissionID: "sub-c", Status: "Rejudged", CreatedAt: at(1)},
			{ID: "after-a", SubmissionID: "sub-a", Status: "Rejudged", CreatedAt: at(1)},
			{ID: "after-b", SubmissionID: "sub-b", Status: "Rejudged", CreatedAt: at(1)},
		}

		got := pairSubmissionComparisons(afters, nil)

		require.Len(t, got, 3)
		assert.Equal(t, []string{"sub-a", "sub-b", "sub-c"},
			[]string{got[0].ID, got[1].ID, got[2].ID})
	})

	t.Run("returns an empty list when the rejudge produced no judgments", func(t *testing.T) {
		got := pairSubmissionComparisons(nil, nil)
		assert.NotNil(t, got, "an empty list must serialise as [] rather than null")
		assert.Empty(t, got)
	})
}

// Once an admin approves a rejudge its judgment becomes Judged, which makes it
// eligible as a baseline — including for itself, and for rejudges that ran
// before it. These pin the timestamp anchoring that keeps each comparison fixed
// to what it actually replaced, whatever order the decisions land in.
func TestPairSubmissionComparisons_BaselineIsAnchoredInTime(t *testing.T) {
	t.Run("an approved earlier rejudge becomes the baseline for a later one", func(t *testing.T) {
		afters := []judgmentOutcomeRow{
			{ID: "rejudge-2", SubmissionID: "sub-e", Score: 70, TotalScore: 100, Status: "Rejudged", CreatedAt: at(3)},
		}
		priors := []judgmentOutcomeRow{
			{ID: "original", SubmissionID: "sub-e", Score: 10, TotalScore: 100, Status: "Judged", CreatedAt: at(1)},
			{ID: "rejudge-1", SubmissionID: "sub-e", Score: 40, TotalScore: 100, Status: "Judged", CreatedAt: at(2)},
		}

		got := pairSubmissionComparisons(afters, priors)

		require.Len(t, got, 1)
		require.NotNil(t, got[0].Before)
		assert.Equal(t, float64(40), got[0].Before.Score,
			"the contestant was carrying the approved first rejudge's score, not the original")
	})

	t.Run("an approved rejudge is not its own baseline", func(t *testing.T) {
		// The same judgment appears on both sides: it is this rejudge's result, and
		// approval has already flipped it to Judged.
		approved := judgmentOutcomeRow{ID: "rejudge-1", SubmissionID: "sub-e", Score: 40, TotalScore: 100, Status: "Judged", CreatedAt: at(2)}
		afters := []judgmentOutcomeRow{approved}
		priors := []judgmentOutcomeRow{
			{ID: "original", SubmissionID: "sub-e", Score: 10, TotalScore: 100, Status: "Judged", CreatedAt: at(1)},
			approved,
		}

		got := pairSubmissionComparisons(afters, priors)

		require.Len(t, got, 1)
		require.NotNil(t, got[0].Before)
		assert.Equal(t, float64(10), got[0].Before.Score,
			"an approved rejudge must still be compared against what it replaced")
		assert.Equal(t, float64(40), got[0].After.Score)
	})

	t.Run("a later approved rejudge does not become an earlier one's baseline", func(t *testing.T) {
		// Approved out of order: the second rejudge was decided first, so its
		// judgment is Judged while the first rejudge is still awaiting a decision.
		afters := []judgmentOutcomeRow{
			{ID: "rejudge-1", SubmissionID: "sub-e", Score: 40, TotalScore: 100, Status: "Rejudged", CreatedAt: at(2)},
		}
		priors := []judgmentOutcomeRow{
			{ID: "original", SubmissionID: "sub-e", Score: 10, TotalScore: 100, Status: "Judged", CreatedAt: at(1)},
			{ID: "rejudge-2", SubmissionID: "sub-e", Score: 70, TotalScore: 100, Status: "Judged", CreatedAt: at(3)},
		}

		got := pairSubmissionComparisons(afters, priors)

		require.Len(t, got, 1)
		require.NotNil(t, got[0].Before)
		assert.Equal(t, float64(10), got[0].Before.Score,
			"a judgment produced after this one cannot be what this one replaced")
	})
}
