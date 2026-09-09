package query

import (
	"context"
	"errors"
	"judge-backend/internal/app/contest_domain/activity/application/query"
	"judge-backend/internal/app/contest_domain/activity/application/query/dto"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// judgedStatus is the judgment status that means "this is the score the
// contestant is carrying". Spelled out here rather than imported: the judge
// domain owns that entity, and reaching into it from another bounded context is
// the coupling this codebase already decided against.
const judgedStatus = "Judged"

type RejudgeQuery struct {
	mdb *mongo.Client
}

func NewRejudgeQuery(
	mdb *mongo.Client,
) query.RejudgeQuery {
	return &RejudgeQuery{
		mdb: mdb,
	}
}

type RejudgeInfo struct {
	ID                 string    `bson:"_id"`
	ContestProjectID   int32     `bson:"contestProjectID"`
	Reason             string    `bson:"reason"`
	CreatedAt          time.Time `bson:"createdAt"`
	Status             string    `bson:"status"`
	EffectedJudgments  int32     `bson:"effectedJudgments"`
	CompletedJudgments int32     `bson:"completedJudgments"`
}

// terminalJudgmentStatuses are the judgment statuses that mean "this judgment
// is not going to change on its own again", which is what makes a rejudge's
// progress bar move. Rejudged is the normal landing state; Judged is where an
// approved one ends up; Error and Cancelled are finished too, just not
// successfully. Spelled out here rather than imported for the same reason
// judgedStatus is.
var terminalJudgmentStatuses = []string{"Rejudged", "Judged", "Error", "Cancelled"}

func (q *RejudgeQuery) ListRejudges() ([]*dto.RejudgeInfo, error) {
	return q.listRejudges(nil)
}

// ListRejudgesByContestProjectID is the same listing narrowed to one contest
// project. The filter is a $match on the rejudge's own contestProjectID rather
// than a join through its judgments: a rejudge raised against a project with no
// submissions has no judgments to join through, and its project would otherwise
// be unknowable.
func (q *RejudgeQuery) ListRejudgesByContestProjectID(contestProjectID int32) ([]*dto.RejudgeInfo, error) {
	return q.listRejudges(bson.M{"contestProjectID": contestProjectID})
}

func (q *RejudgeQuery) listRejudges(match bson.M) ([]*dto.RejudgeInfo, error) {
	ctx := context.Background()
	collection := q.mdb.Database("judge").Collection("rejudges")

	pipeline := mongo.Pipeline{}
	if match != nil {
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: match}})
	}
	// $size over the judgments array rather than $unwind-then-count.
	//
	// $unwind drops a document whose array is empty, so a rejudge that produced
	// no judgments used to disappear from this listing entirely — and one of the
	// ways that happens is a genuine failure: the request handler logs and
	// continues when a judgment cannot be created, then saves the rejudge
	// regardless, so a run where every creation failed leaves a rejudge with an
	// empty array. Nothing advances its status either (the rollup is driven by
	// judgments finishing, and there are none), so it sits Pending forever while
	// being invisible to the admin who asked for it. Reporting it with a count
	// of zero is both the honest answer and the one that lets them cancel it.
	//
	// Cheaper, too: $unwind expanded a rejudge covering three hundred
	// submissions into three hundred documents just to group them back again.
	pipeline = append(pipeline,
		bson.D{
			{
				Key: "$project",
				Value: bson.M{
					"_id":              1,
					"contestProjectID": 1,
					"reason":           1,
					"createdAt":        1,
					"status":           1,
					// $ifNull guards a document written before the field
					// existed; the repository always writes an array.
					"effectedJudgments": bson.M{"$size": bson.M{"$ifNull": bson.A{"$judgments", bson.A{}}}},
				},
			},
		},
		bson.D{
			{
				Key: "$sort",
				Value: bson.M{
					"createdAt": -1,
				},
			},
		},
	)

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var result []RejudgeInfo
	if err := cursor.All(ctx, &result); err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return []*dto.RejudgeInfo{}, nil
	}

	rejudgeIDs := make([]string, 0, len(result))
	for _, v := range result {
		rejudgeIDs = append(rejudgeIDs, v.ID)
	}
	completed, err := q.countCompletedJudgments(ctx, rejudgeIDs)
	if err != nil {
		return nil, err
	}

	rejudges := []*dto.RejudgeInfo{}
	for _, v := range result {
		rejudges = append(rejudges, &dto.RejudgeInfo{
			ID:                 v.ID,
			ContestProjectID:   v.ContestProjectID,
			Reason:             v.Reason,
			CreatedAt:          v.CreatedAt,
			Status:             v.Status,
			EffectedJudgments:  v.EffectedJudgments,
			CompletedJudgments: completed[v.ID],
		})
	}

	return rejudges, nil
}

