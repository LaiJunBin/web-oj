package repository

import (
	"errors"

	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/gradingitem"
)

// ErrGradingItemsNotOwned reports a submission whose recorded grading-item ids
// cannot all be resolved within its own contest project's problem. Handlers
// that create judgments translate it into a terminal ERROR judgment: judging
// a set the submission was never allowed to name must fail loudly, and the
// only ids allowed to be silently absent are this problem's own soft-deleted
// items (see the repository implementation).
var ErrGradingItemsNotOwned = errors.New("submission grading items cannot be resolved for its project")

type GradingItemRepository interface {
	FindBySubmissionID(submissionID string) ([]*gradingitem.GradingItem, error)
}
