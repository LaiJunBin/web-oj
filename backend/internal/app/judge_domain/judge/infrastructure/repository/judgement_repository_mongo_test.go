package repository

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/gradingitem"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/gradingitem/scoringstrategy"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgestatus"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgetype"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/lifecycle"
	judgmentproject "judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/project"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/status"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/testcase"
	sharedlogger "judge-backend/internal/app/shared/infrastructure/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// This file holds the repository-level tests that need a real MongoDB,
// because what they assert is the behaviour of the conditional-update filters
// themselves — MatchedCount semantics under concurrent writers. Mocking the
// collection would only re-assert our own assumptions about Mongo.
//
// The container is started once for the whole package. When no Docker daemon
// is reachable these tests skip, so `go test ./internal/...` still runs on a
// developer machine without Docker. Skipping silently would make the
// guarantees here easy to lose, so CI sets REQUIRE_MONGO_TESTS=1, which turns
// an unusable Docker/MongoDB into a hard failure instead of a skip. Keep at
// least one CI job running with that variable set.
const requireMongoTestsEnv = "REQUIRE_MONGO_TESTS"

var (
	sharedMongoClient *mongo.Client
	mongoSetupOnce    sync.Once
	mongoSetupErr     error
	mongoTerminate    func()
)

func mongoClientForTest(t *testing.T) *mongo.Client {
	t.Helper()

	mongoSetupOnce.Do(func() {
		ctx := context.Background()
		container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image:        "mongo:4.4.18",
				ExposedPorts: []string{"27017/tcp"},
				Env: map[string]string{
					"MONGO_INITDB_ROOT_USERNAME": "root",
					"MONGO_INITDB_ROOT_PASSWORD": "root",
				},
				WaitingFor: wait.ForLog("Waiting for connections").WithStartupTimeout(2 * time.Minute),
			},
			Started: true,
		})
		if err != nil {
			mongoSetupErr = err
			return
		}
		mongoTerminate = func() { _ = container.Terminate(context.Background()) }

		host, err := container.Host(ctx)
		if err != nil {
			mongoSetupErr = err
			return
		}
		port, err := container.MappedPort(ctx, "27017")
		if err != nil {
			mongoSetupErr = err
			return
		}

		uri := fmt.Sprintf("mongodb://root:root@%s:%d/", host, port.Int())
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
		if err != nil {
			mongoSetupErr = err
			return
		}
		pingCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		if err := client.Ping(pingCtx, nil); err != nil {
			mongoSetupErr = err
			return
		}
		sharedMongoClient = client
	})

	if mongoSetupErr != nil {
		if os.Getenv(requireMongoTestsEnv) != "" {
			t.Fatalf("%s is set, so these tests must run, but no usable Docker/MongoDB was available: %v",
				requireMongoTestsEnv, mongoSetupErr)
		}
		t.Skipf("skipping MongoDB-backed repository test, no usable Docker/MongoDB (set %s to make this a failure): %v",
			requireMongoTestsEnv, mongoSetupErr)
	}
	return sharedMongoClient
}

func TestMain(m *testing.M) {
	code := m.Run()
	if mongoTerminate != nil {
		mongoTerminate()
	}
	os.Exit(code)
}

// newRepoForTest returns a repository wired to the test Mongo. The gorm
// handle is nil on purpose: none of the conditional-update methods under
// test touch the relational side.
func newRepoForTest(t *testing.T) *JudgmentRepository {
	t.Helper()
	return &JudgmentRepository{
		mdb:    mongoClientForTest(t),
		logger: sharedlogger.NewMockLogger(),
	}
}

func seedJudgment(t *testing.T, r *JudgmentRepository, id string, status *judgestatus.JudgeStatus) *judgment.Judgment {
	t.Helper()

	tc := testcase.NewTestcase(&testcase.TestcaseConfig{ID: 1, Command: "echo hello", Score: 100})
	tc.SetLifecycle(lifecycle.NewTestcaseLifecycle(nil))
	gi := gradingitem.NewGradingItem(&gradingitem.GradingItemConfig{
		ID:              1,
		Score:           100,
		ScoringStrategy: scoringstrategy.IncreaseScoreStrategy,
		Testcases:       []*testcase.Testcase{tc},
	})
	j := judgment.New(&judgment.JudgmentConfig{
		ContestProjectID: 1,
		SubmissionID:     "sub-1",
		Project:          judgmentproject.NewProject(&judgmentproject.ProjectConfig{ID: 1}),
		GradingItems:     []*gradingitem.GradingItem{gi},
		Lifecycle:        lifecycle.NewJudgmentLifecycle(nil),
		JudgeType:        judgetype.JUDGE,
		CreatedAt:        time.Now(),
	})
	j.SetID(id)
	j.SetStatus(status)
	require.NoError(t, r.Save(j))
	return j
}

