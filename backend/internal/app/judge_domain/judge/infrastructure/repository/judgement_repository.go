package repository

import (
	"context"
	"fmt"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/execresult"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/gradingitem"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/gradingitem/scoringstrategy"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgestatus"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgetype"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/lifecycle"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/project"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/status"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/testcase"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/testcase/resultparser"
	"judge-backend/internal/app/judge_domain/judge/domain/repository"
	sharedtypes "judge-backend/internal/app/shared/types"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
	"judge-backend/internal/app/shared/infrastructure/mongodb"
)

type JudgmentRepository struct {
	db     *gorm.DB
	mdb    *mongo.Client
	logger sharedtypes.Logger
}

type Result struct {
	ExitCode int32  `bson:"exitCode"`
	ExecTime int64  `bson:"execTime"`
	Stdout   string `bson:"stdout"`
	Stderr   string `bson:"stderr"`
}
type Lifecycle struct {
	ID                   int32  `bson:"id"`
	ExecutionEnvironment string `bson:"executionEnvironment"`
	Command              string `bson:"command"`
	Result               *Result
}

type judgmentTestcase struct {
	ID   int32  `bson:"id"`
	Name string `bson:"name"`

	// Score is the score this testcase EARNED and TotalScore its configured max.
	// Only TotalScore rehydrates onto the entity's configuredScore — feeding the
	// earned score back in as the configured one would corrupt every later
	// recalculation of this judgment.
	Score       float64 `bson:"score"`
	TotalScore  float64 `bson:"totalScore"`
	ParsedScore float64 `bson:"parsedScore"`
	Command     string  `bson:"command"`
	Status      string  `bson:"status"`
	Result      *Result `bson:"result"`
	Lifecycle   struct {
		BeforeEach []Lifecycle
		AfterEach  []Lifecycle
	} `bson:"lifecycle"`
}

type judgmentGradingItem struct {
	ID              int32              `bson:"id"`
	Label           string             `bson:"label"`
	Name            string             `bson:"name"`
	Score           float64            `bson:"score"`
	TotalScore      float64            `bson:"totalScore"`
	ScoringStrategy string             `bson:"scoringStrategy"`
	ResultParser    string             `bson:"resultParser"`
	Status          string             `bson:"status"`
	Testcases       []judgmentTestcase `bson:"testcases"`
}

type Judgment struct {
	ID               string  `bson:"_id"`
	ContestProjectID int32   `bson:"contestProjectID"`
	SubmissionID     string  `bson:"submissionID"`
	RejudgeID        *string `bson:"rejudgeID"`
	Project          struct {
		ID           int32    `bson:"id"`
		ImageTag     string   `bson:"imageTag"`
		TestImageTag string   `bson:"testImageTag"`
		ImageEnv     []string `bson:"env"`
		TestImageEnv []string `bson:"testEnv"`
		MountPath    string   `bson:"mountPath"`
		ResultParser string   `bson:"resultParser"`
	} `bson:"project"`
	GradingItems []judgmentGradingItem `bson:"gradingItems"`
	Lifecycle    struct {
		BeforeAll []Lifecycle
		AfterAll  []Lifecycle
	} `bson:"lifecycle"`
	Status            string     `bson:"status"`
	JudgeType         string     `bson:"judgeType"`
	Score             float64    `bson:"totalScore"`
	ErrorMessage      *string    `bson:"errorMessage"`
	Reason            *string    `bson:"reason"`
	StaffNote         *string    `bson:"staffNote"`
	CreatedAt         time.Time  `bson:"createdAt"`
	JudgingStartedAt  *time.Time `bson:"judgingStartedAt"`
	JudgingFinishedAt *time.Time `bson:"judgingFinishedAt"`
}

func NewJudgmentRepository(
	db *gorm.DB,
	mdb *mongo.Client,
	logger sharedtypes.Logger,
) repository.JudgmentRepository {
	logger.Infof("[JudgmentRepository] Initializing new judgment repository")
	return &JudgmentRepository{
		db:     db,
		mdb:    mdb,
		logger: logger,
	}
}

