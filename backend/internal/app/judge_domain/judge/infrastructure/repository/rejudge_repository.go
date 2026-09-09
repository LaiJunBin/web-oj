package repository

import (
	"context"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgestatus"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/rejudge"
	"judge-backend/internal/app/judge_domain/judge/domain/repository"
	sharedtypes "judge-backend/internal/app/shared/types"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"judge-backend/internal/app/shared/infrastructure/mongodb"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RejudgeRepository struct {
	mdb    *mongo.Client
	logger sharedtypes.Logger
}

func NewRejudgeRepository(
	mdb *mongo.Client,
	logger sharedtypes.Logger,
) repository.RejudgeRepository {
	logger.Infof("[RejudgeRepository] Initializing new rejudge repository")
	return &RejudgeRepository{
		mdb:    mdb,
		logger: logger,
	}
}

func (r *RejudgeRepository) Save(rejudge *rejudge.Rejudge) error {
	r.logger.Infof("[RejudgeRepository] Starting to save rejudge %s", rejudge.ID())
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := mongodb.NewRetryCollection(r.mdb.Database("judge").Collection("rejudges"), r.logger)
	filter := bson.M{"_id": rejudge.ID()}

	payload := bson.M{
		"_id":              rejudge.ID(),
		"contestProjectID": rejudge.ContestProjectID(),
		"judgments":        rejudge.Judgments(),
		"reason":           rejudge.Reason(),
		"status":           rejudge.Status().String(),
		"createdAt":        rejudge.CreatedAt(),
	}

	r.logger.Debugf("[RejudgeRepository] Saving rejudge with payload: %+v", payload)

	if _, err := collection.UpdateOne(ctx, filter, bson.M{"$set": payload}, options.Update().SetUpsert(true)); err != nil {
		r.logger.Errorf("[RejudgeRepository] Failed to save rejudge %s: %v", rejudge.ID(), err)
		return err
	}

	r.logger.Infof("[RejudgeRepository] Successfully saved rejudge %s with status: %s", rejudge.ID(), rejudge.Status().String())
	return nil
}

type Rejudge struct {
	ID               string    `bson:"_id"`
	ContestProjectID int32     `bson:"contestProjectID"`
	Judgments        []string  `bson:"judgments"`
	Reason           string    `bson:"reason"`
	Status           string    `bson:"status"`
	CreatedAt        time.Time `bson:"createdAt"`
}

func (r *RejudgeRepository) FindByID(id string) (*rejudge.Rejudge, error) {
	r.logger.Infof("[RejudgeRepository] Finding rejudge by ID: %s", id)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := mongodb.NewRetryCollection(r.mdb.Database("judge").Collection("rejudges"), r.logger)

	filter := bson.M{"_id": id}
	r.logger.Debugf("[RejudgeRepository] Executing find operation with filter: %+v", filter)

	var result *Rejudge
	if err := collection.FindOne(ctx, filter).Decode(&result); err != nil {
		r.logger.Errorf("[RejudgeRepository] Failed to find rejudge %s: %v", id, err)
		return nil, err
	}

	r.logger.Debugf("[RejudgeRepository] Successfully decoded rejudge %s with status: %s", id, result.Status)

	status, err := rejudge.NewStatus(result.Status)
	if err != nil {
		r.logger.Errorf("[RejudgeRepository] Failed to create status for rejudge %s: %v", id, err)
		return nil, err
	}

	rejudge := rejudge.New(&rejudge.RejudgeConfig{
		ID:               result.ID,
		ContestProjectID: result.ContestProjectID,
		Judgments:        result.Judgments,
		Reason:           result.Reason,
		Status:           status,
		CreatedAt:        result.CreatedAt,
	})

	r.logger.Infof("[RejudgeRepository] Successfully found and converted rejudge %s", id)
	return rejudge, nil
}

// TransitionStatus performs a single conditional update: `status` is only set
// to target.String() when the document's CURRENT status is one of
// allowedFrom. The condition and the write happen atomically inside MongoDB
// (a single UpdateOne with a status filter), so there is no read-modify-write
// gap for a concurrent caller to land a lost update in. See the interface
// doc comment (domain/repository/rejudge_repository.go) for why this exists.
func (r *RejudgeRepository) TransitionStatus(rejudgeID string, allowedFrom []*rejudge.Status, target *rejudge.Status) (bool, error) {
	r.logger.Infof("[RejudgeRepository] Attempting to transition rejudge %s to %s", rejudgeID, target.String())
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := mongodb.NewRetryCollection(r.mdb.Database("judge").Collection("rejudges"), r.logger)

	allowedValues := make([]string, len(allowedFrom))
	for i, s := range allowedFrom {
		allowedValues[i] = s.String()
	}

	filter := bson.M{"_id": rejudgeID, "status": bson.M{"$in": allowedValues}}
	update := bson.M{"$set": bson.M{"status": target.String()}}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Errorf("[RejudgeRepository] Failed to transition rejudge %s to %s: %v", rejudgeID, target.String(), err)
		return false, err
	}

	applied := result.MatchedCount > 0
	if applied {
		r.logger.Infof("[RejudgeRepository] Transitioned rejudge %s to %s", rejudgeID, target.String())
	} else {
		r.logger.Debugf("[RejudgeRepository] Rejudge %s was not in an allowed state (%v) for transition to %s, skipped", rejudgeID, allowedValues, target.String())
	}
	return applied, nil
}

