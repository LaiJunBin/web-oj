package judge_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"judge-backend/internal/config"
	"judge-backend/internal/projectpath"
	"judge-backend/tests/testutils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	mongoDBName            = "judge"
	submissionsCollection  = "submissions"
	judgmentsCollection    = "judgments"
	rejudgesCollection     = "rejudges"
	defaultPollingTimeout  = 30 * time.Second
	defaultPollingInterval = 100 * time.Millisecond
)

type judgeFeature struct {
	mdb *mongo.Client
	cfg *config.Config
}

func newJudgeFeature(mdb *mongo.Client, cfg *config.Config) *judgeFeature {
	return &judgeFeature{mdb: mdb, cfg: cfg}
}

func (f *judgeFeature) collection(name string) *mongo.Collection {
	return f.mdb.Database(mongoDBName).Collection(name)
}

// --- Verification steps: query MongoDB and assert ---

func (f *judgeFeature) ASubmissionShouldExistForContestProject(contestProjectID int) error {
	return f.pollUntil(func() error {
		ctx := context.Background()
		col := f.collection(submissionsCollection)
		filter := bson.M{"contestProjectID": int32(contestProjectID)}
		count, err := col.CountDocuments(ctx, filter)
		if err != nil {
			return err
		}
		if count == 0 {
			return fmt.Errorf("expected a submission for contest project %d, found none", contestProjectID)
		}
		return nil
	})
}

func (f *judgeFeature) AJudgmentShouldExistWithStatus(expectedStatus string) error {
	return f.pollUntil(func() error {
		ctx := context.Background()
		col := f.collection(judgmentsCollection)
		filter := bson.M{"status": expectedStatus}
		count, err := col.CountDocuments(ctx, filter)
		if err != nil {
			return err
		}
		if count == 0 {
			return fmt.Errorf("expected a judgment with status %q, found none", expectedStatus)
		}
		return nil
	})
}

func (f *judgeFeature) AJudgmentShouldExistForContestProject(contestProjectID int) error {
	return f.pollUntil(func() error {
		ctx := context.Background()
		col := f.collection(judgmentsCollection)
		filter := bson.M{"contestProjectID": int32(contestProjectID)}
		count, err := col.CountDocuments(ctx, filter)
		if err != nil {
			return err
		}
		if count == 0 {
			return fmt.Errorf("expected a judgment for contest project %d, found none", contestProjectID)
		}
		return nil
	})
}

func (f *judgeFeature) AJudgmentShouldExistForSubmissionID(submissionID string) error {
	return f.pollUntil(func() error {
		ctx := context.Background()
		col := f.collection(judgmentsCollection)
		filter := bson.M{"submissionID": submissionID}
		count, err := col.CountDocuments(ctx, filter)
		if err != nil {
			return err
		}
		if count == 0 {
			return fmt.Errorf("expected a judgment for submission %q, found none", submissionID)
		}
		return nil
	})
}

func (f *judgeFeature) TheJudgmentShouldHaveStatus(judgmentID, expectedStatus string) error {
	return f.pollUntil(func() error {
		ctx := context.Background()
		col := f.collection(judgmentsCollection)
		var result bson.M
		err := col.FindOne(ctx, bson.M{"_id": judgmentID}).Decode(&result)
		if err != nil {
			return fmt.Errorf("judgment %q not found: %w", judgmentID, err)
		}
		actual, _ := result["status"].(string)
		if actual != expectedStatus {
			return fmt.Errorf("expected judgment %q status %q, got %q", judgmentID, expectedStatus, actual)
		}
		return nil
	})
}

func (f *judgeFeature) TheJudgmentShouldHaveReason(judgmentID, expectedReason string) error {
	return f.pollUntil(func() error {
		ctx := context.Background()
		col := f.collection(judgmentsCollection)
		var result bson.M
		err := col.FindOne(ctx, bson.M{"_id": judgmentID}).Decode(&result)
		if err != nil {
			return fmt.Errorf("judgment %q not found: %w", judgmentID, err)
		}
		actual, _ := result["reason"].(string)
		if actual != expectedReason {
			return fmt.Errorf("expected judgment %q reason %q, got %q", judgmentID, expectedReason, actual)
		}
		return nil
	})
}

