package query

import (
	"context"
	"judge-backend/internal/app/contest_domain/activity/application/query"
	"judge-backend/internal/app/contest_domain/activity/application/query/dto"
	"judge-backend/internal/app/shared/model"
	"judge-backend/internal/app/shared/rolebits"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

type ContestProjectQuery struct {
	db                *gorm.DB
	mdb               *mongo.Client
	contestantBitMask int
	logger            sharedtypes.Logger
}

type scoreExportSubmission struct {
	ID        string    `bson:"_id"`
	UserID    int32     `bson:"userID"`
	CreatedAt time.Time `bson:"createdAt"`
}

type scoreExportGradingItem struct {
	ID         int32   `bson:"id"`
	Score      float64 `bson:"score"`
	TotalScore float64 `bson:"totalScore"`
	Status     string  `bson:"status"`
}

type scoreExportJudgment struct {
	ID           string                    `bson:"_id"`
	SubmissionID string                    `bson:"submissionID"`
	Score        float64                   `bson:"score"`
	GradingItems []*scoreExportGradingItem `bson:"gradingItems"`
}

func NewContestProjectQuery(
	db *gorm.DB,
	mdb *mongo.Client,
	masks rolebits.RoleMasks,
	logger sharedtypes.Logger,
) query.ContestProjectQuery {
	return &ContestProjectQuery{
		db:                db,
		mdb:               mdb,
		contestantBitMask: int(masks.Contestant),
		logger:            logger,
	}
}

func (q *ContestProjectQuery) GetContestProjectByID(id int32) (*dto.ContestProject, error) {
	p := &model.ContestProject{}
	if err := q.db.Preload("Project").First(p, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, httperror.New(http.StatusNotFound, "contest project not found")
		}

		return nil, err
	}

	return &dto.ContestProject{
		ID:                        p.ID,
		ProjectID:                 p.ProjectID,
		Name:                      p.Project.Name,
		Image:                     p.Project.ImageTag,
		TestImage:                 p.Project.TestImageTag,
		Description:               p.Project.Description,
		IsPublic:                  p.IsPublic,
		JudgePriorityStrategy:     p.JudgePriorityStrategy,
		ActivateTime:              p.ActivateTime,
		DeactivateTime:            p.DeactivateTime,
		StartTime:                 p.StartTime,
		EndTime:                   p.EndTime,
		MaxJudgeDurationMinutes:   p.MaxJudgeDurationMinutes,
		SubmissionCooldownSeconds: p.SubmissionCooldownSeconds,
		PidsLimit:                 p.PidsLimit,
		MemoryLimitMB:             p.MemoryLimitMB,
		ScoreboardCompact:         p.ScoreboardCompact,
		ScoreboardMaskNames:       p.ScoreboardMaskNames,
	}, nil
}

