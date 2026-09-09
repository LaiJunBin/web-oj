package query

import (
	"context"
	"fmt"
	"judge-backend/internal/app/contest_domain/contest/application/query"
	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	"judge-backend/internal/app/contest_domain/contest/domain/common"
	"judge-backend/internal/app/contest_domain/contest/domain/provider"
	"judge-backend/internal/app/shared/model"
	"judge-backend/internal/app/shared/rolebits"
	sharedtypes "judge-backend/internal/app/shared/types"
	"os"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

type ProjectQuery struct {
	db                *gorm.DB
	mdb               *mongo.Client
	storagePath       string
	contestantBitMask int32
	logger            sharedtypes.Logger
}

func NewProjectQuery(
	db *gorm.DB,
	mdb *mongo.Client,
	storagePathProvider provider.StoragePathProvider,
	masks rolebits.RoleMasks,
	logger sharedtypes.Logger,
) query.ProjectQuery {
	return &ProjectQuery{
		db:                db,
		mdb:               mdb,
		storagePath:       storagePathProvider.GetStoragePath(),
		contestantBitMask: int32(masks.Contestant),
		logger:            logger,
	}
}

func (pq *ProjectQuery) transformContestProjectToProjectDetail(cp *model.ContestProject) (*dto.ProjectDetail, error) {
	pq.logger.Debugf("Starting to transform contest project to project detail", "contestProjectId", cp.ID)

	if !cp.IsActivated() {
		pq.logger.Debugf("Contest project is not activated", "contestProjectId", cp.ID)
		return nil, query.ErrorContestProjectNotFound
	}

	project := &model.Project{}
	if err := pq.db.Preload("DynamicFields").Preload("GradingItems", func(tx *gorm.DB) *gorm.DB {
		return tx.Preload("Testcases")
	}).Preload("Assets").First(project, cp.ProjectID).Error; err != nil {
		pq.logger.Errorf("Failed to find project", "projectId", cp.ProjectID, "error", err)
		return nil, err
	}

	dynamicFields := make([]dto.DynamicField, 0, len(project.DynamicFields))
	for _, df := range project.DynamicFields {
		dynamicFields = append(dynamicFields, dto.DynamicField{
			Type:         df.Type,
			Label:        df.Label,
			Name:         df.Name,
			DefaultValue: df.DefaultValue,
			IsRequired:   df.IsRequired,
		})
	}

	assets := make([]dto.AssetInfo, 0, len(project.Assets))
	for _, asset := range project.Assets {
		if !asset.IsEnabled {
			continue
		}

		assets = append(assets, dto.AssetInfo{
			ID:   asset.ID,
			Name: asset.Name,
		})
	}

	gradingItems := make([]dto.GradingItem, 0, len(project.GradingItems))
	for _, gi := range project.GradingItems {
		testcases := make([]dto.Testcase, 0, len(gi.Testcases))
		for _, tc := range gi.Testcases {
			if !tc.IsSample {
				continue
			}
			testcases = append(testcases, dto.Testcase{
				ID:   tc.ID,
				Name: tc.Name,
			})
		}

		gradingItems = append(gradingItems, dto.GradingItem{
			ID:              gi.ID,
			Name:            gi.Name,
			Label:           gi.Label,
			Description:     gi.Description,
			Score:           gi.Score,
			OutputFormat:    gi.OutputFormat,
			PublicTestcases: testcases,
		})
	}

	pq.logger.Debugf("Successfully transformed contest project to project detail", "contestProjectId", cp.ID)
	return &dto.ProjectDetail{
		Project: dto.Project{
			ID:     cp.ID,
			Name:   cp.Name,
			Status: cp.Status(),
		},
		Description:   project.Description,
		DynamicFields: dynamicFields,
		GradingItems:  gradingItems,
		Assets:        assets,
		TimeLeft:      int64(cp.TimeLeft().Seconds()),
	}, nil
}

func (pq *ProjectQuery) GetPublicProjectByID(contestProjectID int32) (dto.ProjectDetail, error) {
	pq.logger.Debugf("Starting to get public project by ID", "contestProjectId", contestProjectID)

	cp := &model.ContestProject{}
	if err := pq.db.Preload("Contest").Preload("Project").First(cp, "id = ? and is_public = ?", contestProjectID, true).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			pq.logger.Debugf("Public contest project not found", "contestProjectId", contestProjectID)
			return dto.ProjectDetail{}, query.ErrorContestProjectNotFound
		}

		pq.logger.Errorf("Failed to find public contest project", "contestProjectId", contestProjectID, "error", err)
		return dto.ProjectDetail{}, err
	}

	if !cp.IsActivated() {
		pq.logger.Debugf("Public contest project is not activated", "contestProjectId", contestProjectID)
		return dto.ProjectDetail{}, query.ErrorContestProjectNotFound
	}

	if !cp.IsStarted() {
		pq.logger.Debugf("Public contest project has not started", "contestProjectId", contestProjectID)
		return dto.ProjectDetail{
			Project: dto.Project{
				ID:     cp.ID,
				Name:   cp.Name,
				Status: cp.Status(),
			},
			Description:   cp.Project.Description,
			DynamicFields: []dto.DynamicField{},
			GradingItems:  []dto.GradingItem{},
			Assets:        []dto.AssetInfo{},
			TimeLeft:      int64(cp.TimeLeft().Seconds()),
		}, nil
	}

	result, err := pq.transformContestProjectToProjectDetail(cp)
	if err != nil {
		pq.logger.Errorf("Failed to transform public contest project", "contestProjectId", contestProjectID, "error", err)
		return dto.ProjectDetail{}, err
	}

	pq.logger.Debugf("Successfully got public project", "contestProjectId", contestProjectID)
	return *result, nil
}