// countCompletedJudgments returns, per rejudge id, how many of its judgments
// have reached a terminal state.
//
// A second query grouped on the judgments side rather than a $lookup inside the
// listing pipeline: a $lookup per rejudge against the whole judgments
// collection on every status-page poll is dearer than one grouped pass here, and
// it follows what GetRejudgeByID below already does for the same reason —
// several plain queries joined in Go beat one clever pipeline here.
func (q *RejudgeQuery) countCompletedJudgments(ctx context.Context, rejudgeIDs []string) (map[string]int32, error) {
	completed := map[string]int32{}
	if len(rejudgeIDs) == 0 {
		return completed, nil
	}

	collection := q.mdb.Database("judge").Collection("judgments")
	cursor, err := collection.Aggregate(ctx, mongo.Pipeline{
		bson.D{
			{
				Key: "$match",
				Value: bson.M{
					"rejudgeID": bson.M{"$in": rejudgeIDs},
					"status":    bson.M{"$in": terminalJudgmentStatuses},
				},
			},
		},
		bson.D{
			{
				Key: "$group",
				Value: bson.M{
					"_id":   "$rejudgeID",
					"count": bson.M{"$sum": 1},
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rows []struct {
		RejudgeID string `bson:"_id"`
		Count     int32  `bson:"count"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, err
	}

	for _, row := range rows {
		completed[row.RejudgeID] = row.Count
	}
	return completed, nil
}

type Rejudge struct {
	ID        string    `bson:"_id"`
	Reason    string    `bson:"reason"`
	Status    string    `bson:"status"`
	CreatedAt time.Time `bson:"createdAt"`
}

// judgmentOutcomeRow is one judgment reduced to what a before/after comparison
// needs.
//
// SubmissionID is carried deliberately. An earlier version of this query built
// the comparisons inside the aggregation by zipping two arrays on a shared
// index, and projected submissionID away because the index was doing the
// pairing. Those two arrays come from different $lookup stages whose orderings
// are unrelated ($group does not preserve order), and a submission with no
// previous judgment contributes no element at all — so the arrays are not even
// the same length. The result was that one contestant could be shown another
// contestant's previous score, which is the number an admin approves or rejects
// a rejudge on. The pairing now happens in Go, keyed by this field.
type judgmentOutcomeRow struct {
	ID           string    `bson:"_id"`
	SubmissionID string    `bson:"submissionID"`
	Score        float64   `bson:"score"`
	TotalScore   float64   `bson:"totalScore"`
	Status       string    `bson:"status"`
	CreatedAt    time.Time `bson:"createdAt"`
}

func (r judgmentOutcomeRow) outcome() dto.JudgmentOutcomeChange {
	return dto.JudgmentOutcomeChange{
		Status:     r.Status,
		Score:      r.Score,
		TotalScore: r.TotalScore,
	}
}

// pairSubmissionComparisons matches every judgment a rejudge produced ("after")
// with the judgment that submission had adopted beforehand ("before"), keyed by
// submission id. A submission with no adopted judgment before the rejudge gets a
// nil Before rather than a zero-valued one.
//
// `priors` may hold several judgments per submission; the baseline is the most
// recent one that predates this rejudge's own judgment. Anchoring on the
// timestamp rather than just taking the latest is what keeps a comparison stable
// once decisions land: approving a rejudge flips its judgment to Judged, which
// would otherwise make it its own baseline and collapse the diff to no change,
// and approving a LATER rejudge would otherwise reach backwards and become the
// baseline for an earlier one.
func pairSubmissionComparisons(afters, priors []judgmentOutcomeRow) []*dto.SubmissionComparison {
	priorsBySubmission := map[string][]judgmentOutcomeRow{}
	for _, prior := range priors {
		priorsBySubmission[prior.SubmissionID] = append(priorsBySubmission[prior.SubmissionID], prior)
	}

	// The sort key travels with its comparison: sorting the comparisons while
	// reading the key out of `afters` by index would desynchronise the two on the
	// first swap.
	type ordered struct {
		comparison *dto.SubmissionComparison
		judgedAt   time.Time
	}

	rows := make([]ordered, 0, len(afters))
	for _, after := range afters {
		comparison := &dto.SubmissionComparison{
			ID:    after.SubmissionID,
			After: after.outcome(),
		}

		var baseline *judgmentOutcomeRow
		candidates := priorsBySubmission[after.SubmissionID]
		for i := range candidates {
			if candidates[i].ID == after.ID || !candidates[i].CreatedAt.Before(after.CreatedAt) {
				continue
			}
			if baseline == nil || candidates[i].CreatedAt.After(baseline.CreatedAt) {
				baseline = &candidates[i]
			}
		}
		if baseline != nil {
			outcome := baseline.outcome()
			comparison.Before = &outcome
		}

		rows = append(rows, ordered{comparison: comparison, judgedAt: after.CreatedAt})
	}

	// Neither query above promises an order, and an unstable list makes the page
	// reshuffle between refreshes. Order by when the rejudge produced each
	// judgment, falling back to the submission id so equal timestamps still land
	// somewhere fixed.
	sort.SliceStable(rows, func(i, j int) bool {
		if !rows[i].judgedAt.Equal(rows[j].judgedAt) {
			return rows[i].judgedAt.Before(rows[j].judgedAt)
		}
		return rows[i].comparison.ID < rows[j].comparison.ID
	})

	comparisons := make([]*dto.SubmissionComparison, 0, len(rows))
	for _, row := range rows {
		comparisons = append(comparisons, row.comparison)
	}

	return comparisons
}

// findJudgmentOutcomes reads the judgments matching filter down to the fields a
// comparison needs.
func (q *RejudgeQuery) findJudgmentOutcomes(ctx context.Context, filter bson.M) ([]judgmentOutcomeRow, error) {
	collection := q.mdb.Database("judge").Collection("judgments")
	projection := bson.M{
		"_id":          1,
		"submissionID": 1,
		"score":        1,
		"totalScore":   1,
		"status":       1,
		"createdAt":    1,
	}

	cursor, err := collection.Find(ctx, filter, options.Find().SetProjection(projection))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	rows := []judgmentOutcomeRow{}
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, err
	}
	return rows, nil
}

// GetRejudgeByID returns the rejudge together with a before/after outcome for
// every submission it touched.
//
// Deliberately three plain queries rather than one aggregation. The pairing has
// to be by submission id (see judgmentOutcomeRow), and picking each submission's
// baseline also depends on that submission's own "after" timestamp — a
// correlation that turns into an awkward nested sub-pipeline in the aggregation
// but reads as a few lines of Go. All three queries are unindexed collection
// scans either way, so this costs round trips, not scans.
func (q *RejudgeQuery) GetRejudgeByID(id string) (*dto.Rejudge, error) {
	ctx := context.Background()

	var rejudge Rejudge
	err := q.mdb.Database("judge").Collection("rejudges").
		FindOne(ctx, bson.M{"_id": id}).Decode(&rejudge)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, httperror.New(http.StatusNotFound, "rejudge not found")
	}
	if err != nil {
		return nil, err
	}

	afters, err := q.findJudgmentOutcomes(ctx, bson.M{"rejudgeID": id})
	if err != nil {
		return nil, err
	}

	submissionIDs := make([]string, 0, len(afters))
	for _, after := range afters {
		submissionIDs = append(submissionIDs, after.SubmissionID)
	}

	priors := []judgmentOutcomeRow{}
	if len(submissionIDs) > 0 {
		// Only adopted judgments can be a baseline, and status is the whole test —
		// judgeType deliberately is not. A rejudge's judgments sit in Rejudged until
		// an admin approves them, and rejected ones stay there forever, so neither
		// was ever the contestant's score. But an already-approved earlier rejudge's
		// judgment IS Judged, and is exactly what the contestant is carrying going
		// into this one. Filtering those out (as this query used to) makes a second
		// rejudge measure itself against the original judge run and overstate the
		// change by a delta the admin already approved.
		priors, err = q.findJudgmentOutcomes(ctx, bson.M{
			"submissionID": bson.M{"$in": submissionIDs},
			"status":       judgedStatus,
		})
		if err != nil {
			return nil, err
		}
	}

	return &dto.Rejudge{
		ID:          rejudge.ID,
		Reason:      rejudge.Reason,
		Status:      rejudge.Status,
		Submissions: pairSubmissionComparisons(afters, priors),
		CreatedAt:   rejudge.CreatedAt,
	}, nil
}

// judgmentDetailRow is a judgment down to the grading-item and testcase detail a
// comparison needs. Label, name and each level's scores all come from the
// judgment's own snapshot rather than from the project tables, so a comparison
// describes the configuration each side was actually scored against — the point
// of a rejudge is usually that the configuration changed between the two.
type judgmentDetailRow struct {
	ID           string    `bson:"_id"`
	SubmissionID string    `bson:"submissionID"`
	Status       string    `bson:"status"`
	Score        float64   `bson:"score"`
	TotalScore   float64   `bson:"totalScore"`
	CreatedAt    time.Time `bson:"createdAt"`
	GradingItems []struct {
		ID              int32   `bson:"id"`
		Label           string  `bson:"label"`
		Name            string  `bson:"name"`
		Status          string  `bson:"status"`
		Score           float64 `bson:"score"`
		TotalScore      float64 `bson:"totalScore"`
		ScoringStrategy string  `bson:"scoringStrategy"`
		Testcases       []struct {
			ID          int32   `bson:"id"`
			Name        string  `bson:"name"`
			Status      string  `bson:"status"`
			Score       float64 `bson:"score"`
			TotalScore  float64 `bson:"totalScore"`
			ParsedScore float64 `bson:"parsedScore"`
		} `bson:"testcases"`
	} `bson:"gradingItems"`
}

func (r judgmentDetailRow) detail() *dto.JudgmentOutcomeDetail {
	gradingItems := make([]*dto.GradingItemOutcome, 0, len(r.GradingItems))
	for _, gi := range r.GradingItems {
		testcases := make([]*dto.TestcaseOutcome, 0, len(gi.Testcases))
		for _, tc := range gi.Testcases {
			testcases = append(testcases, &dto.TestcaseOutcome{
				ID:          tc.ID,
				Name:        tc.Name,
				Status:      tc.Status,
				Score:       tc.Score,
				TotalScore:  tc.TotalScore,
				ParsedScore: tc.ParsedScore,
			})
		}

		gradingItems = append(gradingItems, &dto.GradingItemOutcome{
			ID:              gi.ID,
			Label:           gi.Label,
			Name:            gi.Name,
			Status:          gi.Status,
			Score:           gi.Score,
			TotalScore:      gi.TotalScore,
			ScoringStrategy: gi.ScoringStrategy,
			Testcases:       testcases,
		})
	}

	return &dto.JudgmentOutcomeDetail{
		JudgmentID:   r.ID,
		Status:       r.Status,
		Score:        r.Score,
		TotalScore:   r.TotalScore,
		GradingItems: gradingItems,
	}
}

func (q *RejudgeQuery) findJudgmentDetails(ctx context.Context, filter bson.M) ([]judgmentDetailRow, error) {
	collection := q.mdb.Database("judge").Collection("judgments")
	projection := bson.M{
		"_id":          1,
		"submissionID": 1,
		"status":       1,
		"score":        1,
		"totalScore":   1,
		"createdAt":    1,
		"gradingItems": 1,
	}

	cursor, err := collection.Find(ctx, filter, options.Find().SetProjection(projection))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	rows := []judgmentDetailRow{}
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, err
	}
	return rows, nil
}

// GetSubmissionComparison returns one submission's before/after at grading-item
// and testcase resolution.
//
// Fetched per submission on request rather than folded into GetRejudgeByID: a
// contest-project rejudge covers every submission in the project, and attaching
// every grading item and testcase of both sides to each one would make that list
// response grow without bound.
//
// The baseline is chosen on the same terms as the list view — see
// pairSubmissionComparisons for why it is anchored to the rejudge's own judgment
// rather than to whatever is adopted now.
func (q *RejudgeQuery) GetSubmissionComparison(rejudgeID string, submissionID string) (*dto.SubmissionComparisonDetail, error) {
	ctx := context.Background()

	var rejudge Rejudge
	err := q.mdb.Database("judge").Collection("rejudges").
		FindOne(ctx, bson.M{"_id": rejudgeID}).Decode(&rejudge)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, httperror.New(http.StatusNotFound, "rejudge not found")
	}
	if err != nil {
		return nil, err
	}

	afters, err := q.findJudgmentDetails(ctx, bson.M{
		"rejudgeID":    rejudgeID,
		"submissionID": submissionID,
	})
	if err != nil {
		return nil, err
	}
	if len(afters) == 0 {
		return nil, httperror.New(http.StatusNotFound, "submission not found")
	}

	// One rejudge produces one judgment per submission. If somehow more exist,
	// report the newest rather than an arbitrary one.
	after := afters[0]
	for _, candidate := range afters[1:] {
		if candidate.CreatedAt.After(after.CreatedAt) {
			after = candidate
		}
	}

	priors, err := q.findJudgmentDetails(ctx, bson.M{
		"submissionID": submissionID,
		"status":       judgedStatus,
	})
	if err != nil {
		return nil, err
	}

	var baseline *judgmentDetailRow
	for i := range priors {
		if priors[i].ID == after.ID || !priors[i].CreatedAt.Before(after.CreatedAt) {
			continue
		}
		if baseline == nil || priors[i].CreatedAt.After(baseline.CreatedAt) {
			baseline = &priors[i]
		}
	}

	comparison := &dto.SubmissionComparisonDetail{
		ID:    submissionID,
		After: after.detail(),
	}
	if baseline != nil {
		comparison.Before = baseline.detail()
	}

	return comparison, nil
}
