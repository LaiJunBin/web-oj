package application

import (
	canceljudge "judge-backend/internal/app/contest_domain/activity/application/usecase/judgment/cancel_judge"
	setstaffnote "judge-backend/internal/app/contest_domain/activity/application/usecase/judgment/set_staff_note"
)

type JudgmentApplication struct {
	cancelJudge  canceljudge.Usecase
	setStaffNote setstaffnote.Usecase
}

func NewJudgmentApplication(
	cancelJudge canceljudge.Usecase,
	setStaffNote setstaffnote.Usecase,
) *JudgmentApplication {
	return &JudgmentApplication{
		cancelJudge:  cancelJudge,
		setStaffNote: setStaffNote,
	}
}

func (a JudgmentApplication) CancelJudge(input *canceljudge.Input) (*canceljudge.Output, error) {
	return a.cancelJudge.Execute(input)
}

func (a JudgmentApplication) SetStaffNote(input *setstaffnote.Input) (*setstaffnote.Output, error) {
	return a.setStaffNote.Execute(input)
}