func (pq *ProjectQuery) GetUserProjectByID(userID, contestProjectID int32) (dto.UserProjectDetail, error) {
	pq.logger.Debugf("Starting to get user project by ID", "userId", userID, "contestProjectId", contestProjectID)

	cp := &model.ContestProject{}
	if err := pq.db.Preload("Contest").Preload("Project").First(cp, "id = ?", contestProjectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			pq.logger.Debugf("Contest project not found", "contestProjectId", contestProjectID)
			return dto.UserProjectDetail{}, query.ErrorContestProjectNotFound
		}

		pq.logger.Errorf("Failed to find contest project", "contestProjectId", contestProjectID, "error", err)
		return dto.UserProjectDetail{}, err
	}

	if !cp.IsActivated() {
		pq.logger.Debugf("Contest project is not activated", "contestProjectId", contestProjectID)
		return dto.UserProjectDetail{}, query.ErrorContestProjectNotFound
	}

	if err := pq.checkUserCanAccessContest(&cp.Contest, userID); err != nil {
		pq.logger.Debugf("User cannot access contest project", "userId", userID, "contestProjectId", contestProjectID)
		return dto.UserProjectDetail{}, err
	}

	isParticipant, err := isContestParticipant(pq.db, cp.ContestID, userID)
	if err != nil {
		pq.logger.Errorf("Failed to check user is participant", "userId", userID, "contestProjectId", contestProjectID, "error", err)
		return dto.UserProjectDetail{}, err
	}

	if !cp.IsStarted() {
		pq.logger.Debugf("Contest project has not started", "contestProjectId", contestProjectID)
		return dto.UserProjectDetail{
			ProjectDetail: dto.ProjectDetail{
				Project: dto.Project{
					ID:     cp.ID,
					Name:   cp.Name,
					Status: cp.Status(),
				},
				Description:   cp.Project.Description,
				DynamicFields: []dto.DynamicField{},
				GradingItems:  []dto.GradingItem{},
				Assets:        []dto.AssetInfo{},
				TimeLeft:      int64(cp.TimeLeft().Seconds()),
			},
			IsParticipant: isParticipant,
		}, nil
	}

	result, err := pq.transformContestProjectToProjectDetail(cp)
	if err != nil {
		pq.logger.Errorf("Failed to transform contest project", "contestProjectId", contestProjectID, "error", err)
		return dto.UserProjectDetail{}, err
	}

	pq.logger.Debugf("Successfully got user project", "userId", userID, "contestProjectId", contestProjectID)
	return dto.UserProjectDetail{
		ProjectDetail: *result,
		IsParticipant: isParticipant,
	}, nil
}

func (pq *ProjectQuery) checkUserCanAccessContest(contest *model.Contest, userID int32) error {
	if contest.IsPublic {
		return nil
	}

	ok, err := isContestParticipant(pq.db, contest.ID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return errNotParticipant
	}
	return nil
}

