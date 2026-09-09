package query

import (
	"context"
	"errors"
	"judge-backend/internal/app/contest_domain/activity/application/query"
	"judge-backend/internal/app/contest_domain/activity/application/query/dto"
	"judge-backend/internal/app/shared/model"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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

type JudgmentInfo struct {
	ID         string  `bson:"_id"`
	Score      float64 `bson:"score"`
	TotalScore float64 `bson:"totalScore"`
	JudgeType  string  `bson:"judgeType"`
	Status     string  `bson:"status"`
}

type SubmissionInfo struct {
	ID               string       `bson:"_id"`
	UserID           int32        `bson:"userID"`
	ContestProjectID int32        `bson:"contestProjectID"`
	CreatedAt        time.Time    `bson:"createdAt"`
	IsPartial        bool         `bson:"isPartial"`
	Ignored          bool         `bson:"ignored"`
	Judgment         JudgmentInfo `bson:"judgment"`
}

func (q *SubmissionQuery) GetSubmissionsByContestProjectID(contestProjectID int32) ([]*dto.Submission, error) {
	cp := &model.ContestProject{}
	if err := q.db.Preload("Contest").Preload("Project").First(cp, contestProjectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, httperror.New(http.StatusNotFound, "contest project not found")
		}
		return nil, err
	}

	ctx := context.Background()
	collection := q.mdb.Database("judge").Collection("submissions")
	pipeline := mongo.Pipeline{
		{
			{
				Key: "$match",
				Value: bson.M{
					"contestProjectID": contestProjectID,
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
					"_id":                 1,
					"userID":              1,
					"contestProjectID":    1,
					"createdAt":           1,
					"isPartial":           1,
					"ignored":             1,
					"judgment._id":        1,
					"judgment.score":      1,
					"judgment.totalScore": 1,
					"judgment.judgeType":  1,
					"judgment.status":     1,
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

	if len(result) == 0 {
		return []*dto.Submission{}, nil
	}

	users := []*model.User{}
	if err := q.db.Find(&users).Error; err != nil {
		return nil, err
	}

	userMap := make(map[int32]*model.User)
	for _, v := range users {
		userMap[int32(v.ID)] = v
	}

	submissions := make([]*dto.Submission, 0)
	for _, v := range result {
		user := userMap[v.UserID]
		if user == nil {
			q.logger.Warnf(
				"User not found for submission %s (user %d, contest project %d), skipping",
				v.ID,
				v.UserID,
				v.ContestProjectID,
			)
			continue
		}
		submissions = append(submissions, &dto.Submission{
			ID:         v.ID,
			Contest:    &dto.ContestInfo{ID: cp.Contest.ID, Name: cp.Contest.Name},
			Project:    &dto.ProjectInfo{ID: cp.Project.ID, Name: cp.Project.Name},
			User:       &dto.User{ID: int32(user.ID), Username: user.Username, Name: user.Name},
			IsPartial:  v.IsPartial,
			Ignored:    v.Ignored,
			Status:     v.Judgment.Status,
			Type:       v.Judgment.JudgeType,
			Score:      v.Judgment.Score,
			TotalScore: v.Judgment.TotalScore,
			Time:       v.CreatedAt,
		})
	}

	return submissions, nil
}

type Submission struct {
	ID               string          `bson:"_id"`
	UserID           int32           `bson:"userID"`
	ContestProjectID int32           `bson:"contestProjectID"`
	Judgments        []*dto.Judgment `bson:"judgments"`
	IsPartial        bool            `bson:"isPartial"`
	Ignored          bool            `bson:"ignored"`
	GradingItemIds   []int32         `bson:"gradingItemIds"`
	CreatedAt        time.Time       `bson:"createdAt"`
}

func (q *SubmissionQuery) GetSubmissionInfoByID(submissionID string) (*dto.Submission, error) {
	ctx := context.Background()
	collection := q.mdb.Database("judge").Collection("submissions")
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"_id": submissionID}}},
		{{Key: "$lookup", Value: bson.M{
			"from": "judgments",
			"as":   "judgments",
			"let":  bson.M{"submissionID": "$_id"},
			"pipeline": mongo.Pipeline{
				{{Key: "$match", Value: bson.M{"$expr": bson.M{
					"$and": []interface{}{
						bson.M{"$eq": []interface{}{"$submissionID", "$$submissionID"}},
						bson.M{"$or": []interface{}{
							bson.M{"$ne": []interface{}{"$judgeType", "REJUDGE"}},
							bson.M{"$eq": []interface{}{"$status", "Judged"}},
						}},
					},
				}}}},
				{{Key: "$sort", Value: bson.M{"createdAt": -1}}},
				{{Key: "$limit", Value: 1}},
			},
		}}},
		{{Key: "$set", Value: bson.M{"judgment": bson.M{"$arrayElemAt": []interface{}{"$judgments", 0}}}}},
		{{Key: "$match", Value: bson.M{"judgment._id": bson.M{"$exists": true}}}},
		{{Key: "$project", Value: bson.M{
			"_id":                 1,
			"userID":              1,
			"contestProjectID":    1,
			"createdAt":           1,
			"isPartial":           1,
			"ignored":             1,
			"judgment._id":        1,
			"judgment.score":      1,
			"judgment.totalScore": 1,
			"judgment.judgeType":  1,
			"judgment.status":     1,
		}}},
		{{Key: "$limit", Value: 1}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rows []SubmissionInfo
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, httperror.New(http.StatusNotFound, "submission not found")
	}
	row := rows[0]

	user := &model.User{}
	if err := q.db.First(user, row.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			q.logger.Warnf(
				"User not found for submission %s (user %d, contest project %d)",
				row.ID,
				row.UserID,
				row.ContestProjectID,
			)
			return nil, httperror.New(http.StatusNotFound, "submission not found")
		}
		return nil, err
	}

	cp := &model.ContestProject{}
	if err := q.db.Preload("Contest").Preload("Project").First(cp, row.ContestProjectID).Error; err != nil {
		return nil, err
	}

	return &dto.Submission{
		ID:         row.ID,
		Contest:    &dto.ContestInfo{ID: cp.Contest.ID, Name: cp.Contest.Name},
		Project:    &dto.ProjectInfo{ID: cp.Project.ID, Name: cp.Project.Name},
		User:       &dto.User{ID: int32(user.ID), Username: user.Username, Name: user.Name},
		IsPartial:  row.IsPartial,
		Ignored:    row.Ignored,
		Status:     row.Judgment.Status,
		Type:       row.Judgment.JudgeType,
		Score:      row.Judgment.Score,
		TotalScore: row.Judgment.TotalScore,
		Time:       row.CreatedAt,
	}, nil
}