// CommitJudgmentsByRejudgeID finalizes an admin's approve/reject decision.
//
// Deliberately NOT a multi-document transaction: the deployed MongoDB is a
// standalone instance, which rejects transactions outright (IllegalOperation:
// "Transaction numbers are only allowed on a replica set member or mongos") —
// and every other race in this codebase is already settled with
// single-document conditional writes for the same reason. The conditional
// rejudge flip below is the race arbiter: two admins deciding the same
// rejudge both pass the application layer's CanApprove/CanReject check (a
// stale read, not a lock), and only the request whose update still matches
// the FINISHED document commits; the loser gets applied=false. The judgments
// update runs after the flip, so a crash between the two writes leaves an
// APPROVED rejudge whose judgments still read Rejudged —
// RepairUncommittedApprovals closes exactly that window, at startup and on
// its periodic sweep.
func (r *RejudgeRepository) CommitJudgmentsByRejudgeID(rejudgeID string, approval bool) (bool, error) {
	r.logger.Infof("[RejudgeRepository] Starting to commit judgments for rejudge %s with approval: %v", rejudgeID, approval)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var finalStatus *rejudge.Status
	if approval {
		finalStatus = rejudge.APPROVED
	} else {
		finalStatus = rejudge.REJECTED
	}

	rejudgesCollection := mongodb.NewRetryCollection(r.mdb.Database("judge").Collection("rejudges"), r.logger)
	result, err := rejudgesCollection.UpdateOne(ctx,
		bson.M{"_id": rejudgeID, "status": rejudge.FINISHED.String()},
		bson.M{"$set": bson.M{"status": finalStatus.String()}},
	)
	if err != nil {
		r.logger.Errorf("[RejudgeRepository] Failed to update rejudge status for %s: %v", rejudgeID, err)
		return false, err
	}
	if result.MatchedCount == 0 {
		r.logger.Debugf("[RejudgeRepository] Rejudge %s was no longer Finished, skipping commit", rejudgeID)
		return false, nil
	}

	if approval {
		r.logger.Debugf("[RejudgeRepository] Updating judgments status to JUDGED for rejudge %s", rejudgeID)
		judgmentsCollection := mongodb.NewRetryCollection(r.mdb.Database("judge").Collection("judgments"), r.logger)
		if _, err := judgmentsCollection.UpdateMany(ctx, bson.M{"rejudgeID": rejudgeID}, bson.M{"$set": bson.M{"status": judgestatus.JUDGED.String()}}); err != nil {
			// The decision itself is already durably APPROVED above; only the
			// judgments flip is missing, which the periodic repair replays.
			r.logger.Errorf("[RejudgeRepository] Failed to update judgments for rejudge %s (the periodic repair sweep will retry within minutes): %v", rejudgeID, err)
			return false, err
		}
	}

	r.logger.Infof("[RejudgeRepository] Successfully committed judgments for rejudge %s with status: %s", rejudgeID, finalStatus.String())
	return true, nil
}