func (f *judgeFeature) TheJudgmentShouldHaveStaffNote(judgmentID, expectedNote string) error {
	return f.pollUntil(func() error {
		ctx := context.Background()
		col := f.collection(judgmentsCollection)
		var result bson.M
		err := col.FindOne(ctx, bson.M{"_id": judgmentID}).Decode(&result)
		if err != nil {
			return fmt.Errorf("judgment %q not found: %w", judgmentID, err)
		}
		actual, ok := result["staffNote"].(string)
		if !ok {
			return fmt.Errorf("expected judgment %q staff note %q, got none", judgmentID, expectedNote)
		}
		if actual != expectedNote {
			return fmt.Errorf("expected judgment %q staff note %q, got %q", judgmentID, expectedNote, actual)
		}
		return nil
	})
}

// TheJudgmentShouldHaveNoStaffNote polls like its positive twin — but
// note what that means for a "the write was rejected" assertion: a never-
// written note is absent from the very first poll, so the step passes at
// once and does NOT wait to see whether a stray write lands later. It is a
// meaningful check only after a clear, or as a sanity pin, never as proof
// that a rejected request had no side effect.
func (f *judgeFeature) TheJudgmentShouldHaveNoStaffNote(judgmentID string) error {
	return f.pollUntil(func() error {
		ctx := context.Background()
		col := f.collection(judgmentsCollection)
		var result bson.M
		err := col.FindOne(ctx, bson.M{"_id": judgmentID}).Decode(&result)
		if err != nil {
			return fmt.Errorf("judgment %q not found: %w", judgmentID, err)
		}
		if note, ok := result["staffNote"].(string); ok {
			return fmt.Errorf("expected judgment %q to have no staff note, got %q", judgmentID, note)
		}
		return nil
	})
}

func (f *judgeFeature) ThereShouldBeNJudgmentsForSubmission(count int, submissionID string) error {
	return f.pollUntil(func() error {
		ctx := context.Background()
		col := f.collection(judgmentsCollection)
		filter := bson.M{"submissionID": submissionID}
		actual, err := col.CountDocuments(ctx, filter)
		if err != nil {
			return err
		}
		if int(actual) != count {
			return fmt.Errorf("expected %d judgments for submission %q, found %d", count, submissionID, actual)
		}
		return nil
	})
}

// pollJudgmentReachesStatus polls until the judgment matching filter reaches
// the expected terminal status. On mismatch it surfaces the judgment's
// errorMessage/reason so pipeline failures are diagnosable from the step
// output alone.
func (f *judgeFeature) pollJudgmentReachesStatus(filter bson.M, describe string, expectedStatus string) error {
	return f.pollUntil(func() error {
		ctx := context.Background()
		col := f.collection(judgmentsCollection)
		var result bson.M
		err := col.FindOne(ctx, filter).Decode(&result)
		if err != nil {
			return fmt.Errorf("no judgment for %s: %w", describe, err)
		}
		actual, _ := result["status"].(string)
		if actual != expectedStatus {
			errorMessage, _ := result["errorMessage"].(string)
			reason, _ := result["reason"].(string)
			return fmt.Errorf("expected judgment status %q for %s, got %q (errorMessage: %q, reason: %q)", expectedStatus, describe, actual, errorMessage, reason)
		}
		return nil
	})
}

// TheJudgmentForContestProjectShouldReachStatus polls until the (single)
// judgment of a contest project reaches the expected terminal status. With the
// fake runtime wired in feature tests, judging is deterministic, so asserting
// "Judged" here covers the whole submission → judged pipeline.
func (f *judgeFeature) TheJudgmentForContestProjectShouldReachStatus(contestProjectID int, expectedStatus string) error {
	return f.pollJudgmentReachesStatus(
		bson.M{"contestProjectID": int32(contestProjectID)},
		fmt.Sprintf("contest project %d", contestProjectID),
		expectedStatus,
	)
}