func parseLifecycles(lifecycleType string, lifecycles []Lifecycle) []*lifecycle.Lifecycle {
	ls := []*lifecycle.Lifecycle{}
	for _, l := range lifecycles {
		ls = append(ls, lifecycle.NewLifecycle(&lifecycle.LifecycleConfig{
			ID:                   l.ID,
			ExecutionEnvironment: l.ExecutionEnvironment,
			Type:                 lifecycleType,
			Command:              l.Command,
		}))
	}
	return ls
}

func (r *JudgmentRepository) FindByID(id string) (*judgment.Judgment, error) {
	r.logger.Infof("[JudgmentRepository] Finding judgment by ID: %s", id)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := mongodb.NewRetryCollection(r.mdb.Database("judge").Collection("judgments"), r.logger)
	filter := bson.M{"_id": id}
	projection := bson.M{
		"_id":               1,
		"contestProjectID":  1,
		"submissionID":      1,
		"rejudgeID":         1,
		"project":           1,
		"status":            1,
		"score":             1,
		"totalScore":        1,
		"judgeType":         1,
		"gradingItems":      1,
		"lifecycle":         1,
		"reason":            1,
		"errorMessage":      1,
		"staffNote":         1,
		"createdAt":         1,
		"judgingStartedAt":  1,
		"judgingFinishedAt": 1,
	}

	r.logger.Debugf("[JudgmentRepository] Executing find operation with filter: %+v", filter)
	result := collection.FindOne(ctx, filter, options.FindOne().SetProjection(projection))
	if result.Err() != nil {
		r.logger.Errorf("[JudgmentRepository] Failed to find judgment %s: %v", id, result.Err())
		return nil, result.Err()
	}

	var j *Judgment
	if err := result.Decode(&j); err != nil {
		r.logger.Errorf("[JudgmentRepository] Failed to decode judgment %s: %v", id, err)
		return nil, err
	}

	r.logger.Debugf("[JudgmentRepository] Successfully decoded judgment %s with status: %s", id, j.Status)

	domainModel, err := r.toDomainModel(j)
	if err != nil {
		r.logger.Errorf("[JudgmentRepository] Failed to convert judgment %s to domain model: %v", id, err)
		return nil, err
	}

	r.logger.Infof("[JudgmentRepository] Successfully found and converted judgment %s", id)
	return domainModel, nil
}

func transformLifecycles(lifeycles []lifecycle.Lifecycle) []bson.M {
	lifecycles := []bson.M{}
	for _, l := range lifeycles {
		var result bson.M
		if l.Result() != nil {
			result = bson.M{
				"exitCode": l.Result().ExitCode(),
				"execTime": l.Result().ExecTime(),
				"stdout":   l.Result().Stdout(),
				"stderr":   l.Result().Stderr(),
			}
		}

		lifecycles = append(lifecycles, bson.M{
			"id":                   l.ID(),
			"executionEnvironment": l.ExecutionEnvironment().String(),
			"command":              l.Command().String(),
			"result":               result,
		})
	}

	return lifecycles
}

func transformGradingItems(gradingItems []*gradingitem.GradingItem) []bson.M {
	items := []bson.M{}
	for _, gradingItem := range gradingItems {
		testcases := []bson.M{}
		for _, testcase := range gradingItem.Testcases() {
			beforeEach := transformLifecycles(testcase.Lifecycle().BeforeEach())
			afterEach := transformLifecycles(testcase.Lifecycle().AfterEach())
			var result bson.M
			if testcase.Result() != nil {
				result = bson.M{
					"exitCode": testcase.Result().ExitCode(),
					"execTime": testcase.Result().ExecTime(),
					"stdout":   testcase.Result().Stdout(),
					"stderr":   testcase.Result().Stderr(),
				}
			}
			// "score" is the score this testcase actually earned and "totalScore"
			// its configured max, matching how the grading item and judgment levels
			// already name the pair. "parsedScore" stays alongside as the raw
			// parser output the earned score is derived from.
			testcases = append(testcases, bson.M{
				"id":          testcase.ID(),
				"name":        testcase.Name(),
				"score":       gradingItem.ScoringStrategy().TestcaseScore(testcase),
				"totalScore":  testcase.Score(),
				"parsedScore": testcase.ParsedScore(),
				"status":      testcase.Status().String(),
				"command":     testcase.Command().String(),
				"lifecycle": bson.M{
					"beforeEach": beforeEach,
					"afterEach":  afterEach,
				},
				"result": result,
			})
		}

		items = append(items, bson.M{
			"id":              gradingItem.ID(),
			"label":           gradingItem.Label(),
			"name":            gradingItem.Name(),
			"scoringStrategy": gradingItem.ScoringStrategy().String(),
			"resultParser":    gradingItem.ResultParser().String(),
			"score":           gradingItem.CalculateScore(),
			"totalScore":      gradingItem.Score(),
			"status":          gradingItem.Status().String(),
			"testcases":       testcases,
		})
	}
	return items
}

