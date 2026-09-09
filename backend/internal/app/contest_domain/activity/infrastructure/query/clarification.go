package query

import (
	"judge-backend/internal/app/contest_domain/activity/application/query"
	"judge-backend/internal/app/contest_domain/activity/application/query/dto"
	"judge-backend/internal/app/shared/model"
	"judge-backend/internal/app/shared/rolebits"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"

	"gorm.io/gorm"
)

type ClarificationQuery struct {
	db        *gorm.DB
	adminMask rolebits.AdminRoleMask
}

func NewClarificationQuery(
	db *gorm.DB,
	masks rolebits.RoleMasks,
) query.ClarificationQuery {
	return &ClarificationQuery{
		db:        db,
		adminMask: masks.Admin,
	}
}

func (q *ClarificationQuery) ListClarifications(contestProjectID int32) ([]*dto.Clarification, error) {
	var count int64
	if err := q.db.Model(&model.ContestProject{}).Where("id = ?", contestProjectID).Count(&count).Error; err != nil {
		return nil, err
	}

	if count == 0 {
		return nil, httperror.New(http.StatusNotFound, "contest project not found")
	}

	var mcs []*model.Clarification
	if err := q.db.
		Preload("Sender").
		Preload("Receiver").
		Where("contest_project_id = ?", contestProjectID).
		Order("created_at DESC").Find(&mcs).Error; err != nil {
		return nil, err
	}

	var cs []*dto.Clarification
	for _, mc := range mcs {
		cs = append(cs, q.transformClarification(mc))
	}

	return cs, nil
}

func preloadReplyClarification(db *gorm.DB) *gorm.DB {
	return db.Preload("Sender").Preload("Receiver").Preload("ReplyClarification", preloadReplyClarification)
}

func (q *ClarificationQuery) GetClarificationByID(id int32) (*dto.ClarificationDetail, error) {
	mc := &model.Clarification{}
	if err := q.db.
		Preload("Sender").
		Preload("Receiver").
		Preload("ReplyClarification", preloadReplyClarification).
		First(mc, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, err
	}

	return q.transformClarificationDetail(mc), nil
}

func (q *ClarificationQuery) transformClarification(mc *model.Clarification) *dto.Clarification {
	if mc == nil {
		return nil
	}

	isAdmin := false
	senderUser := dto.UnknownClarificationUserInfo
	if mc.Sender != nil {
		senderUser = dto.ClarificationUserInfo{
			ID:       int32(mc.Sender.ID),
			Name:     mc.Sender.Name,
			Username: mc.Sender.Username,
		}
		isAdmin = rolebits.HasRole(mc.Sender.Roles, q.adminMask)
	}
	var receiverUser *dto.ClarificationUserInfo
	if mc.Receiver != nil {
		receiverUser = &dto.ClarificationUserInfo{
			ID:       int32(mc.Receiver.ID),
			Name:     mc.Receiver.Name,
			Username: mc.Receiver.Username,
		}
	}

	return &dto.Clarification{
		ID:           mc.ID,
		Subject:      mc.Subject,
		Content:      mc.Content,
		SenderUser:   senderUser,
		ReceiverUser: receiverUser,
		SendToAll:    mc.SendToAll,
		IsAdmin:      isAdmin,
		IsAnswered:   mc.IsAnswered,
		CreatedAt:    mc.CreatedAt,
		UpdatedAt:    mc.UpdatedAt,
	}
}

func (q *ClarificationQuery) transformClarificationDetail(mc *model.Clarification) *dto.ClarificationDetail {
	if mc == nil {
		return nil
	}

	return &dto.ClarificationDetail{
		Clarification:      q.transformClarification(mc),
		ReplyClarification: q.transformClarificationDetail(mc.ReplyClarification),
	}
}