// TheJudgmentForContestProjectShouldHaveReason reads the contestant-facing
// reason off the judgment a submit-triggered run produced, whose id the
// scenario never learns.
func (f *judgeFeature) TheJudgmentForContestProjectShouldHaveReason(contestProjectID int, expectedReason string) error {
	return f.pollUntil(func() error {
		ctx := context.Background()
		col := f.collection(judgmentsCollection)
		var result bson.M
		err := col.FindOne(ctx, bson.M{"contestProjectID": int32(contestProjectID)}).Decode(&result)
		if err != nil {
			return fmt.Errorf("no judgment for contest project %d: %w", contestProjectID, err)
		}
		actual, _ := result["reason"].(string)
		if actual != expectedReason {
			return fmt.Errorf("expected judgment reason %q for contest project %d, got %q", expectedReason, contestProjectID, actual)
		}
		return nil
	})
}

// EveryTestcaseOfTheJudgmentForContestProjectShouldHaveStatus walks the
// judgment's persisted grading-item snapshot; it is how a scenario proves an
// aborted run left nothing looking "Pending" underneath an ERROR judgment.
func (f *judgeFeature) EveryTestcaseOfTheJudgmentForContestProjectShouldHaveStatus(contestProjectID int, expectedStatus string) error {
	return f.pollUntil(func() error {
		ctx := context.Background()
		col := f.collection(judgmentsCollection)
		var result bson.M
		if err := col.FindOne(ctx, bson.M{"contestProjectID": int32(contestProjectID)}).Decode(&result); err != nil {
			return fmt.Errorf("no judgment for contest project %d: %w", contestProjectID, err)
		}
		items, _ := result["gradingItems"].(bson.A)
		seen := 0
		for _, gi := range items {
			giDoc, _ := gi.(bson.M)
			tcs, _ := giDoc["testcases"].(bson.A)
			for _, tc := range tcs {
				tcDoc, _ := tc.(bson.M)
				seen++
				if got, _ := tcDoc["status"].(string); got != expectedStatus {
					return fmt.Errorf("testcase %v of contest project %d has status %q, want %q", tcDoc["id"], contestProjectID, got, expectedStatus)
				}
			}
		}
		if seen == 0 {
			return fmt.Errorf("judgment for contest project %d has no testcases to check", contestProjectID)
		}
		return nil
	})
}

func (f *judgeFeature) TheJudgmentScoreForContestProjectShouldBe(contestProjectID int, expectedScore int) error {
	return f.pollUntil(func() error {
		ctx := context.Background()
		col := f.collection(judgmentsCollection)
		var result bson.M
		err := col.FindOne(ctx, bson.M{"contestProjectID": int32(contestProjectID)}).Decode(&result)
		if err != nil {
			return fmt.Errorf("no judgment for contest project %d: %w", contestProjectID, err)
		}
		actual, _ := result["score"].(float64)
		if actual != float64(expectedScore) {
			return fmt.Errorf("expected judgment score %d for contest project %d, got %v", expectedScore, contestProjectID, actual)
		}
		return nil
	})
}

func (f *judgeFeature) ARejudgeShouldExistWithStatus(expectedStatus string) error {
	return f.pollUntil(func() error {
		ctx := context.Background()
		col := f.collection(rejudgesCollection)
		filter := bson.M{"status": expectedStatus}
		count, err := col.CountDocuments(ctx, filter)
		if err != nil {
			return err
		}
		if count == 0 {
			return fmt.Errorf("expected a rejudge with status %q, found none", expectedStatus)
		}
		return nil
	})
}

// TheJudgmentForSubmissionShouldReachStatus polls until the judgment of a
// seeded submission reaches the expected terminal status.
func (f *judgeFeature) TheJudgmentForSubmissionShouldReachStatus(submissionID, expectedStatus string) error {
	return f.pollJudgmentReachesStatus(
		bson.M{"submissionID": submissionID},
		fmt.Sprintf("submission %q", submissionID),
		expectedStatus,
	)
}

// --- Seeding steps ---

