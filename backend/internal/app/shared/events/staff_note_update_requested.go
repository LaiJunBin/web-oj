package sharedevent

import (
	sharedtypes "judge-backend/internal/app/shared/types"

	"github.com/google/uuid"
)

// StaffNoteUpdateRequested asks the judge domain to replace a judgment's
// staff note. A nil note clears it — the admin API maps an empty string to
// nil before publishing, so "clear" and "set" travel the same path.
type StaffNoteUpdateRequested struct {
	id         string
	judgmentID string
	staffNote  *string
}

func NewStaffNoteUpdateRequested(judgmentID string, staffNote *string) sharedtypes.Event {
	return &StaffNoteUpdateRequested{
		id:         uuid.New().String(),
		judgmentID: judgmentID,
		staffNote:  staffNote,
	}
}

func (e *StaffNoteUpdateRequested) Topic() string      { return "shared:staff_note_update_requested" }
func (e *StaffNoteUpdateRequested) ID() string         { return e.id }
func (e *StaffNoteUpdateRequested) JudgmentID() string { return e.judgmentID }
func (e *StaffNoteUpdateRequested) StaffNote() *string { return e.staffNote }
