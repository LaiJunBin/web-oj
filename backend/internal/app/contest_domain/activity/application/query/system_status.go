package query

import "judge-backend/internal/app/contest_domain/activity/application/query/dto"

// SystemStatusQuery serves the whole-instance views the admin status page is
// built on. Deliberately separate from the per-contest-project queries
// alongside it: everything here spans every contest project, and mixing the two
// scopes behind one interface is how a project-scoped filter ends up quietly
// missing from a system-wide read.
type SystemStatusQuery interface {
	// GetSubmissionRate returns one bucket per minute over the last `minutes`
	// minutes, oldest first, with empty minutes present as zero. A nil
	// contestProjectID means every contest project.
	GetSubmissionRate(minutes int32, contestProjectID *int32) ([]*dto.SubmissionRatePoint, error)

	// ListJudgmentErrors returns the most recently created errored judgments,
	// newest first.
	ListJudgmentErrors(limit int32) ([]*dto.JudgmentError, error)
}