func (pq *ProjectQuery) GetUserProjectAssetFile(userID, contestProjectID int32, assetID string) (dto.AssetFile, error) {
	pq.logger.Debugf("Starting to get user project asset file", "userId", userID, "contestProjectId", contestProjectID, "assetId", assetID)

	cp := &model.ContestProject{}
	if err := pq.db.Preload("Contest").First(cp, "id = ?", contestProjectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			pq.logger.Debugf("Contest project not found", "contestProjectId", contestProjectID)
			return dto.AssetFile{}, query.ErrorContestProjectNotFound
		}

		pq.logger.Errorf("Failed to find contest project", "contestProjectId", contestProjectID, "error", err)
		return dto.AssetFile{}, err
	}

	if !cp.IsActivated() {
		pq.logger.Debugf("Contest project is not activated", "contestProjectId", contestProjectID)
		return dto.AssetFile{}, query.ErrorContestProjectNotFound
	}

	if err := pq.checkUserCanAccessContest(&cp.Contest, userID); err != nil {
		pq.logger.Debugf("User cannot access contest project asset", "userId", userID, "contestProjectId", contestProjectID)
		return dto.AssetFile{}, err
	}

	asset := &model.Asset{}
	if err := pq.db.First(asset, "id = ?", assetID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			pq.logger.Debugf("Asset not found", "assetId", assetID)
			return dto.AssetFile{}, query.ErrorAssetNotFound
		}

		pq.logger.Errorf("Failed to find asset", "assetId", assetID, "error", err)
		return dto.AssetFile{}, err
	}

	if !asset.IsEnabled {
		pq.logger.Debugf("Asset is not enabled", "assetId", assetID)
		return dto.AssetFile{}, query.ErrorAssetNotFound
	}

	assetPath := fmt.Sprintf("%s/assets/%s", pq.storagePath, assetID)

	bytes, err := os.ReadFile(assetPath)
	if err != nil {
		pq.logger.Errorf("Failed to read asset file", "assetPath", assetPath, "error", err)
		return dto.AssetFile{}, err
	}

	pq.logger.Debugf("Successfully got user project asset file", "userId", userID, "contestProjectId", contestProjectID, "assetId", assetID)
	return dto.AssetFile{
		Name:    asset.Name,
		Content: bytes,
		Size:    int64(len(bytes)),
	}, nil
}

type ScoreboardItem struct {
	UserID       int32 `json:"userID"`
	HighestScore int32 `json:"highestScore"`
	Statuses     []struct {
		ID     int32  `json:"id"`
		Status string `json:"status"`
	} `json:"statuses"`
}

func (pq *ProjectQuery) GetProjectScoreboardByID(contestProjectID int32) (*dto.Scoreboard, error) {
	pq.logger.Debugf("Starting to get project scoreboard by ID", "contestProjectId", contestProjectID)

	cp := &model.ContestProject{}
	if err := pq.db.Preload("Contest").Preload("Project", func(tx *gorm.DB) *gorm.DB {
		return tx.Preload("GradingItems")
	}).First(cp, contestProjectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			pq.logger.Debugf("Contest project not found", "contestProjectId", contestProjectID)
			return nil, query.ErrorContestProjectNotFound
		}
		pq.logger.Errorf("Failed to find contest project", "contestProjectId", contestProjectID, "error", err)
		return nil, err
	}

	if !cp.IsActivated() {
		pq.logger.Debugf("Contest project is not activated", "contestProjectId", contestProjectID)
		return nil, query.ErrorContestProjectNotFound
	}

	ctx := context.Background()
	submissionCollection := pq.mdb.Database("judge").Collection("submissions")
	judgmentCollection := pq.mdb.Database("judge").Collection("judgments")

	type Submission struct {
		ID               string    `bson:"_id"`
		ContestProjectID int32     `bson:"contestProjectID"`
		CreatedAt        time.Time `bson:"createdAt"`
		GradingItemIds   []int32   `bson:"gradingItemIds"`
		IsPartial        bool      `bson:"isPartial"`
		UserID           int32     `bson:"userID"`
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
		ID           string  `bson:"_id"`
		SubmissionID string  `bson:"submissionID"`
		JudgeType    string  `bson:"judgeType"`
		Score        float64 `bson:"score"`
		Status       string  `bson:"status"`
		GradingItems []struct {
			ID     int32  `bson:"id"`
			Status string `bson:"status"`
		} `bson:"gradingItems"`
		CreatedAt time.Time `bson:"createdAt"`

		Submission *Submission
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
			pq.logger.Warnf("submission not found", "submissionId", j.SubmissionID)
			continue
		}

		judgments[i].Submission = s
	}

	highestScoreJudgment := make(map[int32]*Judgment)
	for _, j := range judgments {
		if j.Submission == nil {
			pq.logger.Warnf("submission not found", "submissionId", j.SubmissionID)
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
		return userJudgments[i].Score > userJudgments[j].Score || (userJudgments[i].Score == userJudgments[j].Score && userJudgments[i].CreatedAt.Before(userJudgments[j].CreatedAt))
	})

	result := make([]*ScoreboardItem, 0, len(userJudgments))
	for _, j := range userJudgments {
		statuses := make([]struct {
			ID     int32  `json:"id"`
			Status string `json:"status"`
		}, 0, len(j.GradingItems))
		for _, gi := range j.GradingItems {
			statuses = append(statuses, struct {
				ID     int32  `json:"id"`
				Status string `json:"status"`
			}{
				ID:     gi.ID,
				Status: gi.Status,
			})
		}
		result = append(result, &ScoreboardItem{
			UserID:       j.Submission.UserID,
			HighestScore: int32(j.Score),
			Statuses:     statuses,
		})
	}

	pq.logger.Debugf("Successfully got project scoreboard", "contestProjectId", contestProjectID)
	return pq.toDomainModel(cp, result)
}