// SubmissionHasStoredJudgeFiles writes the on-disk artifacts a judgment run
// reads back (currently metadata.json). Seeded submissions only exist in the
// database, so scenarios that expect judging to succeed must seed the files
// too. The location is derived from the same config the server reads
// (server.storagepath in config_test.yaml, relative to the backend root) —
// the pattern api_feature.go already uses for image assets.
func (f *judgeFeature) SubmissionHasStoredJudgeFiles(submissionID string) error {
	dir := filepath.Join(projectpath.Root, f.cfg.Server.StoragePath, "submissions", submissionID)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create submission storage dir: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "metadata.json"), []byte("{}"), 0600); err != nil {
		return fmt.Errorf("failed to write metadata.json: %w", err)
	}
	return nil
}

// --- Polling helper for async event propagation ---

func (f *judgeFeature) pollUntil(check func() error) error {
	deadline := time.Now().Add(defaultPollingTimeout)
	var lastErr error
	for time.Now().Before(deadline) {
		lastErr = check()
		if lastErr == nil {
			return nil
		}
		time.Sleep(defaultPollingInterval)
	}
	return fmt.Errorf("timed out after %v: %w", defaultPollingTimeout, lastErr)
}

// --- Register all judge steps ---

func initJudgeSteps(ctx *testutils.Context, judge *judgeFeature) {
	// Verification
	ctx.Context.Step(`^a submission should exist for contest project (\d+)$`, judge.ASubmissionShouldExistForContestProject)
	ctx.Context.Step(`^a judgment should exist with status "([^"]*)"$`, judge.AJudgmentShouldExistWithStatus)
	ctx.Context.Step(`^a judgment should exist for contest project (\d+)$`, judge.AJudgmentShouldExistForContestProject)
	ctx.Context.Step(`^a judgment should exist for submission "([^"]*)"$`, judge.AJudgmentShouldExistForSubmissionID)
	ctx.Context.Step(`^the judgment "([^"]*)" should have status "([^"]*)"$`, judge.TheJudgmentShouldHaveStatus)
	ctx.Context.Step(`^the judgment for contest project (\d+) should reach status "([^"]*)"$`, judge.TheJudgmentForContestProjectShouldReachStatus)
	ctx.Context.Step(`^the judgment for submission "([^"]*)" should reach status "([^"]*)"$`, judge.TheJudgmentForSubmissionShouldReachStatus)
	ctx.Context.Step(`^the judgment score for contest project (\d+) should be (\d+)$`, judge.TheJudgmentScoreForContestProjectShouldBe)
	ctx.Context.Step(`^the judgment for contest project (\d+) should have reason "([^"]*)"$`, judge.TheJudgmentForContestProjectShouldHaveReason)
	ctx.Context.Step(`^every testcase of the judgment for contest project (\d+) should have status "([^"]*)"$`, judge.EveryTestcaseOfTheJudgmentForContestProjectShouldHaveStatus)
	// Seeding
	ctx.Context.Step(`^submission "([^"]*)" has stored judge files$`, judge.SubmissionHasStoredJudgeFiles)
	ctx.Context.Step(`^the judgment "([^"]*)" should have reason "([^"]*)"$`, judge.TheJudgmentShouldHaveReason)
	ctx.Context.Step(`^the judgment "([^"]*)" should have staff note "([^"]*)"$`, judge.TheJudgmentShouldHaveStaffNote)
	ctx.Context.Step(`^the judgment "([^"]*)" should have no staff note$`, judge.TheJudgmentShouldHaveNoStaffNote)
	ctx.Context.Step(`^there should be (\d+) judgments for submission "([^"]*)"$`, judge.ThereShouldBeNJudgmentsForSubmission)
	ctx.Context.Step(`^a rejudge should exist with status "([^"]*)"$`, judge.ARejudgeShouldExistWithStatus)
}

// --- Helpers ---

func toInt32(s string) int32 {
	var v int32
	fmt.Sscanf(s, "%d", &v)
	return v
}

func toFloat64(s string) float64 {
	var v float64
	fmt.Sscanf(s, "%f", &v)
	return v
}