func statusInDB(t *testing.T, r *JudgmentRepository, id string) string {
	t.Helper()
	var doc struct {
		Status string `bson:"status"`
	}
	err := r.mdb.Database("judge").Collection("judgments").
		FindOne(context.Background(), bson.M{"_id": id}).Decode(&doc)
	require.NoError(t, err)
	return doc.Status
}

func reasonInDB(t *testing.T, r *JudgmentRepository, id string) string {
	t.Helper()
	var doc struct {
		Reason *string `bson:"reason"`
	}
	err := r.mdb.Database("judge").Collection("judgments").
		FindOne(context.Background(), bson.M{"_id": id}).Decode(&doc)
	require.NoError(t, err)
	if doc.Reason == nil {
		return ""
	}
	return *doc.Reason
}

// TestTryStart_OnlyOneCallerAcquiresOwnership is the test that was missing
// when TryStart still inferred applied=true by re-reading the status: with
// that inference, the loser of the race re-read the document, saw Judging
// (written by the winner), and reported success — so both callers would go
// on to execute the same judgment. Exactly one caller may observe
// applied=true.
func TestTryStart_OnlyOneCallerAcquiresOwnership(t *testing.T) {
	r := newRepoForTest(t)

	const workers = 8
	for round := 0; round < 20; round++ {
		id := fmt.Sprintf("judgment-trystart-%d", round)
		seedJudgment(t, r, id, judgestatus.QUEUED)

		var wg sync.WaitGroup
		results := make([]bool, workers)
		errs := make([]error, workers)
		start := make(chan struct{})

		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				<-start
				results[idx], errs[idx] = r.TryStart(id, time.Now())
			}(i)
		}
		close(start)
		wg.Wait()

		acquired := 0
		for i := 0; i < workers; i++ {
			require.NoError(t, errs[i])
			if results[i] {
				acquired++
			}
		}
		assert.Equal(t, 1, acquired, "round %d: exactly one caller must acquire the judgment, got %d", round, acquired)
		assert.Equal(t, "Judging", statusInDB(t, r, id))
	}
}

// TestTryStart_DoesNotResurrectCancelledJudgment covers the queued-
// cancellation case at the persistence level: the document is already
// Cancelled, so no caller may claim it and its status must be left alone.
func TestTryStart_DoesNotResurrectCancelledJudgment(t *testing.T) {
	r := newRepoForTest(t)
	id := "judgment-trystart-cancelled"
	seedJudgment(t, r, id, judgestatus.CANCELLED)

	applied, err := r.TryStart(id, time.Now())

	assert.NoError(t, err)
	assert.False(t, applied)
	assert.Equal(t, "Cancelled", statusInDB(t, r, id))
}

// TestSaveTerminalIfJudging_NotAppliedWhenAnotherWriterReachedTheSameStatus
// is the regression test for the case where both writers converge on
// Cancelled. The status alone looks like success — the document says exactly
// what this worker wanted it to say — but this worker's own document write
// (its grading/testcase progress) was NOT stored, so reporting applied=true
// would let the caller publish an in-memory snapshot that never reached the
// DB.
func TestSaveTerminalIfJudging_NotAppliedWhenAnotherWriterReachedTheSameStatus(t *testing.T) {
	r := newRepoForTest(t)
	id := "judgment-terminal-converge"
	j := seedJudgment(t, r, id, judgestatus.JUDGING)

	// Another writer (the cancel handler) gets there first.
	cancelled, err := r.TryCancel(id, "cancelled by admin", time.Now())
	require.NoError(t, err)
	require.NotNil(t, cancelled)

	// This worker independently computes the very same terminal status, and
	// marks a testcase along the way that only exists in its own copy.
	j.GradingItems()[0].Testcases()[0].SetParsedScore(100)
	require.NoError(t, j.Cancel(""))

	applied, err := r.SaveTerminalIfJudging(j)

	assert.NoError(t, err)
	assert.False(t, applied, "must not claim success: this call stored nothing, even though the status matches what it wanted")
	assert.Equal(t, "cancelled by admin", reasonInDB(t, r, id), "the first writer's reason must survive")
}

