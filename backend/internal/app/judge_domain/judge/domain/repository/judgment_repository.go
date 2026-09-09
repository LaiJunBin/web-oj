package repository

import (
	"time"

	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment"
)

// JudgmentRepository is the persistence port for the judgment aggregate.
//
// A judgment's lifecycle after creation (Queued -> Judging -> a terminal
// status) is written by independent, uncoordinated actors: the worker
// executing it, the cancel-request handler, and (for rejudges) the rollup
// reacting to per-judgment updates. Save is unconditional and therefore only
// safe where nothing else can write concurrently; the conditional methods
// below each guard one specific transition so that "read status, decide,
// write" has no gap another writer can land in.
//
// All four conditional methods report their outcome the same way:
//
//   - applied=true — the update matched and was written. Confirmed persisted.
//   - applied=false — this call did not confirm that it applied. The caller
//     must treat that as "I do not hold this transition": it must not assume
//     its own in-memory state reached the database, and must not act as
//     though it owns the judgment.
//   - err != nil — the outcome is unknown; same caller obligations as false,
//     plus the failure itself is worth reporting.
//
// applied=false is deliberately never inferred by re-reading the document
// and comparing its status to the one we wanted. Status cannot distinguish
// "my write landed" from "another actor wrote this same status", and two
// writers converging on one status (typically Cancelled) is normal rather
// than exceptional — so that inference silently hands the same transition to
// several callers at once.
//
// The consequence is a deliberate bias toward false negatives: a caller can
// come away believing it did not get a transition that in fact was written.
// The bias is chosen because the two error directions are not comparable. A
// false negative leaves a visible, repairable document — a Judging judgment
// with no worker, which OrphanJudgmentRepairService reconciles — whereas a
// false positive means two workers executing one judgment, or telling
// downstream consumers about a result the database does not hold.
//
// For the three state-transition methods, applied=false is additionally as
// close to "nothing was written" as this layer can make it, because they
// bypass the application-level retry wrapper; see their doc comments.
type JudgmentRepository interface {
	FindByID(id string) (*judgment.Judgment, error)
	FindJudgings() ([]*judgment.Judgment, error)
	// Save is an unconditional $set of the whole document (with upsert).
	// Safe only for single-writer paths: creating a judgment
	// (submission/rejudge handlers), or a background repair sweep correcting
	// a judgment nothing else can be touching. Never use it for a save
	// issued while a judgment could be executing or could be the target of a
	// cancel request.
	Save(judgment *judgment.Judgment) error
	LoadQueuedJudgments() ([]*judgment.Judgment, error)

	// TryStart atomically moves judgmentID from Queued to Judging, stamping
	// judgingStartedAt and touching nothing else. This is the executor's
	// ownership claim: exactly one caller can observe applied=true for a
	// given Queued->Judging transition, so it is what keeps two workers
	// holding the same queue entry from both executing it. It also rejects a
	// judgment cancelled while it was still queued — the worker's popped
	// copy still reads Queued and cannot know that on its own. Issued as a
	// single un-retried command, so applied=false means the transition was
	// not taken rather than possibly-resent-and-lost.
	TryStart(judgmentID string, startedAt time.Time) (applied bool, err error)

	// SaveProgressIfJudging persists judgment's full state, only while the
	// document's current status is still Judging. Used for every mid-run
	// save; an unconditional Save here would resurrect a judgment that an
	// out-of-band cancel already moved out of Judging. Unlike the three
	// transition methods this one is retried on transient failures, which is
	// safe because its filter matches the same status it writes: a resent
	// command still matches, so a retry cannot turn a completed write into a
	// spurious applied=false. The entity itself must be Judging — anything
	// else is rejected with an error before any database write, because the
	// filter only constrains the document's status while the payload carries
	// the entity's.
	SaveProgressIfJudging(judgment *judgment.Judgment) (applied bool, err error)

	// SaveTerminalIfJudging persists judgment's computed terminal outcome
	// (Judged/Rejudged/Cancelled/Error), only while the document's current
	// status is still Judging, so a cancellation already accepted in the DB
	// cannot be overwritten by a worker that finished without knowing about
	// it. On applied=false the caller must re-read the document and publish
	// that, never its own computed state — which this call just declined to
	// store. The entity itself must already be terminal — a Judging entity is
	// rejected with an error before any database write. Issued as a single
	// un-retried command.
	SaveTerminalIfJudging(judgment *judgment.Judgment) (applied bool, err error)

	// TryCancel atomically moves judgmentID from Queued or Judging to
	// Cancelled, updating ONLY status/reason/judgingFinishedAt. Deliberately
	// not a Save of a FindByID'd-then-mutated snapshot: that snapshot goes
	// stale the moment the executing worker writes progress it never saw,
	// and a whole-document write would revert that progress. An empty reason
	// leaves any existing reason untouched. Issued as a single un-retried
	// command.
	//
	// Returns the post-update judgment when THIS call won the transition —
	// produced by the same atomic command (find-and-update), never a
	// follow-up read, so a winner always has the entity it must publish
	// events from: with a separate re-read, a transient failure there left
	// the cancellation persisted but unannounced forever, since any retry
	// of TryCancel reports "already cancelled" and no redelivery can
	// recover the lost event. Returns (nil, nil) when the document was
	// already terminal — this call wrote nothing, including its reason.
	TryCancel(judgmentID string, reason string, finishedAt time.Time) (cancelled *judgment.Judgment, err error)
	// UpdateStaffNote sets (nil clears) the judge's contestant-facing note and
	// returns the judgment as stored; (nil, nil) when no such judgment exists.
	// Unconditional and last-write-wins by design — see the implementation.
	UpdateStaffNote(judgmentID string, note *string) (*judgment.Judgment, error)
}
