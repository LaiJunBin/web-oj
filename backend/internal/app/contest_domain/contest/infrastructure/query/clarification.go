package query

import (
	"judge-backend/internal/app/contest_domain/contest/application/query"
	"judge-backend/internal/app/contest_domain/contest/application/query/dto"
	"judge-backend/internal/app/shared/model"
	"judge-backend/internal/app/shared/rolebits"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"

	"gorm.io/gorm"
)

type ClarificationQuery struct {
	db        *gorm.DB
	adminMask rolebits.AdminRoleMask
	logger    sharedtypes.Logger
}

func NewClarificationQuery(
	db *gorm.DB,
	masks rolebits.RoleMasks,
	logger sharedtypes.Logger,
) query.ClarificationQuery {
	return &ClarificationQuery{
		db:        db,
		adminMask: masks.Admin,
		logger:    logger,
	}
}

func (q *ClarificationQuery) ListUserClarificationsByProjectID(userID, contestProjectID int32) ([]*dto.Clarification, error) {
	q.logger.Debugf("Starting to list user clarifications by project ID", "userId", userID, "contestProjectId", contestProjectID)

	cp := &model.ContestProject{}
	if err := q.db.Preload("Contest").First(cp, contestProjectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			q.logger.Debugf("Contest project not found", "contestProjectId", contestProjectID)
			return nil, httperror.New(http.StatusNotFound, "contest project not found")
		}

		q.logger.Errorf("Failed to find contest project", "contestProjectId", contestProjectID, "error", err)
		return nil, err
	}

	if !cp.IsActivated() {
		q.logger.Debugf("Contest project is not activated", "contestProjectId", contestProjectID)
		return nil, httperror.New(http.StatusNotFound, "contest project not found")
	}

	var mcs []*model.Clarification
	if err := q.db.
		Preload("Sender").
		Preload("Receiver").
		Where("contest_project_id = ? AND (send_to_all = ? OR sender_user_id = ? OR receiver_user_id = ?)", contestProjectID, true, userID, userID).
		Order("created_at DESC").
		Find(&mcs).Error; err != nil {
		q.logger.Errorf("Failed to find clarifications", "userId", userID, "contestProjectId", contestProjectID, "error", err)
		return nil, err
	}

	var cs []*dto.Clarification
	for _, mc := range mcs {
		cs = append(cs, &dto.Clarification{
			ID:               mc.ID,
			Subject:          mc.Subject,
			Content:          mc.Content,
			IsAdminBroadcast: mc.SendToAll,
			IsAdminResponse:  rolebits.HasRole(mc.Sender.Roles, q.adminMask) && mc.Receiver != nil && mc.Receiver.ID == int(userID),
			CreatedAt:        mc.CreatedAt,
		})
	}

	q.logger.Debugf("Successfully listed user clarifications", "userId", userID, "contestProjectId", contestProjectID, "count", len(cs))
	return cs, nil
}

func preloadReplyClarification(db *gorm.DB) *gorm.DB {
	return db.Preload("Sender").Preload("Receiver").Preload("ReplyClarification", preloadReplyClarification)
}

func (q *ClarificationQuery) GetUserClarificationByID(userID, clarificationID int32) (*dto.ClarificationDetail, error) {
	q.logger.Debugf("Starting to get user clarification by ID", "userId", userID, "clarificationId", clarificationID)

	mc := &model.Clarification{}
	if err := q.db.
		Preload("ContestProject", func(tx *gorm.DB) *gorm.DB {
			return tx.Preload("Contest").Preload("Project")
		}).
		Preload("Sender").
		Preload("Receiver").
		Preload("ReplyClarification", preloadReplyClarification).
		First(mc, clarificationID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			q.logger.Debugf("Clarification not found", "clarificationId", clarificationID)
			return nil, httperror.New(http.StatusNotFound, "clarification not found")
		}

		q.logger.Errorf("Failed to find clarification", "clarificationId", clarificationID, "error", err)
		return nil, err
	}

	if !mc.ContestProject.IsActivated() {
		q.logger.Debugf("Contest project is not activated", "contestProjectId", mc.ContestProjectID)
		return nil, httperror.New(http.StatusNotFound, "clarification not found")
	}

	if !q.userCanViewClarification(userID, mc) {
		q.logger.Debugf("User cannot view clarification", "userId", userID, "clarificationId", clarificationID)
		return nil, httperror.New(http.StatusNotFound, "clarification not found")
	}

	result := q.transformClarification(userID, mc)
	q.logger.Debugf("Successfully got user clarification", "userId", userID, "clarificationId", clarificationID)
	return result, nil
}

func (q *ClarificationQuery) userCanViewClarification(userID int32, mc *model.Clarification) bool {
	if mc.SenderUserID != nil && int32(*mc.SenderUserID) == userID {
		return true
	}
	if mc.ReceiverUserID != nil && int32(*mc.ReceiverUserID) == userID {
		return true
	}

	if !mc.SendToAll {
		return false
	}

	// broadcasts of non-public contests are limited to participants
	if mc.ContestProject.Contest.IsPublic {
		return true
	}
	ok, err := isContestParticipant(q.db, mc.ContestProject.ContestID, userID)
	if err != nil {
		q.logger.Errorf("Failed to check participant", "userId", userID, "contestId", mc.ContestProject.ContestID, "error", err)
		return false
	}
	return ok
}

func (q *ClarificationQuery) transformClarification(userID int32, mc *model.Clarification) *dto.ClarificationDetail {
	if mc == nil {
		return nil
	}

	return &dto.ClarificationDetail{
		Clarification: dto.Clarification{
			ID:               mc.ID,
			Subject:          mc.Subject,
			Content:          mc.Content,
			IsAdminBroadcast: mc.SendToAll,
			IsAdminResponse:  rolebits.HasRole(mc.Sender.Roles, q.adminMask) && mc.Receiver != nil,
			CreatedAt:        mc.CreatedAt,
		},
		ReplyClarification: q.transformClarification(userID, mc.ReplyClarification),
	}
}