// TestTryCancel_SecondCancelDoesNotClaimItsReasonWasStored guards TryCancel's
// applied semantics: the second of two competing cancellations writes
// nothing, so it must not report success — its reason is not what the
// document ends up holding.
func TestTryCancel_SecondCancelDoesNotClaimItsReasonWasStored(t *testing.T) {
	r := newRepoForTest(t)
	id := "judgment-double-cancel"
	seedJudgment(t, r, id, judgestatus.JUDGING)

	first, err := r.TryCancel(id, "first reason", time.Now())
	require.NoError(t, err)
	require.NotNil(t, first)

	second, err := r.TryCancel(id, "second reason", time.Now())

	assert.NoError(t, err)
	assert.Nil(t, second, "the second cancel wrote nothing and must not report a won transition")
	assert.Equal(t, "first reason", reasonInDB(t, r, id), "the stored reason must be the one that actually won")
}

// TestFindByID_PreservesContestProjectID pins FindByID's projection: it uses
// an explicit field list, and contestProjectID was missing from it, so every
// re-fetched judgment came back with ContestProjectID=0. That is not an
// exotic path — a user-requested cancellation of an executing judgment
// ALWAYS flows through "terminal save loses → FindByID → publish
// judgeupdated", and the contest-side stream filters those events by
// contest project ID, so the Cancelled update never reached the frontend.
func TestFindByID_PreservesContestProjectID(t *testing.T) {
	r := newRepoForTest(t)
	id := "judgment-findbyid-cpid"
	seeded := seedJudgment(t, r, id, judgestatus.JUDGING)
	require.EqualValues(t, 1, seeded.ContestProjectID(), "seed helper is expected to set a non-zero contest project ID")

	found, err := r.FindByID(id)

	require.NoError(t, err)
	assert.EqualValues(t, seeded.ContestProjectID(), found.ContestProjectID(),
		"FindByID must preserve contestProjectID; a zero value here silently breaks contest-stream filtering downstream")
}

// TestTryCancel_WinnerReceivesThePostUpdateEntity pins TryCancel's P1-1
// contract: the winner's entity comes back from the SAME atomic
// FindOneAndUpdate command, complete enough to publish judgeupdated from —
// no follow-up read. (With a separate re-read, a transient failure there
// left the cancellation persisted but unannounced forever: retrying
// TryCancel reports "already cancelled" and publishes nothing.)
func TestTryCancel_WinnerReceivesThePostUpdateEntity(t *testing.T) {
	r := newRepoForTest(t)
	id := "judgment-trycancel-entity"
	seedJudgment(t, r, id, judgestatus.JUDGING)

	cancelled, err := r.TryCancel(id, "cancelled by test", time.Now())

	require.NoError(t, err)
	require.NotNil(t, cancelled)
	assert.True(t, cancelled.Status().IsCancelled(), "returned entity must reflect the post-update state")
	require.NotNil(t, cancelled.Reason())
	assert.Equal(t, "cancelled by test", *cancelled.Reason())
	assert.NotNil(t, cancelled.JudgingFinishedAt())
	assert.EqualValues(t, 1, cancelled.ContestProjectID(), "the full document must come back, not just the fields the update touched")
	assert.Len(t, cancelled.GradingItems(), 1, "grading progress persisted by the worker must be present in the returned entity")
	assert.Equal(t, "Cancelled", statusInDB(t, r, id))
}

// TestSaveProgressIfJudging_RejectsTerminalEntity guards the method against
// its own payload: the Mongo filter only constrains the DOCUMENT's status,
// while the payload carries the ENTITY's. A terminal entity passed through
// the progress method would therefore match a still-Judging document and
// flip it terminal, bypassing SaveTerminalIfJudging entirely. The entity
// check must reject it before any write.
func TestSaveProgressIfJudging_RejectsTerminalEntity(t *testing.T) {
	r := newRepoForTest(t)
	id := "judgment-progress-rejects-terminal"
	j := seedJudgment(t, r, id, judgestatus.JUDGING)

	require.NoError(t, j.Cancel("terminal in memory"))

	applied, err := r.SaveProgressIfJudging(j)

	assert.Error(t, err, "a terminal entity must be rejected, not silently written as progress")
	assert.False(t, applied)
	assert.Equal(t, "Judging", statusInDB(t, r, id), "the document must be untouched")
}