func (pq *ProjectQuery) getContestParticipantUserIDs(contestID int32) ([]int32, error) {
	var userIDs []int32
	err := pq.db.Raw(`
		SELECT DISTINCT u.id FROM users u
		WHERE EXISTS (
			SELECT 1 FROM contest_participants cp
			WHERE cp.contest_id = ?
			  AND (
			    cp.user_id = u.id
			    OR EXISTS (
			      SELECT 1 FROM user_groups ug
			      INNER JOIN `+"`groups`"+` g ON g.id = ug.group_id AND g.deleted_at IS NULL
			      WHERE ug.group_id = cp.group_id AND ug.user_id = u.id
			    )
			  )
		)
	`, contestID).Scan(&userIDs).Error
	return userIDs, err
}

func (pq *ProjectQuery) toDomainModel(cp *model.ContestProject, result []*ScoreboardItem) (*dto.Scoreboard, error) {
	scoreboard := &dto.Scoreboard{
		Heads:   make([]*dto.ScoreboardHead, 0, len(cp.Project.GradingItems)),
		Rows:    make([]*dto.ScoreboardRow, 0, len(result)),
		Compact: cp.ScoreboardCompact,
	}

	// Masking happens here — in the query that serves contestant/public
	// scoreboards (GET and SSE alike) — so unmasked names never leave the
	// backend for those audiences. Admin views use different queries and
	// stay unmasked.
	displayName := func(name string) string {
		if cp.ScoreboardMaskNames {
			return common.MaskName(name)
		}
		return name
	}

	participantIDs, err := pq.getContestParticipantUserIDs(cp.ContestID)
	if err != nil {
		return nil, err
	}

	var users []*model.User
	if len(participantIDs) > 0 {
		if err := pq.db.Where("id IN ?", participantIDs).Find(&users).Error; err != nil {
			return nil, err
		}
	}

	userMap := make(map[int32]*model.User)
	for _, u := range users {
		userMap[int32(u.ID)] = u
	}

	gradingItemIDs := make([]int32, 0, len(cp.Project.GradingItems))
	for _, gi := range cp.Project.GradingItems {
		scoreboard.Heads = append(scoreboard.Heads, &dto.ScoreboardHead{
			ID:   gi.ID,
			Name: gi.Name,
		})
		gradingItemIDs = append(gradingItemIDs, gi.ID)
	}

	visitedUsers := make(map[int32]bool)
	for _, r := range result {
		user := userMap[r.UserID]
		if user == nil {
			pq.logger.Warnf("user not found", "userId", r.UserID)
			continue
		}

		visitedUsers[r.UserID] = true
		statuses := make([]string, 0, len(gradingItemIDs))
		for _, giID := range gradingItemIDs {
			status := "NotSubmitted"
			for _, s := range r.Statuses {
				if s.ID == giID {
					status = s.Status
					break
				}
			}

			statuses = append(statuses, status)
		}

		scoreboard.Rows = append(scoreboard.Rows, &dto.ScoreboardRow{
			Name:     displayName(user.Name),
			Username: user.Username,
			Statuses: statuses,
		})
	}

	for _, u := range users {
		if visitedUsers[int32(u.ID)] {
			continue
		}

		row := &dto.ScoreboardRow{
			Name:     displayName(u.Name),
			Username: u.Username,
			Statuses: make([]string, len(gradingItemIDs)),
		}

		for i := range gradingItemIDs {
			row.Statuses[i] = "NotSubmitted"
		}

		scoreboard.Rows = append(scoreboard.Rows, row)
	}

	return scoreboard, nil
}