func (q *ContestProjectQuery) GetUserScores(contestProjectID int32) ([]*dto.UserScore, error) {
	ctx := context.Background()
	submissionCollection := q.mdb.Database("judge").Collection("submissions")
	judgmentCollection := q.mdb.Database("judge").Collection("judgments")

	type Submission struct {
		ID        string    `bson:"_id"`
		CreatedAt time.Time `bson:"createdAt"`
		UserID    int32     `bson:"userID"`
	}
	var submissions []*Submission
	cursor, err := submissionCollection.Find(ctx, bson.M{
		"contestProjectID": contestProjectID,
		"isPartial":        false,
		"ignored":          false,
	})
	if err != nil {
		return nil, err
	}
	if err := cursor.All(ctx, &submissions); err != nil {
		return nil, err
	}

	submissionIds := make([]string, 0, len(submissions))
	for _, s := range submissions {
		submissionIds = append(submissionIds, s.ID)
	}

	type Judgment struct {
		SubmissionID string  `bson:"submissionID"`
		Score        float64 `bson:"score"`
		Submission   *Submission
	}
	var judgments []*Judgment
	cursor, err = judgmentCollection.Find(ctx, bson.M{
		"submissionID": bson.M{
			"$in": submissionIds,
		},
		"status": "Judged",
	}, options.Find().SetSort(bson.M{"createdAt": -1}))
	if err != nil {
		return nil, err
	}
	if err := cursor.All(ctx, &judgments); err != nil {
		return nil, err
	}

	submissionMap := make(map[string]*Submission)
	for _, s := range submissions {
		submissionMap[s.ID] = s
	}

	for i, j := range judgments {
		s := submissionMap[j.SubmissionID]
		if s == nil {
			q.logger.Warnf("submission not found", "submissionId", j.SubmissionID)
			continue
		}

		judgments[i].Submission = s
	}

	highestScoreJudgment := make(map[int32]*Judgment)
	for _, j := range judgments {
		if j.Submission == nil {
			q.logger.Warnf("submission not found", "submissionId", j.SubmissionID)
			continue
		}

		if _, ok := highestScoreJudgment[j.Submission.UserID]; !ok {
			highestScoreJudgment[j.Submission.UserID] = j
			continue
		}

		if j.Score > highestScoreJudgment[j.Submission.UserID].Score {
			highestScoreJudgment[j.Submission.UserID] = j
		}
	}

	userJudgments := make([]*Judgment, 0, len(highestScoreJudgment))
	for _, j := range highestScoreJudgment {
		userJudgments = append(userJudgments, j)
	}

	sort.SliceStable(userJudgments, func(i, j int) bool {
		return userJudgments[i].Score > userJudgments[j].Score
	})

	users := []*model.User{}
	if err := q.db.Where("roles & ? != 0", q.contestantBitMask).Find(&users).Error; err != nil {
		return nil, err
	}

	userMap := make(map[int]*model.User)
	for _, u := range users {
		userMap[u.ID] = u
	}

	visitedUsers := make(map[int]struct{})

	result := make([]*dto.UserScore, 0, len(userJudgments))
	for _, j := range userJudgments {
		user := userMap[int(j.Submission.UserID)]
		if user == nil {
			continue
		}

		visitedUsers[int(j.Submission.UserID)] = struct{}{}
		result = append(result, &dto.UserScore{
			User: &dto.User{
				ID:       int32(user.ID),
				Username: user.Username,
				Name:     user.Name,
			},
			Score:        j.Score,
			SubmissionID: j.Submission.ID,
		})
	}

	for _, u := range users {
		if _, ok := visitedUsers[u.ID]; ok {
			continue
		}

		result = append(result, &dto.UserScore{
			User: &dto.User{
				ID:       int32(u.ID),
				Username: u.Username,
				Name:     u.Name,
			},
			Score:        0,
			SubmissionID: "",
		})
	}

	return result, nil
}