// TestSaveTerminalIfJudging_RejectsJudgingEntity is the mirror image: a
// still-Judging entity would match the still-Judging document, write Judging
// back, and report a "terminal save" that stored nothing terminal.
func TestSaveTerminalIfJudging_RejectsJudgingEntity(t *testing.T) {
	r := newRepoForTest(t)
	id := "judgment-terminal-rejects-judging"
	j := seedJudgment(t, r, id, judgestatus.JUDGING)

	applied, err := r.SaveTerminalIfJudging(j)

	assert.Error(t, err, "a non-terminal entity must be rejected, not written back as a terminal save")
	assert.False(t, applied)
	assert.Equal(t, "Judging", statusInDB(t, r, id), "the document must be untouched")
}

// TestSaveProgressIfJudging_StopsApplyingOnceCancelled confirms the mid-run
// guard: once another writer moves the document out of Judging, progress
// saves stop landing instead of resurrecting it.
func TestSaveProgressIfJudging_StopsApplyingOnceCancelled(t *testing.T) {
	r := newRepoForTest(t)
	id := "judgment-progress-guard"
	j := seedJudgment(t, r, id, judgestatus.JUDGING)

	applied, err := r.SaveProgressIfJudging(j)
	require.NoError(t, err)
	assert.True(t, applied, "progress must apply while still Judging")

	cancelled, err := r.TryCancel(id, "stop it", time.Now())
	require.NoError(t, err)
	require.NotNil(t, cancelled)

	applied, err = r.SaveProgressIfJudging(j)

	assert.NoError(t, err)
	assert.False(t, applied)
	assert.Equal(t, "Cancelled", statusInDB(t, r, id), "a declined progress save must not revert the cancellation")
}

