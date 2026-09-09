package query

import (
	"judge-backend/internal/app/contest_domain/contest/application/query"
	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	"judge-backend/internal/app/shared/model"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"

	"gorm.io/gorm"
)

type ContestQuery struct {
	db     *gorm.DB
	logger sharedtypes.Logger
}

func NewContestQuery(
	db *gorm.DB,
	logger sharedtypes.Logger,
) query.ContestQuery {
	return &ContestQuery{
		db:     db,
		logger: logger,
	}
}

func (cq *ContestQuery) ListPublicContests() ([]dto.Contest, error) {
	cq.logger.Debugf("Starting to list public contests")

	contests := []*model.Contest{}
	if err := cq.db.Find(&contests, "is_public = ?", true).Error; err != nil {
		cq.logger.Errorf("Failed to find public contests", "error", err)
		return nil, err
	}

	result := make([]dto.Contest, 0, len(contests))
	for _, contest := range contests {
		if !contest.IsActivated() {
			continue
		}

		result = append(result, dto.Contest{
			ID:   contest.ID,
			Name: contest.Name,
		})
	}

	cq.logger.Debugf("Successfully listed public contests", "count", len(result))
	return result, nil
}

func (cq *ContestQuery) ListUserContests(userID int32) ([]dto.Contest, error) {
	cq.logger.Debugf("Starting to list user contests", "userId", userID)

	contest := []*model.Contest{}
	err := cq.db.Where(`
		is_public = true
		OR EXISTS (
			SELECT 1 FROM contest_participants cp
			WHERE cp.contest_id = contests.id AND cp.user_id = ?
		)
		OR EXISTS (
			SELECT 1 FROM contest_participants cp
			INNER JOIN user_groups ug ON ug.group_id = cp.group_id
			INNER JOIN `+"`groups`"+` g ON g.id = ug.group_id AND g.deleted_at IS NULL
			WHERE cp.contest_id = contests.id AND ug.user_id = ?
		)
	`, userID, userID).Find(&contest).Error
	if err != nil {
		cq.logger.Errorf("Failed to find user contests", "userId", userID, "error", err)
		return nil, err
	}

	result := make([]dto.Contest, 0, len(contest))
	for _, c := range contest {
		if !c.IsActivated() {
			continue
		}

		result = append(result, dto.Contest{
			ID:   c.ID,
			Name: c.Name,
		})
	}

	cq.logger.Debugf("Successfully listed user contests", "userId", userID, "count", len(result))
	return result, nil
}

func (cq *ContestQuery) GetPublicContestByID(contestID int32) (dto.ContestDetail, error) {
	cq.logger.Debugf("Starting to get public contest by ID", "contestId", contestID)

	contest := model.Contest{}
	if err := cq.db.Preload("Projects", func(tx *gorm.DB) *gorm.DB {
		return tx.Preload("Contest").Preload("Project").Where("is_public = ?", true)
	}).First(&contest, "id = ? AND is_public = ?", contestID, true).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			cq.logger.Debugf("Public contest not found", "contestId", contestID)
			return dto.ContestDetail{}, query.ErrorContestNotFound
		}
		cq.logger.Errorf("Failed to find public contest", "contestId", contestID, "error", err)
		return dto.ContestDetail{}, err
	}

	if !contest.IsActivated() {
		cq.logger.Debugf("Public contest is not activated", "contestId", contestID)
		return dto.ContestDetail{}, query.ErrorContestNotFound
	}

	projects := make([]dto.Project, 0, len(contest.Projects))
	for _, project := range contest.Projects {
		if !project.IsActivated() {
			continue
		}

		projects = append(projects, dto.Project{
			ID:     project.ID,
			Name:   project.Name,
			Status: project.Status(),
		})
	}

	cq.logger.Debugf("Successfully got public contest", "contestId", contestID, "projectCount", len(projects))
	return dto.ContestDetail{
		Contest: dto.Contest{
			ID:   contest.ID,
			Name: contest.Name,
		},
		Projects: projects,
	}, nil
}

func (cq *ContestQuery) isUserParticipant(contestID, userID int32) (bool, error) {
	return isContestParticipant(cq.db, contestID, userID)
}

func (cq *ContestQuery) GetUserContestByID(userID, contestID int32) (dto.ContestDetail, error) {
	cq.logger.Debugf("Starting to get user contest by ID", "userId", userID, "contestId", contestID)

	contest := model.Contest{}
	if err := cq.db.Preload("Projects", func(tx *gorm.DB) *gorm.DB {
		return tx.Preload("Contest").Preload("Project")
	}).First(&contest, "id = ?", contestID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			cq.logger.Debugf("User contest not found", "userId", userID, "contestId", contestID)
			return dto.ContestDetail{}, query.ErrorContestNotFound
		}
		cq.logger.Errorf("Failed to find user contest", "userId", userID, "contestId", contestID, "error", err)
		return dto.ContestDetail{}, err
	}

	if !contest.IsActivated() {
		cq.logger.Debugf("User contest is not activated", "userId", userID, "contestId", contestID)
		return dto.ContestDetail{}, query.ErrorContestNotFound
	}

	if !contest.IsPublic {
		ok, err := cq.isUserParticipant(contestID, userID)
		if err != nil {
			cq.logger.Errorf("Failed to check participant", "userId", userID, "contestId", contestID, "error", err)
			return dto.ContestDetail{}, err
		}
		if !ok {
			cq.logger.Debugf("User is not a participant of non-public contest", "userId", userID, "contestId", contestID)
			return dto.ContestDetail{}, httperror.New(http.StatusForbidden, "access denied")
		}
	}

	projects := make([]dto.Project, 0, len(contest.Projects))
	for _, project := range contest.Projects {
		if !project.IsActivated() {
			continue
		}

		projects = append(projects, dto.Project{
			ID:     project.ID,
			Name:   project.Name,
			Status: project.Status(),
		})
	}

	cq.logger.Debugf("Successfully got user contest", "userId", userID, "contestId", contestID, "projectCount", len(projects))
	return dto.ContestDetail{
		Contest: dto.Contest{
			ID:   contest.ID,
			Name: contest.Name,
		},
		Projects: projects,
	}, nil
}