func transformProject(p *project.Project) bson.M {
	return bson.M{
		"id":           p.ID(),
		"imageTag":     p.ImageTag(),
		"testImageTag": p.TestImageTag(),
		"env":          p.Env(),
		"testEnv":      p.TestEnv(),
		"mountPath":    p.MountPath(),
		"resultParser": p.ResultParser(),
	}
}

func transformJudgmentLifecycle(l *lifecycle.JudgmentLifecycle) bson.M {
	return bson.M{
		"beforeAll": transformLifecycles(l.BeforeAll()),
		"afterAll":  transformLifecycles(l.AfterAll()),
	}
}

func (r *JudgmentRepository) toDomainModel(j *Judgment) (*judgment.Judgment, error) {
	judgeStatus, ok := judgestatus.JudgeStatusMap[j.Status]
	if !ok {
		return nil, fmt.Errorf("judgment %s: unrecognized status %q", j.ID, j.Status)
	}

	judgeType, ok := judgetype.JudgeTypeMap[j.JudgeType]
	if !ok {
		return nil, fmt.Errorf("judgment %s: unrecognized judgeType %q", j.ID, j.JudgeType)
	}

	gradingItems := []*gradingitem.GradingItem{}
	for _, gradingItem := range j.GradingItems {
		testcases := []*testcase.Testcase{}
		for _, tc := range gradingItem.Testcases {
			beforeEach := parseLifecycles(lifecycle.BeforeEach.String(), tc.Lifecycle.BeforeEach)
			afterEach := parseLifecycles(lifecycle.AfterEach.String(), tc.Lifecycle.AfterEach)
			t := testcase.NewTestcase(&testcase.TestcaseConfig{
				ID:          tc.ID,
				Name:        tc.Name,
				Command:     tc.Command,
				Score:       tc.TotalScore,
				ParsedScore: tc.ParsedScore,
			})
			t.SetLifecycle(lifecycle.NewTestcaseLifecycle(append(beforeEach, afterEach...)))
			if tc.Result != nil {
				result := execresult.NewExecResult(&execresult.ExecResultConfig{
					ExitCode: int(tc.Result.ExitCode),
					ExecTime: tc.Result.ExecTime,
					Stdout:   tc.Result.Stdout,
					Stderr:   tc.Result.Stderr,
				})
				t.SetResult(result)
			}
			status, ok := status.StatusMap[tc.Status]
			if !ok {
				return nil, fmt.Errorf("judgment %s: testcase %d: unrecognized status %q", j.ID, tc.ID, tc.Status)
			}
			t.SetStatus(status)
			testcases = append(testcases, t)
		}

		scoringStrategy, ok := scoringstrategy.Strategies[gradingItem.ScoringStrategy]
		if !ok {
			return nil, fmt.Errorf("judgment %s: gradingItem %d: unrecognized scoringStrategy %q", j.ID, gradingItem.ID, gradingItem.ScoringStrategy)
		}

		parser := resultparser.Parsers[gradingItem.ResultParser]
		if parser == nil {
			parser = resultparser.AutoParser
		}
		gi := gradingitem.NewGradingItem(&gradingitem.GradingItemConfig{
			ID:              gradingItem.ID,
			Label:           gradingItem.Label,
			Name:            gradingItem.Name,
			Score:           gradingItem.TotalScore,
			ScoringStrategy: scoringStrategy,
			Testcases:       testcases,
			ResultParser:    parser,
		})
		gradingItems = append(gradingItems, gi)
	}

	judgment := judgment.New(&judgment.JudgmentConfig{
		ContestProjectID: j.ContestProjectID,
		Project: project.NewProject(&project.ProjectConfig{
			ID:           j.Project.ID,
			ImageTag:     j.Project.ImageTag,
			TestImageTag: j.Project.TestImageTag,
			Env:          strings.Join(j.Project.ImageEnv, "\n"),
			TestEnv:      strings.Join(j.Project.TestImageEnv, "\n"),
			MountPath:    j.Project.MountPath,
			ResultParser: j.Project.ResultParser,
		}),
		SubmissionID:      j.SubmissionID,
		Status:            judgeStatus,
		JudgeType:         judgeType,
		Reason:            j.Reason,
		CreatedAt:         j.CreatedAt,
		JudgingStartedAt:  j.JudgingStartedAt,
		JudgingFinishedAt: j.JudgingFinishedAt,
		GradingItems:      gradingItems,
		Lifecycle: lifecycle.NewJudgmentLifecycle(append(
			parseLifecycles(lifecycle.BeforeAll.String(), j.Lifecycle.BeforeAll),
			parseLifecycles(lifecycle.AfterAll.String(), j.Lifecycle.AfterAll)...,
		)),
	})

	judgment.SetID(j.ID)
	judgment.SetScore(j.Score)
	if j.RejudgeID != nil {
		judgment.SetRejudgeID(*j.RejudgeID)
	}
	judgment.SetStaffNote(j.StaffNote)
	if j.ErrorMessage != nil {
		judgment.SetErrorMessage(*j.ErrorMessage)
	}
	if j.JudgingStartedAt != nil {
		judgment.SetJudgingStartedAt(*j.JudgingStartedAt)
	}
	if j.JudgingFinishedAt != nil {
		judgment.SetJudgingFinishedAt(*j.JudgingFinishedAt)
	}

	return judgment, nil
}

