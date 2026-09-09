package repository

import "judge-backend/internal/app/judge_domain/judge/domain/entity/rejudge"

type RejudgeRepository interface {
	FindByID(id string) (*rejudge.Rejudge, error)
	Save(rejudge *rejudge.Rejudge) error
	AllJudgmentsFinished(rejudgeID string) (bool, error)

	// CommitJudgmentsByRejudgeID moves rejudgeID from FINISHED to
	// APPROVED/REJECTED and (on approval) marks its judgments JUDGED. The
	// FINISHED→final flip is a conditional write and the race arbiter: two
	// admins deciding the same rejudge both pass the application layer's
	// CanApprove/CanReject check (a stale read, not a lock), so this DB-level
	// condition is what actually prevents both from committing —
	// applied=false means a concurrent commit already won the race, and the
	// caller should treat that as a no-op, not an error. The two writes are
	// not atomic (the deployed standalone MongoDB has no transactions); a
	// crash in between is healed by RepairUncommittedApprovals.
	CommitJudgmentsByRejudgeID(rejudgeID string, approval bool) (applied bool, err error)

	// RepairUncommittedApprovals replays the judgments→JUDGED flip for
	// rejudges that are APPROVED but whose judgments were left in Rejudged by
	// a crash between CommitJudgmentsByRejudgeID's two writes. Run at
	// startup and then on RejudgeCommitRepairService's periodic sweep;
	// returns the number of judgments repaired.
	RepairUncommittedApprovals() (repaired int64, err error)

	// TransitionStatus atomically moves rejudgeID's status to target, but
	// only if its current stored status is one of allowedFrom — checked and
	// applied as a single conditional DB update, not a separate read then
	// write. Returns applied=false (no error) when the current status
	// wasn't in allowedFrom; the caller should treat that as "a concurrent
	// update already moved this rejudge past the point where this
	// transition still applies" and not retry or error.
	//
	// Use this instead of a FindByID-then-Save cycle for any status change
	// after the rejudge already exists: two concurrent writers (the
	// automatic per-judgment status rollup, and manual admin cancellation)
	// can run against the same rejudge document, and Save persists the
	// caller's full in-memory snapshot — whichever Save happens to run last
	// wins even if it was computed from a staler read, silently reverting a
	// newer (possibly terminal) status.
	TransitionStatus(rejudgeID string, allowedFrom []*rejudge.Status, target *rejudge.Status) (applied bool, err error)
}
