package query

import (
	"context"
	"judge-backend/internal/app/contest_domain/contest/application/query"
	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	"judge-backend/internal/app/shared/model"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

type SubmissionQuery struct {
	db     *gorm.DB
	mdb    *mongo.Client
	logger sharedtypes.Logger
}

func NewSubmissionQuery(
	db *gorm.DB,
	mdb *mongo.Client,
	logger sharedtypes.Logger,
) query.SubmissionQuery {
	return &SubmissionQuery{
		db:     db,
		mdb:    mdb,
		logger: logger,
	}
}

type SubmissionInfo struct {
	ID         string    `bson:"_id"`
	Status     string    `bson:"status"`
	IsPartial  bool      `bson:"isPartial"`
	Score      float64   `bson:"score"`
	TotalScore float64   `bson:"totalScore"`
	Reason     *string   `bson:"reason"`
	StaffNote  *string   `bson:"staffNote"`
	CreatedAt  time.Time `bson:"createdAt"`
}

func (q *SubmissionQuery) ListUserProjectSubmissionsByProjectID(userID, contestProjectID int32) ([]*dto.Submission, error) {
	q.logger.Debugf("Starting to list user project submissions by project ID", "userId", userID, "contestProjectId", contestProjectID)

	cp := &model.ContestProject{}
	if err := q.db.Preload("Contest").Preload("Project").First(cp, contestProjectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			q.logger.Debugf("Contest project not found", "contestProjectId", contestProjectID)
			return nil, httperror.New(http.StatusNotFound, "contest project not found")
		}
		q.logger.Errorf("Failed to find contest project", "contestProjectId", contestProjectID, "error", err)
		return nil, err
	}

	ctx := context.Background()
	collection := q.mdb.Database("judge").Collection("submissions")
	pipeline := mongo.Pipeline{
		{
			{
				Key: "$match",
				Value: bson.M{
					"userID":           userID,
					"contestProjectID": contestProjectID,
					"ignored":          false,
				},
			},
		},
		{
			{
				Key: "$lookup",
				Value: bson.M{
					"from": "judgments",
					"as":   "judgments",
					"let":  bson.M{"submissionID": "$_id"},
					"pipeline": mongo.Pipeline{
						{
							{
								Key: "$match",
								Value: bson.M{
									"$expr": bson.M{
										"$and": []interface{}{
											bson.M{"$eq": []interface{}{"$submissionID", "$$submissionID"}},
											bson.M{"$or": []interface{}{
												bson.M{"$ne": []interface{}{"$judgeType", "REJUDGE"}},
												bson.M{"$eq": []interface{}{"$status", "Judged"}},
											}},
										},
									},
								},
							},
						},
						{
							{
								Key: "$sort",
								Value: bson.M{
									"createdAt": -1,
								},
							},
						},
						{
							{
								Key:   "$limit",
								Value: 1,
							},
						},
					},
				},
			},
		},
		{
			{
				Key: "$set",
				Value: bson.M{
					"judgment": bson.M{"$arrayElemAt": []interface{}{"$judgments", 0}},
				},
			},
		},
		{
			{
				Key: "$match",
				Value: bson.M{
					"judgment._id": bson.M{"$exists": true},
				},
			},
		},
		{
			{
				Key: "$project",
				Value: bson.M{
					"_id":        1,
					"isPartial":  1,
					"createdAt":  1,
					"status":     "$judgment.status",
					"score":      "$judgment.score",
					"totalScore": "$judgment.totalScore",
					"reason":     "$judgment.reason",
					"staffNote":  "$judgment.staffNote",
				},
			},
		},
		{
			{
				Key: "$sort",
				Value: bson.M{
					"createdAt": -1,
				},
			},
		},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var result []SubmissionInfo
	if err := cursor.All(ctx, &result); err != nil {
		return nil, err
	}

	submissions := make([]*dto.Submission, 0, len(result))
	for _, r := range result {
		submissions = append(submissions, &dto.Submission{
			ID:         r.ID,
			Status:     r.Status,
			IsPartial:  r.IsPartial,
			Score:      r.Score,
			TotalScore: r.TotalScore,
			CreatedAt:  r.CreatedAt,
			Reason:     r.Reason,
			StaffNote:  r.StaffNote,
		})
	}

	q.logger.Debugf("Successfully listed user project submissions", "userId", userID, "contestProjectId", contestProjectID, "count", len(submissions))
	return submissions, nil
}

type SubmissionDetail struct {
	ID           string                  `bson:"_id"`
	Status       string                  `bson:"status"`
	IsPartial    bool                    `bson:"isPartial"`
	Score        float64                 `bson:"score"`
	TotalScore   float64                 `bson:"totalScore"`
	Reason       *string                 `bson:"reason"`
	StaffNote    *string                 `bson:"staffNote"`
	CreatedAt    time.Time               `bson:"createdAt"`
	GradingItems []SubmissionGradingItem `bson:"gradingItems"`
}

type SubmissionGradingItem struct {
	ID         int32                `bson:"id"`
	Score      float64              `bson:"score"`
	TotalScore float64              `bson:"totalScore"`
	Status     string               `bson:"status"`
	Testcases  []SubmissionTestcase `bson:"testcases"`
}

type SubmissionTestcase struct {
	ID       int32  `bson:"id"`
	ExecTime int64  `bson:"execTime"`
	ExitCode int32  `bson:"exitCode"`
	Stdout   string `bson:"stdout"`
	Stderr   string `bson:"stderr"`
	Status   string `bson:"status"`
}

// contestProjectIDOfSubmission resolves which contest project a submission
// belongs to, straight from the submission document. A missing document is
// the same 404 the aggregation below answers, so resolving early changes
// nothing observable.
func (q *SubmissionQuery) contestProjectIDOfSubmission(ctx context.Context, submissionID string) (int32, error) {
	var result struct {
		ContestProjectID int32 `bson:"contestProjectID"`
	}
	err := q.mdb.Database("judge").Collection("submissions").FindOne(ctx,
		bson.M{"_id": submissionID},
		options.FindOne().SetProjection(bson.M{"contestProjectID": 1}),
	).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, httperror.New(http.StatusNotFound, "submission not found")
		}
		return 0, err
	}
	return result.ContestProjectID, nil
}