func (q *SubmissionQuery) GetSubmissionByID(id string) (*dto.SubmissionDetail, error) {
	ctx := context.Background()
	collection := q.mdb.Database("judge").Collection("submissions")
	pipeline := mongo.Pipeline{
		{
			{
				Key: "$match",
				Value: bson.M{
					"_id": id,
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
									"$expr": bson.M{"$eq": []interface{}{"$submissionID", "$$submissionID"}},
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
					},
				},
			},
		},
		{
			{
				Key: "$match",
				Value: bson.M{
					"judgments.0": bson.M{"$exists": true},
				},
			},
		},
		{
			{
				Key: "$project",
				Value: bson.M{
					"_id":              1,
					"userID":           1,
					"contestProjectID": 1,
					"projectID":        1,
					"gradingItemIds":   1,
					"isPartial":        1,
					"ignored":          1,
					"judgments": bson.M{"$map": bson.M{
						"input": "$judgments",
						"as":    "judgment",
						"in": bson.M{
							"id":                "$$judgment._id",
							"gradingItems":      "$$judgment.gradingItems",
							"lifecycle":         "$$judgment.lifecycle",
							"project":           "$$judgment.project",
							"score":             "$$judgment.score",
							"totalScore":        "$$judgment.totalScore",
							"judgeType":         "$$judgment.judgeType",
							"rejudgeID":         "$$judgment.rejudgeID",
							"status":            "$$judgment.status",
							"createdAt":         "$$judgment.createdAt",
							"judgingStartedAt":  "$$judgment.judgingStartedAt",
							"judgingFinishedAt": "$$judgment.judgingFinishedAt",
							"errorMessage":      "$$judgment.errorMessage",
							"staffNote":         "$$judgment.staffNote",
						},
					}},
					"createdAt": 1,
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
		return nil, err
	}
	defer cursor.Close(ctx)

	var result *Submission

	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
	} else {
		return nil, httperror.New(http.StatusNotFound, "submission not found")
	}

	user := &model.User{}
	if err := q.db.First(user, result.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			q.logger.Warnf(
				"User not found for submission %s (user %d, contest project %d)",
				result.ID,
				result.UserID,
				result.ContestProjectID,
			)
			return nil, httperror.New(http.StatusNotFound, "submission not found")
		}
		return nil, err
	}

	cp := &model.ContestProject{}
	if err := q.db.Preload("Contest").Preload("Project").First(cp, result.ContestProjectID).Error; err != nil {
		return nil, err
	}

	gs := []*model.GradingItem{}
	if result.IsPartial {
		if err := q.db.Preload("Testcases").Where("id IN ?", result.GradingItemIds).Find(&gs).Error; err != nil {
			return nil, err
		}
	} else {
		if err := q.db.Preload("Testcases").Where("project_id = ?", cp.Project.ID).Find(&gs).Error; err != nil {
			return nil, err
		}
	}

	gradingItems := []*dto.GradingItemInfo{}
	for _, v := range gs {
		testcases := []*dto.TestcaseInfo{}
		for _, t := range v.Testcases {
			testcases = append(testcases, &dto.TestcaseInfo{
				ID:    t.ID,
				Name:  t.Name,
				Score: t.Score,
			})
		}
		gradingItems = append(gradingItems, &dto.GradingItemInfo{
			ID:        v.ID,
			Name:      v.Name,
			Score:     v.Score,
			Testcases: testcases,
		})
	}

	return &dto.SubmissionDetail{
		ID:           result.ID,
		Contest:      &dto.ContestInfo{ID: cp.Contest.ID, Name: cp.Contest.Name},
		Project:      &dto.ProjectInfo{ID: cp.Project.ID, Name: cp.Project.Name},
		User:         &dto.User{ID: int32(user.ID), Username: user.Username, Name: user.Name},
		Judgments:    result.Judgments,
		Ignored:      result.Ignored,
		IsPartial:    result.IsPartial,
		GradingItems: gradingItems,
		CreatedAt:    result.CreatedAt,
	}, nil
}