// TestSaveThenFind_RoundTripsTheGradingItemSnapshot guards the snapshot round
// trip through a real Mongo, because a mismatch between what Save writes and
// what FindByID reads back is silent and corrupting rather than loud.
//
// The specific trap: at the testcase level the persisted "score" is the score
// the testcase EARNED, while its configured max lives in "totalScore" — the
// reverse of what this document used to store. Rehydrating "score" into the
// entity's configuredScore would hand every later CalculateScore a max score
// that is really an earned score, quietly deflating the judgment on the next
// recalculation. Nothing else in the suite would catch that, so it is pinned
// here field by field.
func TestSaveThenFind_RoundTripsTheGradingItemSnapshot(t *testing.T) {
	r := newRepoForTest(t)
	id := "judgment-snapshot-roundtrip"

	// increase strategy, one passing and one failing testcase:
	//   passing → earns its parsed score (8), configured max 10
	//   failing → earns nothing, configured max 20
	// so the item earns 8 out of a configured 30.
	passing := testcase.NewTestcase(&testcase.TestcaseConfig{
		ID: 1, Name: "adds two numbers", Command: "run 1", Score: 10, ParsedScore: 8,
	})
	passing.SetLifecycle(lifecycle.NewTestcaseLifecycle(nil))
	passing.SetStatus(status.PASSED)

	failing := testcase.NewTestcase(&testcase.TestcaseConfig{
		ID: 2, Name: "handles overflow", Command: "run 2", Score: 20, ParsedScore: 5,
	})
	failing.SetLifecycle(lifecycle.NewTestcaseLifecycle(nil))
	failing.SetStatus(status.FAILED)

	gi := gradingitem.NewGradingItem(&gradingitem.GradingItemConfig{
		ID:              7,
		Label:           "B",
		Name:            "Arithmetic",
		Score:           30,
		ScoringStrategy: scoringstrategy.IncreaseScoreStrategy,
		Testcases:       []*testcase.Testcase{passing, failing},
	})
	j := judgment.New(&judgment.JudgmentConfig{
		ContestProjectID: 1,
		SubmissionID:     "sub-snapshot",
		Project:          judgmentproject.NewProject(&judgmentproject.ProjectConfig{ID: 1}),
		GradingItems:     []*gradingitem.GradingItem{gi},
		Lifecycle:        lifecycle.NewJudgmentLifecycle(nil),
		JudgeType:        judgetype.JUDGE,
		CreatedAt:        time.Now(),
	})
	j.SetID(id)
	j.SetStatus(judgestatus.JUDGED)
	require.NoError(t, r.Save(j))

	// What actually landed in Mongo, before any rehydration can paper over it.
	var doc struct {
		GradingItems []struct {
			ID         int32   `bson:"id"`
			Label      string  `bson:"label"`
			Name       string  `bson:"name"`
			Score      float64 `bson:"score"`
			TotalScore float64 `bson:"totalScore"`
			Testcases  []struct {
				ID          int32   `bson:"id"`
				Name        string  `bson:"name"`
				Score       float64 `bson:"score"`
				TotalScore  float64 `bson:"totalScore"`
				ParsedScore float64 `bson:"parsedScore"`
			} `bson:"testcases"`
		} `bson:"gradingItems"`
	}
	require.NoError(t, r.mdb.Database("judge").Collection("judgments").
		FindOne(context.Background(), bson.M{"_id": id}).Decode(&doc))

	require.Len(t, doc.GradingItems, 1)
	storedItem := doc.GradingItems[0]
	assert.Equal(t, "B", storedItem.Label, "the grading item's label must be snapshotted")
	assert.Equal(t, "Arithmetic", storedItem.Name, "the grading item's name must be snapshotted")
	assert.Equal(t, float64(8), storedItem.Score, "the item's score must be what it earned")
	assert.Equal(t, float64(30), storedItem.TotalScore, "the item's totalScore must be its configured max")

	require.Len(t, storedItem.Testcases, 2)
	storedPassing, storedFailing := storedItem.Testcases[0], storedItem.Testcases[1]

	assert.Equal(t, "adds two numbers", storedPassing.Name, "the testcase's name must be snapshotted")
	assert.Equal(t, float64(8), storedPassing.Score, "a passing testcase's score must be what it earned")
	assert.Equal(t, float64(10), storedPassing.TotalScore, "a testcase's totalScore must be its configured max")
	assert.Equal(t, float64(8), storedPassing.ParsedScore)

	assert.Equal(t, float64(0), storedFailing.Score, "a failing testcase earns nothing under the increase strategy")
	assert.Equal(t, float64(20), storedFailing.TotalScore)
	assert.Equal(t, float64(5), storedFailing.ParsedScore, "the raw parser output must survive alongside the earned score")

	// Rehydrate and confirm the entity got its CONFIGURED scores back, not the
	// earned ones — the corruption this test exists for.
	found, err := r.FindByID(id)
	require.NoError(t, err)
	require.Len(t, found.GradingItems(), 1)

	item := found.GradingItems()[0]
	assert.Equal(t, "B", item.Label())
	assert.Equal(t, "Arithmetic", item.Name())
	assert.Equal(t, float64(30), item.Score(), "the item must rehydrate with its configured max, not its earned score")

	require.Len(t, item.Testcases(), 2)
	rehydratedPassing, rehydratedFailing := item.Testcases()[0], item.Testcases()[1]

	assert.Equal(t, "adds two numbers", rehydratedPassing.Name())
	assert.Equal(t, float64(10), rehydratedPassing.Score(),
		"a testcase must rehydrate with its configured max (10), not the score it earned (8)")
	assert.Equal(t, float64(8), rehydratedPassing.ParsedScore())
	assert.Equal(t, float64(20), rehydratedFailing.Score(),
		"a testcase must rehydrate with its configured max (20), not the score it earned (0)")
	assert.Equal(t, float64(5), rehydratedFailing.ParsedScore())

	// The real payoff: recalculating from the rehydrated entity has to reproduce
	// the score, which it only can if the configured maxes survived the trip.
	assert.Equal(t, float64(8), item.CalculateScore(),
		"recalculating from the rehydrated entity must reproduce the earned score")

	// And saving the rehydrated entity must not drift the document.
	require.NoError(t, r.Save(found))
	var reSaved struct {
		GradingItems []struct {
			Score      float64 `bson:"score"`
			TotalScore float64 `bson:"totalScore"`
			Testcases  []struct {
				Score      float64 `bson:"score"`
				TotalScore float64 `bson:"totalScore"`
			} `bson:"testcases"`
		} `bson:"gradingItems"`
	}
	require.NoError(t, r.mdb.Database("judge").Collection("judgments").
		FindOne(context.Background(), bson.M{"_id": id}).Decode(&reSaved))
	assert.Equal(t, float64(8), reSaved.GradingItems[0].Score, "a save/find/save cycle must not drift the item score")
	assert.Equal(t, float64(30), reSaved.GradingItems[0].TotalScore)
	assert.Equal(t, float64(8), reSaved.GradingItems[0].Testcases[0].Score, "a save/find/save cycle must not drift testcase scores")
	assert.Equal(t, float64(10), reSaved.GradingItems[0].Testcases[0].TotalScore)
}