// judgmentPayload builds the full-document $set payload shared by Save,
// SaveProgressIfJudging, and SaveTerminalIfJudging — a single place for all
// ~20 fields so the three save variants can never drift out of sync with
// each other.
func judgmentPayload(j *judgment.Judgment) bson.M {
	return bson.M{
		"_id":               j.ID(),
		"contestProjectID":  j.ContestProjectID(),
		"submissionID":      j.SubmissionID(),
		"rejudgeID":         j.RejudgeID(),
		"project":           transformProject(j.Project()),
		"gradingItems":      transformGradingItems(j.GradingItems()),
		"lifecycle":         transformJudgmentLifecycle(j.Lifecycle()),
		"status":            j.Status().String(),
		"judgeType":         j.JudgeType().String(),
		"score":             j.CalculateScore(),
		"totalScore":        j.Score(),
		"reason":            j.Reason(),
		"errorMessage":      j.ErrorMessage(),
		"staffNote":         j.StaffNote(),
		"createdAt":         j.CreatedAt(),
		"judgingStartedAt":  j.JudgingStartedAt(),
		"judgingFinishedAt": j.JudgingFinishedAt(),
	}
}

func (r *JudgmentRepository) Save(j *judgment.Judgment) error {
	r.logger.Infof("[JudgmentRepository] Starting to save judgment %s", j.ID())
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := mongodb.NewRetryCollection(r.mdb.Database("judge").Collection("judgments"), r.logger)

	if j.ID() == "" {
		j.SetID(primitive.NewObjectID().Hex())
		r.logger.Infof("[JudgmentRepository] Generated new ID for judgment: %s", j.ID())
	}

	filter := bson.M{"_id": j.ID()}
	payload := judgmentPayload(j)

	r.logger.Debugf("[JudgmentRepository] Saving judgment with payload: %+v", payload)

	if _, err := collection.UpdateOne(ctx, filter, bson.M{"$set": payload}, options.Update().SetUpsert(true)); err != nil {
		r.logger.Errorf("[JudgmentRepository] Failed to save judgment %s: %v", j.ID(), err)
		return err
	}

	r.logger.Infof("[JudgmentRepository] Successfully saved judgment %s with status: %s", j.ID(), j.Status().String())
	return nil
}

