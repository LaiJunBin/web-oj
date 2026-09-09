package query

import (
	"judge-backend/internal/app/contest_domain/contest/application/query"
	"judge-backend/internal/app/shared/model"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/httperror"
	"net/http"

	"gorm.io/gorm"
)

var errNotParticipant = httperror.New(http.StatusForbidden, "not a participant of this contest")

type ContestProjectQuery struct {
	db     *gorm.DB
	logger sharedtypes.Logger
}

func NewContestProjectQuery(
	db *gorm.DB,
	logger sharedtypes.Logger,
) query.ContestProjectQuery {
	return &ContestProjectQuery{
		db:     db,
		logger: logger,
	}
}

func (q *ContestProjectQuery) CheckProjectAndGradingItemsExists(contestProjectID int32, gradingItemIDs []int32) error {
	q.logger.Debugf("Starting to check project and grading items exists", "contestProjectId", contestProjectID, "gradingItemIds", gradingItemIDs)

	cp := &model.ContestProject{}
	if err := q.db.Where("id = ?", contestProjectID).First(cp).Error; err != nil {
		q.logger.Debugf("Contest project not found", "contestProjectId", contestProjectID)
		return httperror.New(http.StatusNotFound, "contest project not found")
	}

	// An empty list is a whole-project submission, not a claim about specific
	// ids — it must pass even for a problem with no grading items at all.
	if len(gradingItemIDs) == 0 {
		return nil
	}

	// The resolved set must EQUAL the requested set (after collapsing
	// duplicates), scoped to this contest project's problem. Anything less —
	// an id of another problem, a deleted item, a ghost — and the whole
	// submission is refused: the ids name what the judge will run, so an id
	// that resolves to nothing here must not reach it.
	unique := make([]int32, 0, len(gradingItemIDs))
	seen := make(map[int32]bool, len(gradingItemIDs))
	for _, id := range gradingItemIDs {
		if !seen[id] {
			seen[id] = true
			unique = append(unique, id)
		}
	}

	var found []int32
	if err := q.db.Model(&model.GradingItem{}).
		Where("project_id = ? AND id IN ?", cp.ProjectID, unique).
		Pluck("id", &found).Error; err != nil {
		q.logger.Errorf("Failed to find grading items", "projectId", cp.ProjectID, "gradingItemIds", gradingItemIDs, "error", err)
		return httperror.New(http.StatusBadRequest, "invalid grading item ids")
	}

	if len(found) != len(unique) {
		q.logger.Debugf("Grading item ids do not all resolve to this problem", "contestProjectId", contestProjectID, "requested", unique, "found", found)
		return httperror.New(http.StatusBadRequest, "invalid grading item ids")
	}

	q.logger.Debugf("Successfully checked project and grading items exists", "contestProjectId", contestProjectID, "gradingItemIds", gradingItemIDs)
	return nil
}

func (q *ContestProjectQuery) CheckUserIsParticipant(contestProjectID int32, userID int32) error {
	q.logger.Debugf("Checking user is participant", "contestProjectId", contestProjectID, "userId", userID)

	cp := &model.ContestProject{}
	if err := q.db.Select("contest_id").Where("id = ?", contestProjectID).First(cp).Error; err != nil {
		return httperror.New(http.StatusNotFound, "contest project not found")
	}

	ok, err := isContestParticipant(q.db, cp.ContestID, userID)
	if err != nil {
		return err
	}

	if !ok {
		q.logger.Debugf("User is not a participant", "contestProjectId", contestProjectID, "userId", userID)
		return errNotParticipant
	}

	q.logger.Debugf("User is a participant", "contestProjectId", contestProjectID, "userId", userID)
	return nil
}

// CheckUserCanAccessProject allows any user on public contests; non-public
// contests require participant membership.
func (q *ContestProjectQuery) CheckUserCanAccessProject(contestProjectID int32, userID int32) error {
	q.logger.Debugf("Checking user can access project", "contestProjectId", contestProjectID, "userId", userID)

	cp := &model.ContestProject{}
	if err := q.db.Preload("Contest").Where("id = ?", contestProjectID).First(cp).Error; err != nil {
		return httperror.New(http.StatusNotFound, "contest project not found")
	}

	if !cp.IsActivated() {
		q.logger.Debugf("Contest project is not activated", "contestProjectId", contestProjectID)
		return httperror.New(http.StatusNotFound, "contest project not found")
	}

	if cp.Contest.IsPublic {
		return nil
	}

	ok, err := isContestParticipant(q.db, cp.ContestID, userID)
	if err != nil {
		return err
	}

	if !ok {
		q.logger.Debugf("User cannot access non-public contest project", "contestProjectId", contestProjectID, "userId", userID)
		return errNotParticipant
	}

	return nil
}

// GetSubmissionCooldownSeconds reads the per-contest-project override. nil
// means the contest did not set one and the system default applies; a stored
// 0 means this contest deliberately runs without a cooldown, so the two must
// not be collapsed.
func (q *ContestProjectQuery) GetSubmissionCooldownSeconds(contestProjectID int32) (*int32, error) {
	cp := &model.ContestProject{}
	if err := q.db.Select("submission_cooldown_seconds").
		Where("id = ?", contestProjectID).First(cp).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, httperror.New(http.StatusNotFound, "contest project not found")
		}
		q.logger.Errorf("Failed to read submission cooldown", "contestProjectId", contestProjectID, "error", err)
		return nil, err
	}

	return cp.SubmissionCooldownSeconds, nil
}