// sampleTestcaseIDsForContestProject returns the ids of the sample testcases
// in the problem this contest project serves — and only that problem. The
// filter used to be a GLOBAL is_sample pluck, so a testcase of ANOTHER
// problem marked sample passed it; that only shows once a judgment carries
// foreign testcases, but the display layer must not be the thing standing
// between a data problem and the contestant's screen. Both subqueries run
// under gorm's default scope, so soft-deleted grading items and testcases
// stop exposing their output.
func (q *SubmissionQuery) sampleTestcaseIDsForContestProject(contestProjectID int32) ([]int32, error) {
	cp := &model.ContestProject{}
	if err := q.db.Select("project_id").Where("id = ?", contestProjectID).First(cp).Error; err != nil {
		return nil, err
	}

	ids := make([]int32, 0)
	if err := q.db.Model(&model.Testcase{}).
		Where("is_sample = ? AND grading_item_id IN (?)", true,
			q.db.Model(&model.GradingItem{}).Select("id").Where("project_id = ?", cp.ProjectID)).
		Pluck("id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

func (q *SubmissionQuery) GetUserSubmissionByID(userID int32, submissionID string) (*dto.SubmissionDetail, error) {
	q.logger.Debugf("Starting to get user submission by ID", "userId", userID, "submissionId", submissionID)

	ctx := context.Background()

	contestProjectID, err := q.contestProjectIDOfSubmission(ctx, submissionID)
	if err != nil {
		q.logger.Debugf("Failed to resolve submission's contest project", "submissionId", submissionID, "error", err)
		return nil, err
	}

	sampleTestcaseIDs, err := q.sampleTestcaseIDsForContestProject(contestProjectID)
	if err != nil {
		q.logger.Errorf("Failed to get sample testcase IDs", "contestProjectId", contestProjectID, "error", err)
		return nil, err
	}
	collection := q.mdb.Database("judge").Collection("submissions")
	pipeline := mongo.Pipeline{
		{
			{
				Key: "$match",
				Value: bson.M{
					"userID":  userID,
					"_id":     submissionID,
					"ignored": false,
				},
			},
		},
		{
			{
				Key: "$lookup",
				Value: bson.M{
					"from": "judgments",
					"as":   "judgments",
					"let":  bson.M{"submissionID": "$_id"},
					"pipeline": mongo.Pipeline{
						{
							{
								Key: "$match",
								Value: bson.M{
									"$expr": bson.M{
										"$and": []interface{}{
											bson.M{"$eq": []interface{}{"$submissionID", "$$submissionID"}},
											bson.M{"$or": []interface{}{
												bson.M{"$ne": []interface{}{"$judgeType", "REJUDGE"}},
												bson.M{"$eq": []interface{}{"$status", "Judged"}},
											}},
										},
									},
								},
							},
						},
						{
							{
								Key: "$sort",
								Value: bson.M{
									"createdAt": -1,
								},
							},
						},
						{
							{
								Key:   "$limit",
								Value: 1,
							},
						},
					},
				},
			},
		},
		{
			{
				Key: "$set",
				Value: bson.M{
					"judgment": bson.M{"$arrayElemAt": []interface{}{"$judgments", 0}},
				},
			},
		},
		{
			{
				Key: "$match",
				Value: bson.M{
					"judgment._id": bson.M{"$exists": true},
				},
			},
		},
		{
			{
				Key: "$project",
				Value: bson.M{
					"_id":        1,
					"isPartial":  1,
					"createdAt":  1,
					"status":     "$judgment.status",
					"score":      "$judgment.score",
					"totalScore": "$judgment.totalScore",
					"reason":     "$judgment.reason",
					"staffNote":  "$judgment.staffNote",
					"gradingItems": bson.M{
						"$map": bson.M{
							"input": "$judgment.gradingItems",
							"as":    "gi",
							"in": bson.M{
								"id":         "$$gi.id",
								"score":      "$$gi.score",
								"totalScore": "$$gi.totalScore",
								"status":     "$$gi.status",
								"testcases": bson.M{
									"$map": bson.M{
										"input": bson.M{
											"$filter": bson.M{
												"input": "$$gi.testcases",
												"as":    "tc",
												"cond": bson.M{
													"$in": []interface{}{"$$tc.id", sampleTestcaseIDs},
												},
											},
										},
										"as": "tc",
										"in": bson.M{
											"id":       "$$tc.id",
											"execTime": "$$tc.result.execTime",
											"exitCode": "$$tc.result.exitCode",
											"stdout":   "$$tc.result.stdout",
											"stderr":   "$$tc.result.stderr",
											"status":   "$$tc.status",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			{
				Key:   "$limit",
				Value: 1,
			},
		},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		q.logger.Errorf("Failed to aggregate submissions", "userId", userID, "submissionId", submissionID, "error", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var result *SubmissionDetail
	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			q.logger.Errorf("Failed to decode submission", "userId", userID, "submissionId", submissionID, "error", err)
			return nil, err
		}
	} else {
		q.logger.Debugf("Submission not found", "userId", userID, "submissionId", submissionID)
		return nil, httperror.New(http.StatusNotFound, "submission not found")
	}

	gradingItems := make([]dto.SubmissionGradingItem, 0, len(result.GradingItems))
	for _, gi := range result.GradingItems {
		testcases := make([]dto.SubmissionTestcase, 0, len(gi.Testcases))
		for _, tc := range gi.Testcases {
			testcases = append(testcases, dto.SubmissionTestcase{
				ID:       tc.ID,
				ExecTime: tc.ExecTime,
				ExitCode: tc.ExitCode,
				Stdout:   tc.Stdout,
				Stderr:   tc.Stderr,
				Status:   tc.Status,
			})
		}

		gradingItems = append(gradingItems, dto.SubmissionGradingItem{
			ID:        gi.ID,
			Score:     gi.Score,
			Status:    gi.Status,
			Testcases: testcases,
		})
	}

	q.logger.Debugf("Successfully got user submission", "userId", userID, "submissionId", submissionID)
	return &dto.SubmissionDetail{
		Submission: dto.Submission{
			ID:         result.ID,
			Status:     result.Status,
			IsPartial:  result.IsPartial,
			Score:      result.Score,
			TotalScore: result.TotalScore,
			CreatedAt:  result.CreatedAt,
			Reason:     result.Reason,
			StaffNote:  result.StaffNote,
		},
		GradingItems:      gradingItems,
		SampleTestcaseIDs: sampleTestcaseIDs,
		ContestProjectID:  contestProjectID,
	}, nil
}

// GetSampleTestcaseIDs is the interface's thin door to
// sampleTestcaseIDsForContestProject: the stream re-resolves the sample set
// through it while open, so a testcase flipped to hidden mid-contest stops
// streaming its output without waiting for a reconnect.
func (q *SubmissionQuery) GetSampleTestcaseIDs(contestProjectID int32) ([]int32, error) {
	return q.sampleTestcaseIDsForContestProject(contestProjectID)
}
