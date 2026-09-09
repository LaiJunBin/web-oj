package repository

import (
	"time"

	"judge-backend/internal/app/judge_domain/judge/domain/entity/submission"
)

type SubmissionRepository interface {
	FindByID(id string) (*submission.Submission, error)
	// FindWithoutJudgment lists submissions that have no judgment and were
	// created before olderThan. The cutoff is what makes the orphan sweep
	// safe: judgment creation is asynchronous, so for a moment after Save
	// every submission is judgment-less — a sweep reading in that window
	// used to race a terminal ERROR judgment against the real one, and the
	// two are written independently (no uniqueness constraint arbitrates).
	// A submission only counts as orphaned once it has been judgment-less
	// for longer than any healthy creation path can take.
	FindWithoutJudgment(olderThan time.Time) ([]*submission.Submission, error)
	Save(submission *submission.Submission) error
}
