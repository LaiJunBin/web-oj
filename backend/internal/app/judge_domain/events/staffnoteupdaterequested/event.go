package staffnoteupdaterequested

// Event carries a staff-note replacement into the judge domain. A nil
// staffNote clears the note.
type Event struct {
	id         string
	judgmentID string
	staffNote  *string
}

func New(id, judgmentID string, staffNote *string) *Event {
	return &Event{id: id, judgmentID: judgmentID, staffNote: staffNote}
}

func (e *Event) Topic() string      { return "judge_staff_note_update_requested" }
func (e *Event) ID() string         { return e.id }
func (e *Event) JudgmentID() string { return e.judgmentID }
func (e *Event) StaffNote() *string { return e.staffNote }
