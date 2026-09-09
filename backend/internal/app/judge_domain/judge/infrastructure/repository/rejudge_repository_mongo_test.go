package repository

import (
	"context"
	"testing"
	"time"

	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgestatus"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/rejudge"
	sharedlogger "judge-backend/internal/app/shared/infrastructure/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// These tests run against the same standalone mongo:4.4.18 the deployment
// uses. CommitJudgmentsByRejudgeID once used a multi-document transaction,
// which a standalone instance rejects with IllegalOperation — every approve
// silently failed in production while the mocked-repository tests stayed
// green. Keeping these on a real standalone Mongo pins the topology
// assumption itself.

func seedRejudgeWithJudgments(t *testing.T, client *mongo.Client, rejudgeID, rejudgeStatus string, judgmentIDs []string, judgmentStatus string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := client.Database("judge").Collection("rejudges").InsertOne(ctx, bson.M{
		"_id":       rejudgeID,
		"judgments": judgmentIDs,
		"reason":    "test",
		"status":    rejudgeStatus,
		"createdAt": time.Now(),
	})
	require.NoError(t, err)

	for _, id := range judgmentIDs {
		_, err := client.Database("judge").Collection("judgments").InsertOne(ctx, bson.M{
			"_id":       id,
			"rejudgeID": rejudgeID,
			"judgeType": "REJUDGE",
			"status":    judgmentStatus,
		})
		require.NoError(t, err)
	}
}

func judgmentStatusOf(t *testing.T, client *mongo.Client, judgmentID string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var doc struct {
		Status string `bson:"status"`
	}
	require.NoError(t, client.Database("judge").Collection("judgments").
		FindOne(ctx, bson.M{"_id": judgmentID}).Decode(&doc))
	return doc.Status
}

func rejudgeStatusOf(t *testing.T, client *mongo.Client, rejudgeID string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var doc struct {
		Status string `bson:"status"`
	}
	require.NoError(t, client.Database("judge").Collection("rejudges").
		FindOne(ctx, bson.M{"_id": rejudgeID}).Decode(&doc))
	return doc.Status
}

// The document schema had no contest project field at all, so the admin list
// could not group rejudges by the project they belong to. Save and FindByID are
// the only two places the field crosses the Mongo boundary; a round trip pins
// both at once, and a Save that forgets it would leave the stored document
// carrying a project of 0.
func TestSaveAndFindByID_RoundTripsTheContestProject(t *testing.T) {
	client := mongoClientForTest(t)
	repo := &RejudgeRepository{mdb: client, logger: sharedlogger.NewMockLogger()}

	created := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	require.NoError(t, repo.Save(rejudge.New(&rejudge.RejudgeConfig{
		ID:               "rj-project",
		ContestProjectID: 42,
		Reason:           "grading item fixed",
		Status:           rejudge.PENDING,
		Judgments:        []string{"jm-project-1"},
		CreatedAt:        created,
	})))

	found, err := repo.FindByID("rj-project")
	require.NoError(t, err)
	assert.EqualValues(t, 42, found.ContestProjectID())
	assert.Equal(t, "grading item fixed", found.Reason())
	assert.Equal(t, []string{"jm-project-1"}, found.Judgments())

	// The list endpoint reads the raw document rather than going through the
	// entity, so the stored field name is part of the contract.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var doc struct {
		ContestProjectID int32 `bson:"contestProjectID"`
	}
	require.NoError(t, client.Database("judge").Collection("rejudges").
		FindOne(ctx, bson.M{"_id": "rj-project"}).Decode(&doc))
	assert.EqualValues(t, 42, doc.ContestProjectID)
}

func TestCommitJudgmentsByRejudgeID_Approval_OnStandaloneMongo(t *testing.T) {
	client := mongoClientForTest(t)
	repo := &RejudgeRepository{mdb: client, logger: sharedlogger.NewMockLogger()}

	seedRejudgeWithJudgments(t, client, "rj-approve", rejudge.FINISHED.String(),
		[]string{"jm-approve-1", "jm-approve-2"}, judgestatus.REJUDGED.String())

	applied, err := repo.CommitJudgmentsByRejudgeID("rj-approve", true)
	require.NoError(t, err)
	assert.True(t, applied)

	assert.Equal(t, rejudge.APPROVED.String(), rejudgeStatusOf(t, client, "rj-approve"))
	assert.Equal(t, judgestatus.JUDGED.String(), judgmentStatusOf(t, client, "jm-approve-1"))
	assert.Equal(t, judgestatus.JUDGED.String(), judgmentStatusOf(t, client, "jm-approve-2"))
}

func TestCommitJudgmentsByRejudgeID_Rejection_LeavesJudgmentsUntouched(t *testing.T) {
	client := mongoClientForTest(t)
	repo := &RejudgeRepository{mdb: client, logger: sharedlogger.NewMockLogger()}

	seedRejudgeWithJudgments(t, client, "rj-reject", rejudge.FINISHED.String(),
		[]string{"jm-reject-1"}, judgestatus.REJUDGED.String())

	applied, err := repo.CommitJudgmentsByRejudgeID("rj-reject", false)
	require.NoError(t, err)
	assert.True(t, applied)

	assert.Equal(t, rejudge.REJECTED.String(), rejudgeStatusOf(t, client, "rj-reject"))
	assert.Equal(t, judgestatus.REJUDGED.String(), judgmentStatusOf(t, client, "jm-reject-1"))
}

func TestCommitJudgmentsByRejudgeID_NotFinished_IsNoOp(t *testing.T) {
	client := mongoClientForTest(t)
	repo := &RejudgeRepository{mdb: client, logger: sharedlogger.NewMockLogger()}

	// Already decided: a second admin's approve must lose the race.
	seedRejudgeWithJudgments(t, client, "rj-decided", rejudge.REJECTED.String(),
		[]string{"jm-decided-1"}, judgestatus.REJUDGED.String())

	applied, err := repo.CommitJudgmentsByRejudgeID("rj-decided", true)
	require.NoError(t, err)
	assert.False(t, applied)

	assert.Equal(t, rejudge.REJECTED.String(), rejudgeStatusOf(t, client, "rj-decided"))
	assert.Equal(t, judgestatus.REJUDGED.String(), judgmentStatusOf(t, client, "jm-decided-1"))
}

func TestRepairUncommittedApprovals_ReplaysLostJudgmentsFlip(t *testing.T) {
	client := mongoClientForTest(t)
	repo := &RejudgeRepository{mdb: client, logger: sharedlogger.NewMockLogger()}

	// Crash window artifact: rejudge already APPROVED, judgments never flipped.
	seedRejudgeWithJudgments(t, client, "rj-crashed", rejudge.APPROVED.String(),
		[]string{"jm-crashed-1", "jm-crashed-2"}, judgestatus.REJUDGED.String())
	// Rejected rejudge: its Rejudged judgments are terminal and must NOT be repaired.
	seedRejudgeWithJudgments(t, client, "rj-rejected-terminal", rejudge.REJECTED.String(),
		[]string{"jm-terminal-1"}, judgestatus.REJUDGED.String())

	repaired, err := repo.RepairUncommittedApprovals()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, repaired, int64(2))

	assert.Equal(t, judgestatus.JUDGED.String(), judgmentStatusOf(t, client, "jm-crashed-1"))
	assert.Equal(t, judgestatus.JUDGED.String(), judgmentStatusOf(t, client, "jm-crashed-2"))
	assert.Equal(t, judgestatus.REJUDGED.String(), judgmentStatusOf(t, client, "jm-terminal-1"))
}
