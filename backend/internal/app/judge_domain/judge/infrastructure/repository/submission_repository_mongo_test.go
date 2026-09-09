package repository

import (
	"context"
	"testing"
	"time"

	sharedlogger "judge-backend/internal/app/shared/infrastructure/logger"

	"go.mongodb.org/mongo-driver/bson"
)

// FindWithoutJudgment feeds the orphan repair sweep, and its grace period is
// what keeps the sweep off submissions whose (asynchronous) judgment creation
// is simply still in flight: the sweep's read and the handler's write are not
// atomic, and without the cutoff the sweep raced a terminal ERROR judgment
// against the real one. The shared Mongo container carries documents from the
// other repository tests, so every assertion here is membership-based —
// counts would be meaningless.

func insertRawSubmission(t *testing.T, e *SubmissionRepository, doc bson.M) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := e.mdb.Database("judge").Collection("submissions").InsertOne(ctx, doc); err != nil {
		t.Fatalf("insert submission doc: %v", err)
	}
}

func insertRawJudgment(t *testing.T, e *SubmissionRepository, id, submissionID string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := e.mdb.Database("judge").Collection("judgments").InsertOne(ctx, bson.M{
		"_id": id, "submissionID": submissionID, "status": "Queued", "judgeType": "JUDGE", "createdAt": time.Now(),
	}); err != nil {
		t.Fatalf("insert judgment doc: %v", err)
	}
}

func TestFindWithoutJudgment_GracePeriod(t *testing.T) {
	lg := sharedlogger.NewMockLogger()
	repo := NewSubmissionRepository(mongoClientForTest(t), lg).(*SubmissionRepository)

	now := time.Now()
	insertRawSubmission(t, repo, bson.M{"_id": "zz-grace-old", "userID": int32(1), "contestProjectID": int32(10),
		"gradingItemIds": bson.A{}, "ignored": false, "createdAt": now.Add(-time.Hour)})
	insertRawSubmission(t, repo, bson.M{"_id": "zz-grace-fresh", "userID": int32(1), "contestProjectID": int32(10),
		"gradingItemIds": bson.A{}, "ignored": false, "createdAt": now})
	insertRawSubmission(t, repo, bson.M{"_id": "zz-grace-nocreated", "userID": int32(1), "contestProjectID": int32(10),
		"gradingItemIds": bson.A{}, "ignored": false})
	insertRawSubmission(t, repo, bson.M{"_id": "zz-grace-judged", "userID": int32(1), "contestProjectID": int32(10),
		"gradingItemIds": bson.A{}, "ignored": false, "createdAt": now.Add(-time.Hour)})
	insertRawJudgment(t, repo, "zz-grace-judged-jdg", "zz-grace-judged")
	insertRawSubmission(t, repo, bson.M{"_id": "zz-grace-ignored", "userID": int32(1), "contestProjectID": int32(10),
		"gradingItemIds": bson.A{}, "ignored": true, "createdAt": now.Add(-time.Hour)})
	// One malformed document must not brick the whole sweep — the query feeds
	// a periodic repair pass, so a batch-level decode failure would abort
	// every future repair deterministically.
	insertRawSubmission(t, repo, bson.M{"_id": "zz-grace-malformed", "userID": int32(1), "contestProjectID": int32(10),
		"gradingItemIds": bson.A{"not-an-int"}, "ignored": false, "createdAt": now.Add(-time.Hour)})

	rows, err := repo.FindWithoutJudgment(now.Add(-10 * time.Minute))
	if err != nil {
		t.Fatalf("FindWithoutJudgment: %v", err)
	}
	got := map[string]bool{}
	for _, r := range rows {
		got[r.ID()] = true
	}

	if !got["zz-grace-old"] {
		t.Error("expected the hour-old orphan to be listed")
	}
	if got["zz-grace-fresh"] {
		t.Error("a submission younger than the cutoff must NOT be listed — its judgment creation may simply still be in flight")
	}
	if !got["zz-grace-nocreated"] {
		t.Error("a document without createdAt counts as old enough — it must never be permanently excluded from repair")
	}
	if got["zz-grace-judged"] {
		t.Error("a submission with a judgment must not be listed")
	}
	if got["zz-grace-ignored"] {
		t.Error("an ignored submission must not be listed")
	}
	if got["zz-grace-malformed"] {
		t.Error("a malformed document is skipped, not repaired")
	}
}