// judgmentsCollection returns the raw collection handle, deliberately NOT
// wrapped in mongodb.RetryCollection. It is used by the state-transition
// methods (TryStart/SaveTerminalIfJudging/TryCancel), where a single attempt
// is what makes MatchedCount meaningful.
//
// RetryCollection retries a transient failure by resending the identical
// command. For a compare-and-set that is not safe: if the first attempt
// actually applied and only its acknowledgment was lost, the resend's filter
// no longer matches, and the caller is told MatchedCount==0 for an update
// that did land — a blind retry manufactures exactly the ambiguity these
// methods must not have. Sending one command keeps applied=false as close to
// "nothing was written" as this layer can get.
//
// Retrying writes is not thereby abandoned, just moved to the layer that can
// do it correctly: the driver's own retryable writes tag a write with a
// transaction number so the server recognizes a resend and replies with the
// original outcome instead of re-evaluating the filter. That is idempotent in
// a way an application-level resend cannot be. (It requires a replica set or
// sharded cluster; against a standalone mongod no layer retries these, and a
// transient error surfaces as an error — which callers already handle by not
// assuming ownership.)
//
// Whole-document writes stay on RetryCollection: Save is unconditional, and
// SaveProgressIfJudging's filter matches the same status it writes (Judging),
// so a resend still matches and retrying is harmless.
func (r *JudgmentRepository) judgmentsCollection() *mongo.Collection {
	return r.mdb.Database("judge").Collection("judgments")
}

// TryStart atomically moves judgmentID from Queued to Judging. See the
// interface doc comment (domain/repository/judgment_repository.go), and
// judgmentsCollection for why this issues a single un-retried command.
func (r *JudgmentRepository) TryStart(judgmentID string, startedAt time.Time) (bool, error) {
	r.logger.Infof("[JudgmentRepository] Attempting to start judgment %s", judgmentID)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := r.judgmentsCollection()

	filter := bson.M{"_id": judgmentID, "status": judgestatus.QUEUED.String()}
	update := bson.M{"$set": bson.M{
		"status":           judgestatus.JUDGING.String(),
		"judgingStartedAt": startedAt,
	}}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Errorf("[JudgmentRepository] Failed to start judgment %s: %v", judgmentID, err)
		return false, err
	}
	if result.MatchedCount == 0 {
		r.logger.Infof("[JudgmentRepository] Judgment %s was no longer Queued, did not acquire it", judgmentID)
		return false, nil
	}
	r.logger.Infof("[JudgmentRepository] Started judgment %s", judgmentID)
	return true, nil
}