// RepairUncommittedApprovals finishes approvals whose judgments flip was lost
// to a crash between CommitJudgmentsByRejudgeID's two writes: it finds
// judgments still in Rejudged whose rejudge is already Approved, and marks
// them Judged. Judgments of REJECTED (or still undecided) rejudges also sit
// in Rejudged — that is their intended terminal state, so they are filtered
// out by the Approved-status lookup, not repaired.
func (r *RejudgeRepository) RepairUncommittedApprovals() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	judgmentsCollection := mongodb.NewRetryCollection(r.mdb.Database("judge").Collection("judgments"), r.logger)
	candidateCursor, err := judgmentsCollection.Find(ctx,
		bson.M{"status": judgestatus.REJUDGED.String(), "rejudgeID": bson.M{"$ne": nil}},
		options.Find().SetProjection(bson.M{"rejudgeID": 1}),
	)
	if err != nil {
		return 0, err
	}
	var candidates []struct {
		RejudgeID string `bson:"rejudgeID"`
	}
	if err := candidateCursor.All(ctx, &candidates); err != nil {
		return 0, err
	}
	seen := map[string]struct{}{}
	candidateIDs := []string{}
	for _, c := range candidates {
		if _, dup := seen[c.RejudgeID]; c.RejudgeID != "" && !dup {
			seen[c.RejudgeID] = struct{}{}
			candidateIDs = append(candidateIDs, c.RejudgeID)
		}
	}
	if len(candidateIDs) == 0 {
		return 0, nil
	}

	rejudgesCollection := mongodb.NewRetryCollection(r.mdb.Database("judge").Collection("rejudges"), r.logger)
	cursor, err := rejudgesCollection.Find(ctx, bson.M{
		"_id":    bson.M{"$in": candidateIDs},
		"status": rejudge.APPROVED.String(),
	})
	if err != nil {
		return 0, err
	}
	var approved []Rejudge
	if err := cursor.All(ctx, &approved); err != nil {
		return 0, err
	}
	if len(approved) == 0 {
		return 0, nil
	}

	approvedIDs := make([]string, len(approved))
	for i, rj := range approved {
		approvedIDs[i] = rj.ID
	}
	result, err := judgmentsCollection.UpdateMany(ctx,
		bson.M{"rejudgeID": bson.M{"$in": approvedIDs}, "status": judgestatus.REJUDGED.String()},
		bson.M{"$set": bson.M{"status": judgestatus.JUDGED.String()}},
	)
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, nil
}

func (r *RejudgeRepository) AllJudgmentsFinished(rejudgeID string) (bool, error) {
	r.logger.Infof("[RejudgeRepository] Checking if all judgments are finished for rejudge %s", rejudgeID)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := mongodb.NewRetryCollection(r.mdb.Database("judge").Collection("judgments"), r.logger)

	r.logger.Debugf("[RejudgeRepository] Counting total judgments for rejudge %s", rejudgeID)
	rejudgeCount, err := collection.CountDocuments(ctx, bson.M{"rejudgeID": rejudgeID})
	if err != nil {
		r.logger.Errorf("[RejudgeRepository] Failed to count total judgments for rejudge %s: %v", rejudgeID, err)
		return false, err
	}

	r.logger.Debugf("[RejudgeRepository] Counting rejudged judgments for rejudge %s", rejudgeID)
	rejudgedCount, err := collection.CountDocuments(ctx, bson.M{"rejudgeID": rejudgeID, "status": judgestatus.REJUDGED.String()})
	if err != nil {
		r.logger.Errorf("[RejudgeRepository] Failed to count rejudged judgments for rejudge %s: %v", rejudgeID, err)
		return false, err
	}

	isFinished := rejudgeCount == rejudgedCount
	r.logger.Infof("[RejudgeRepository] Rejudge %s: %d/%d judgments finished (isFinished: %v)", rejudgeID, rejudgedCount, rejudgeCount, isFinished)
	return isFinished, nil
}
