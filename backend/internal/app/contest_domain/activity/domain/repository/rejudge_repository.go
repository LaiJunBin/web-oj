package repository

import "judge-backend/internal/app/contest_domain/activity/domain/entity/rejudge"

type RejudgeRepository interface {
	FindByID(id string) (*rejudge.Rejudge, error)

	// FindAwaitingDecisionBySubmissionIDs returns the rejudges covering any of the
	// given submissions whose result an admin has not yet accepted or discarded.
	// Starting another rejudge for those submissions while one is outstanding is
	// what allows two of them to be decided out of order, so callers use this to
	// refuse the request rather than to report on it.
	FindAwaitingDecisionBySubmissionIDs(submissionIDs []string) ([]*rejudge.Rejudge, error)
}