// SaveProgressIfJudging persists j's full state conditionally on the
// document's CURRENT status still being Judging. See the interface doc
// comment for why retrying this specific conditional update needs no
// disambiguation (unlike TryStart/TryCancel/SaveTerminalIfJudging).
func (r *JudgmentRepository) SaveProgressIfJudging(j *judgment.Judgment) (bool, error) {
	// The filter below only constrains the DOCUMENT's current status; the
	// payload carries the ENTITY's status. Without this check, the method's
	// name would hold by caller convention alone — a terminal entity passed
	// in here would flip the document terminal through the progress path,
	// bypassing SaveTerminalIfJudging's semantics.
	if !j.Status().IsJudging() {
		return false, fmt.Errorf("SaveProgressIfJudging called for judgment %s with entity status %s; only a Judging entity may be saved as progress", j.ID(), j.Status().String())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := mongodb.NewRetryCollection(r.mdb.Database("judge").Collection("judgments"), r.logger)

	filter := bson.M{"_id": j.ID(), "status": judgestatus.JUDGING.String()}
	payload := judgmentPayload(j)

	result, err := collection.UpdateOne(ctx, filter, bson.M{"$set": payload})
	if err != nil {
		r.logger.Errorf("[JudgmentRepository] Failed to save progress for judgment %s: %v", j.ID(), err)
		return false, err
	}

	applied := result.MatchedCount > 0
	if !applied {
		r.logger.Infof("[JudgmentRepository] Judgment %s was no longer Judging, skipped progress save", j.ID())
	}
	return applied, nil
}

// SaveTerminalIfJudging persists j's full state (a computed terminal
// outcome) conditionally on the document's CURRENT status still being
// Judging. See the interface doc comment (domain/repository/judgment_repository.go),
// and judgmentsCollection for why this issues a single un-retried command.
func (r *JudgmentRepository) SaveTerminalIfJudging(j *judgment.Judgment) (bool, error) {
	// Mirror of SaveProgressIfJudging's entity check: the filter matches a
	// Judging DOCUMENT, so a still-Judging ENTITY would match, write Judging
	// back, and report a "terminal save" that stored nothing terminal.
	if !j.Status().IsTerminal() {
		return false, fmt.Errorf("SaveTerminalIfJudging called for judgment %s with non-terminal entity status %s", j.ID(), j.Status().String())
	}
	r.logger.Infof("[JudgmentRepository] Attempting terminal save of judgment %s with status: %s", j.ID(), j.Status().String())
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := r.judgmentsCollection()

	filter := bson.M{"_id": j.ID(), "status": judgestatus.JUDGING.String()}
	payload := judgmentPayload(j)

	result, err := collection.UpdateOne(ctx, filter, bson.M{"$set": payload})
	if err != nil {
		r.logger.Errorf("[JudgmentRepository] Failed to save terminal judgment %s: %v", j.ID(), err)
		return false, err
	}
	if result.MatchedCount == 0 {
		// Deliberately NOT disambiguated by re-reading the status and
		// treating "already at the status we wanted" as applied. Two
		// different writers routinely converge on the SAME terminal status
		// — most commonly Cancelled, when a cancel request lands in the DB
		// and this worker's own context is cancelled a moment later — and
		// the status alone cannot distinguish "my write landed and only its
		// acknowledgment was lost" from "someone else wrote this status".
		// Reporting applied=true on that evidence would let the caller
		// publish THIS worker's in-memory grading/testcase snapshot, which
		// this very update just failed to persist, as though it were the
		// stored state. Matched-zero is authoritative: the caller re-reads
		// and publishes the actual document instead.
		r.logger.Infof("[JudgmentRepository] Judgment %s was no longer Judging, skipped terminal save to %s", j.ID(), j.Status().String())
		return false, nil
	}
	r.logger.Infof("[JudgmentRepository] Saved terminal judgment %s with status: %s", j.ID(), j.Status().String())
	return true, nil
}

// TryCancel atomically moves judgmentID from Queued or Judging to Cancelled.
// See the interface doc comment (domain/repository/judgment_repository.go)
// for why this touches only status/reason/judgingFinishedAt, never the full
// document, and judgmentsCollection for why it issues a single un-retried
// command. FindOneAndUpdate (ReturnDocument=After) makes the write and the
// read of the resulting document ONE command: the winner gets the entity it
// must publish from with no second read that could fail independently.
func (r *JudgmentRepository) TryCancel(judgmentID string, reason string, finishedAt time.Time) (*judgment.Judgment, error) {
	r.logger.Infof("[JudgmentRepository] Attempting to cancel judgment %s", judgmentID)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := r.judgmentsCollection()

	filter := bson.M{
		"_id":    judgmentID,
		"status": bson.M{"$in": []string{judgestatus.QUEUED.String(), judgestatus.JUDGING.String()}},
	}
	set := bson.M{
		"status":            judgestatus.CANCELLED.String(),
		"judgingFinishedAt": finishedAt,
	}
	if reason != "" {
		set["reason"] = reason
	}

	var doc *Judgment
	err := collection.FindOneAndUpdate(ctx, filter, bson.M{"$set": set},
		options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		// No match means THIS call wrote nothing — including its reason,
		// which may differ from whatever reason the writer that got there
		// first recorded. Returning a judgment here would claim this
		// operation's cancellation had been stored when it hadn't.
		r.logger.Infof("[JudgmentRepository] Judgment %s was no longer Queued/Judging, skipped cancel", judgmentID)
		return nil, nil
	}
	if err != nil {
		r.logger.Errorf("[JudgmentRepository] Failed to cancel judgment %s: %v", judgmentID, err)
		return nil, err
	}

	cancelled, err := r.toDomainModel(doc)
	if err != nil {
		r.logger.Errorf("[JudgmentRepository] Cancelled judgment %s but failed to convert the returned document: %v", judgmentID, err)
		return nil, err
	}
	r.logger.Infof("[JudgmentRepository] Cancelled judgment %s", judgmentID)
	return cancelled, nil
}

// UpdateStaffNote sets (or, with nil, clears) the judge's note on a judgment
// and returns the judgment as stored afterwards, so the caller can publish the
// update. It is deliberately NOT one of the conditional compare-and-set
// methods above: a note may be written to a judgment in any state, and last
// write wins — one contest, one judging desk, no history. What it shares with
// TryCancel is the shape: a single $set of ONE field on a filter of _id only.
// A read-modify-Save would race a worker mid-run and roll its grading
// progress back, which is why writing the whole payload is off the table.
func (r *JudgmentRepository) UpdateStaffNote(judgmentID string, note *string) (*judgment.Judgment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := r.judgmentsCollection()

	var doc *Judgment
	err := collection.FindOneAndUpdate(ctx, bson.M{"_id": judgmentID}, bson.M{"$set": bson.M{"staffNote": note}},
		options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		r.logger.Warnf("[JudgmentRepository] Judgment %s not found, staff note not written", judgmentID)
		return nil, nil
	}
	if err != nil {
		r.logger.Errorf("[JudgmentRepository] Failed to write staff note on judgment %s: %v", judgmentID, err)
		return nil, err
	}

	updated, err := r.toDomainModel(doc)
	if err != nil {
		r.logger.Errorf("[JudgmentRepository] Wrote staff note on judgment %s but failed to convert the returned document: %v", judgmentID, err)
		return nil, err
	}
	return updated, nil
}

func (r *JudgmentRepository) FindJudgings() ([]*judgment.Judgment, error) {
	r.logger.Infof("[JudgmentRepository] Loading judging judgments")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := mongodb.NewRetryCollection(r.mdb.Database("judge").Collection("judgments"), r.logger)
	filter := bson.M{"status": judgestatus.JUDGING.String()}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		r.logger.Errorf("[JudgmentRepository] Failed to find judging judgments: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var judgments []*Judgment
	if err := cursor.All(ctx, &judgments); err != nil {
		r.logger.Errorf("[JudgmentRepository] Failed to decode judging judgments: %v", err)
		return nil, err
	}

	r.logger.Infof("[JudgmentRepository] Found %d judging judgments", len(judgments))

	var domainModels []*judgment.Judgment
	for _, j := range judgments {
		domainModel, err := r.toDomainModel(j)
		if err != nil {
			r.logger.Errorf("[JudgmentRepository] Failed to convert judgment %s to domain model: %v", j.ID, err)
			continue
		}
		domainModels = append(domainModels, domainModel)
	}

	r.logger.Infof("[JudgmentRepository] Successfully loaded %d judging judgments", len(domainModels))
	return domainModels, nil
}

func (r *JudgmentRepository) LoadQueuedJudgments() ([]*judgment.Judgment, error) {
	r.logger.Infof("[JudgmentRepository] Loading queued judgments")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := mongodb.NewRetryCollection(r.mdb.Database("judge").Collection("judgments"), r.logger)
	filter := bson.M{"status": judgestatus.QUEUED.String()}

	r.logger.Debugf("[JudgmentRepository] Executing find operation with filter: %+v", filter)
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		r.logger.Errorf("[JudgmentRepository] Failed to find queued judgments: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var judgments []*Judgment
	if err := cursor.All(ctx, &judgments); err != nil {
		r.logger.Errorf("[JudgmentRepository] Failed to decode queued judgments: %v", err)
		return nil, err
	}

	r.logger.Infof("[JudgmentRepository] Found %d queued judgments", len(judgments))

	var domainModels []*judgment.Judgment
	for _, j := range judgments {
		domainModel, err := r.toDomainModel(j)
		if err != nil {
			r.logger.Errorf("[JudgmentRepository] Failed to convert judgment %s to domain model: %v", j.ID, err)
			continue
		}
		domainModels = append(domainModels, domainModel)
	}

	r.logger.Infof("[JudgmentRepository] Successfully loaded %d queued judgments", len(domainModels))
	return domainModels, nil
}