func (cq *ContestQuery) GetUserContestScheduleByID(userID, contestID int32) (dto.ContestSchedule, error) {
	cq.logger.Debugf("Starting to get user contest schedule by ID", "userId", userID, "contestId", contestID)

	contest := model.Contest{}
	if err := cq.db.Preload("Projects", func(tx *gorm.DB) *gorm.DB {
		return tx.Preload("Contest").Preload("Project", func(tx *gorm.DB) *gorm.DB {
			return tx.Preload("GradingItems")
		})
	}).First(&contest, "id = ?", contestID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			cq.logger.Debugf("User contest schedule not found", "contestId", contestID)
			return dto.ContestSchedule{}, query.ErrorContestNotFound
		}
		cq.logger.Errorf("Failed to find user contest schedule", "contestId", contestID, "error", err)
		return dto.ContestSchedule{}, err
	}

	if !contest.IsActivated() {
		cq.logger.Debugf("User contest is not activated", "contestId", contestID)
		return dto.ContestSchedule{}, query.ErrorContestNotFound
	}

	if !contest.IsPublic {
		ok, err := cq.isUserParticipant(contestID, userID)
		if err != nil {
			cq.logger.Errorf("Failed to check participant", "userId", userID, "contestId", contestID, "error", err)
			return dto.ContestSchedule{}, err
		}
		if !ok {
			cq.logger.Debugf("User is not a participant of non-public contest", "userId", userID, "contestId", contestID)
			return dto.ContestSchedule{}, httperror.New(http.StatusForbidden, "access denied")
		}
	}

	projects := make([]*dto.ContestScheduleProject, 0, len(contest.Projects))
	for _, project := range contest.Projects {
		if !project.IsActivated() {
			continue
		}

		var score float64 = 0
		for _, gradingItem := range project.Project.GradingItems {
			score += gradingItem.Score
		}

		projects = append(projects, &dto.ContestScheduleProject{
			Name:        project.Name,
			Description: project.Project.Description,
			StartTime:   project.StartTime,
			EndTime:     project.EndTime,
			Score:       score,
		})
	}

	cq.logger.Debugf("Successfully got user contest schedule", "contestId", contestID, "projectCount", len(projects))
	return dto.ContestSchedule{
		IsPublic: contest.IsPublic,
		Projects: projects,
	}, nil
}

func (cq *ContestQuery) GetPublicContestScheduleByID(contestID int32) (dto.ContestSchedule, error) {
	cq.logger.Debugf("Starting to get public contest schedule by ID", "contestId", contestID)

	contest := model.Contest{}
	if err := cq.db.Preload("Projects", func(tx *gorm.DB) *gorm.DB {
		return tx.Preload("Contest").Preload("Project", func(tx *gorm.DB) *gorm.DB {
			return tx.Preload("GradingItems")
		})
	}).First(&contest, "id = ?", contestID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			cq.logger.Debugf("Public contest schedule not found", "contestId", contestID)
			return dto.ContestSchedule{}, query.ErrorContestNotFound
		}
		cq.logger.Errorf("Failed to find public contest schedule", "contestId", contestID, "error", err)
		return dto.ContestSchedule{}, err
	}

	if !contest.IsActivated() || !contest.IsPublic {
		cq.logger.Debugf("Public contest is not activated or not public", "contestId", contestID)
		return dto.ContestSchedule{}, query.ErrorContestNotFound
	}

	projects := make([]*dto.ContestScheduleProject, 0, len(contest.Projects))
	for _, project := range contest.Projects {
		if !project.IsActivated() || !project.IsPublic {
			continue
		}

		var score float64 = 0
		for _, gradingItem := range project.Project.GradingItems {
			score += gradingItem.Score
		}

		projects = append(projects, &dto.ContestScheduleProject{
			Name:        project.Name,
			Description: project.Project.Description,
			StartTime:   project.StartTime,
			EndTime:     project.EndTime,
			Score:       score,
		})
	}

	cq.logger.Debugf("Successfully got public contest schedule", "contestId", contestID, "projectCount", len(projects))
	return dto.ContestSchedule{
		IsPublic: contest.IsPublic,
		Projects: projects,
	}, nil
}