func (q *SubmissionQuery) GetSubmissionsByUserID(userID int32) ([]*dto.Submission, error) {
	ctx := context.Background()
	collection := q.mdb.Database("judge").Collection("submissions")
	pipeline := mongo.Pipeline{
		{
			{
				Key: "$match",
				Value: bson.M{
					"userID": userID,
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
					"_id":                 1,
					"userID":              1,
					"contestProjectID":    1,
					"createdAt":           1,
					"isPartial":           1,
					"ignored":             1,
					"judgment._id":        1,
					"judgment.score":      1,
					"judgment.totalScore": 1,
					"judgment.judgeType":  1,
					"judgment.status":     1,
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

	if len(result) == 0 {
		return []*dto.Submission{}, nil
	}

	users := []*model.User{}
	if err := q.db.Find(&users).Error; err != nil {
		return nil, err
	}

	userMap := make(map[int32]*model.User)
	for _, v := range users {
		userMap[int32(v.ID)] = v
	}

	contestProjects := []*model.ContestProject{}
	if err := q.db.Preload("Project").Preload("Contest").Find(&contestProjects).Error; err != nil {
		return nil, err
	}

	cpMap := make(map[int32]*model.ContestProject)
	for _, v := range contestProjects {
		cpMap[v.ID] = v
	}

	submissions := make([]*dto.Submission, 0)
	for _, v := range result {
		cp := cpMap[v.ContestProjectID]
		if cp == nil || cp.Contest.ID == 0 {
			q.logger.Warnf("ContestProject or Contest not found for submission, skipping", "submissionId", v.ID, "contestProjectId", v.ContestProjectID)
			continue
		}
		user := userMap[v.UserID]
		if user == nil {
			q.logger.Warnf("User not found for submission, skipping", "submissionId", v.ID, "userId", v.UserID)
			continue
		}
		submissions = append(submissions, &dto.Submission{
			ID:         v.ID,
			Contest:    &dto.ContestInfo{ID: cp.Contest.ID, Name: cp.Contest.Name},
			Project:    &dto.ProjectInfo{ID: cp.Project.ID, Name: cp.Project.Name},
			User:       &dto.User{ID: int32(user.ID), Username: user.Username, Name: user.Name},
			IsPartial:  v.IsPartial,
			Ignored:    v.Ignored,
			Status:     v.Judgment.Status,
			Type:       v.Judgment.JudgeType,
			Score:      v.Judgment.Score,
			TotalScore: v.Judgment.TotalScore,
			Time:       v.CreatedAt,
		})
	}

	return submissions, nil
}