func (q *ContestProjectQuery) GetContestantScoreDetails(contestProjectID int32) ([]*dto.ContestantScoreDetail, error) {
	cp := &model.ContestProject{}
	if err := q.db.First(cp, contestProjectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, httperror.New(http.StatusNotFound, "contest project not found")
		}
		return nil, err
	}

	// Fetch grading items from MySQL sorted by label
	gradingItems := []*model.GradingItem{}
	if err := q.db.Where("project_id = ?", cp.ProjectID).Order("`order` ASC").Find(&gradingItems).Error; err != nil {
		return nil, err
	}

	// grading item ID → position in label order
	gradingItemOrder := make([]int32, 0, len(gradingItems))
	for _, gi := range gradingItems {
		gradingItemOrder = append(gradingItemOrder, gi.ID)
	}

	// Fetch contestant users only
	users := []*model.User{}
	if err := q.db.Where("roles & ? != 0", q.contestantBitMask).Find(&users).Error; err != nil {
		return nil, err
	}
	userMap := make(map[int32]*model.User)
	for _, u := range users {
		userMap[int32(u.ID)] = u
	}

	// Fetch submissions from MongoDB
	ctx := context.Background()
	submissionCollection := q.mdb.Database("judge").Collection("submissions")
	judgmentCollection := q.mdb.Database("judge").Collection("judgments")

	var submissions []*scoreExportSubmission
	cursor, err := submissionCollection.Find(ctx, bson.M{
		"contestProjectID": contestProjectID,
		"isPartial":        false,
		"ignored":          false,
	})
	if err != nil {
		return nil, err
	}
	if err := cursor.All(ctx, &submissions); err != nil {
		return nil, err
	}

	submissionIDs := make([]string, 0, len(submissions))
	submissionUserMap := make(map[string]int32)
	for _, s := range submissions {
		submissionIDs = append(submissionIDs, s.ID)
		submissionUserMap[s.ID] = s.UserID
	}

	var judgments []*scoreExportJudgment
	cursor, err = judgmentCollection.Find(ctx, bson.M{
		"submissionID": bson.M{"$in": submissionIDs},
		"status":       "Judged",
	})
	if err != nil {
		return nil, err
	}
	if err := cursor.All(ctx, &judgments); err != nil {
		return nil, err
	}

	// Find best judgment per user using tiebreaker
	bestJudgmentPerUser := make(map[int32]*scoreExportJudgment)
	for _, j := range judgments {
		userID := submissionUserMap[j.SubmissionID]
		current, exists := bestJudgmentPerUser[userID]
		if !exists {
			bestJudgmentPerUser[userID] = j
			continue
		}
		if isBetterJudgment(j, current, gradingItemOrder) {
			bestJudgmentPerUser[userID] = j
		}
	}

	// Build grading item lookup from MySQL
	giMap := make(map[int32]*model.GradingItem)
	for _, gi := range gradingItems {
		giMap[gi.ID] = gi
	}

	// Build result: one entry per user
	visitedUsers := make(map[int32]struct{})
	result := make([]*dto.ContestantScoreDetail, 0, len(users))

	for userID, j := range bestJudgmentPerUser {
		user := userMap[userID]
		if user == nil {
			continue
		}
		visitedUsers[userID] = struct{}{}

		// Build grading item score map from judgment
		giScoreMap := make(map[int32]*scoreExportGradingItem)
		for _, gi := range j.GradingItems {
			giScoreMap[gi.ID] = gi
		}

		items := make([]*dto.ScoreDetailItem, 0, len(gradingItems))
		for _, gi := range gradingItems {
			item := &dto.ScoreDetailItem{
				GradingItemID: gi.ID,
				Label:         gi.Label,
				Name:          gi.Name,
				Description:   gi.Description,
				TotalScore:    gi.Score,
				Score:         0,
				Status:        "-",
			}
			if mongoGI, ok := giScoreMap[gi.ID]; ok {
				item.Score = mongoGI.Score
				item.Status = mongoGI.Status
			}
			items = append(items, item)
		}

		result = append(result, &dto.ContestantScoreDetail{
			User: &dto.User{
				ID:       int32(user.ID),
				Name:     user.Name,
				Username: user.Username,
			},
			Items: items,
		})
	}

	// Include users with no submissions
	for _, u := range users {
		if _, ok := visitedUsers[int32(u.ID)]; ok {
			continue
		}
		items := make([]*dto.ScoreDetailItem, 0, len(gradingItems))
		for _, gi := range gradingItems {
			items = append(items, &dto.ScoreDetailItem{
				GradingItemID: gi.ID,
				Label:         gi.Label,
				Name:          gi.Name,
				Description:   gi.Description,
				TotalScore:    gi.Score,
				Score:         0,
				Status:        "-",
			})
		}
		result = append(result, &dto.ContestantScoreDetail{
			User: &dto.User{
				ID:       int32(u.ID),
				Name:     u.Name,
				Username: u.Username,
			},
			Items: items,
		})
	}

	return result, nil
}

type analyticsSubmission struct {
	ID        string    `bson:"_id"`
	CreatedAt time.Time `bson:"createdAt"`
}

type analyticsJudgment struct {
	SubmissionID      string     `bson:"submissionID"`
	JudgeType         string     `bson:"judgeType"`
	Status            string     `bson:"status"`
	CreatedAt         time.Time  `bson:"createdAt"`
	JudgingStartedAt  *time.Time `bson:"judgingStartedAt"`
	JudgingFinishedAt *time.Time `bson:"judgingFinishedAt"`
}

