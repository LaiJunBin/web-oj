package service

type JudgmentService interface {
	RequestCancelJudgeByID(id string, reason string)
	// RequestUpdateStaffNoteByID replaces the judgment's staff note; nil
	// clears it.
	RequestUpdateStaffNoteByID(id string, staffNote *string)
}