func buildJudgmentAnalytics(submissions []*analyticsSubmission, judgments []*analyticsJudgment) *dto.JudgmentAnalytics {
	submissionCreatedAtMap := make(map[string]time.Time, len(submissions))
	for _, s := range submissions {
		submissionCreatedAtMap[s.ID] = s.CreatedAt
	}

	statusMap := make(map[string]int32)
	var totalJudgingMs, totalQueueMs float64
	var completedCount int

	for _, j := range judgments {
		statusMap[j.Status]++

		if j.JudgingStartedAt == nil || j.JudgingFinishedAt == nil {
			continue
		}
		if j.Status != "Judged" && j.Status != "Rejudged" {
			continue
		}

		totalJudgingMs += float64(j.JudgingFinishedAt.Sub(*j.JudgingStartedAt).Milliseconds())

		queueStart := j.CreatedAt
		if j.JudgeType != "REJUDGE" {
			submissionCreatedAt, ok := submissionCreatedAtMap[j.SubmissionID]
			if !ok {
				continue
			}
			queueStart = submissionCreatedAt
		}

		totalQueueMs += float64(j.JudgingStartedAt.Sub(queueStart).Milliseconds())
		completedCount++
	}

	statusCounts := make([]*dto.JudgmentStatusCount, 0, len(statusMap))
	for status, count := range statusMap {
		statusCounts = append(statusCounts, &dto.JudgmentStatusCount{
			Status: status,
			Count:  count,
		})
	}
	sort.Slice(statusCounts, func(i, j int) bool {
		return statusCounts[i].Status < statusCounts[j].Status
	})

	var avgJudgingMs, avgQueueMs *float64
	if completedCount > 0 {
		judgingAvg := totalJudgingMs / float64(completedCount)
		queueAvg := totalQueueMs / float64(completedCount)
		avgJudgingMs = &judgingAvg
		avgQueueMs = &queueAvg
	}

	return &dto.JudgmentAnalytics{
		TotalSubmissions:     int32(len(submissions)),
		TotalJudgments:       int32(len(judgments)),
		StatusCounts:         statusCounts,
		AvgJudgingDurationMs: avgJudgingMs,
		AvgQueueWaitMs:       avgQueueMs,
	}
}

func (q *ContestProjectQuery) GetJudgmentAnalytics(contestProjectID int32) (*dto.JudgmentAnalytics, error) {
	ctx := context.Background()
	submissionCollection := q.mdb.Database("judge").Collection("submissions")
	judgmentCollection := q.mdb.Database("judge").Collection("judgments")

	var submissions []*analyticsSubmission
	cursor, err := submissionCollection.Find(ctx, bson.M{
		"contestProjectID": contestProjectID,
		"isPartial":        false,
		"ignored":          false,
	})
	if err != nil {
		return nil, err
	}
	if err := cursor.All(ctx, &submissions); err != nil {
		return nil, err
	}

	submissionIDs := make([]string, 0, len(submissions))
	for _, s := range submissions {
		submissionIDs = append(submissionIDs, s.ID)
	}
	if len(submissionIDs) == 0 {
		return &dto.JudgmentAnalytics{
			TotalSubmissions: 0,
			TotalJudgments:   0,
			StatusCounts:     []*dto.JudgmentStatusCount{},
		}, nil
	}

	var judgments []*analyticsJudgment
	cursor, err = judgmentCollection.Find(ctx, bson.M{
		"submissionID": bson.M{"$in": submissionIDs},
	})
	if err != nil {
		return nil, err
	}
	if err := cursor.All(ctx, &judgments); err != nil {
		return nil, err
	}

	return buildJudgmentAnalytics(submissions, judgments), nil
}

// isBetterJudgment returns true if candidate is better than current.
// Higher score wins. On tie, compare grading item statuses in label order:
// at the first differing position, "Passed" wins.
func isBetterJudgment(candidate, current *scoreExportJudgment, gradingItemOrder []int32) bool {
	if candidate.Score > current.Score {
		return true
	}
	if candidate.Score < current.Score {
		return false
	}

	// Tiebreaker: compare per grading item in label order
	candidateMap := make(map[int32]string)
	for _, gi := range candidate.GradingItems {
		candidateMap[gi.ID] = gi.Status
	}
	currentMap := make(map[int32]string)
	for _, gi := range current.GradingItems {
		currentMap[gi.ID] = gi.Status
	}

	for _, id := range gradingItemOrder {
		cs := candidateMap[id]
		cur := currentMap[id]
		if cs == cur {
			continue
		}
		return cs == "Passed"
	}
	return false
}
