package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/activity/adapter/presenter/types"
	setstaffnote "judge-backend/internal/app/contest_domain/activity/application/usecase/judgment/set_staff_note"
	"judge-backend/internal/app/shared/types/httperror"
)

type SetStaffNotePresenter types.Presenter[setstaffnote.Output, api.SetJudgmentStaffNoteResponseObject]

type setStaffNotePresenter struct {
}

func NewSetStaffNotePresenter() SetStaffNotePresenter {
	return &setStaffNotePresenter{}
}

func (p *setStaffNotePresenter) Output(output *setstaffnote.Output) api.SetJudgmentStaffNoteResponseObject {
	return api.SetJudgmentStaffNote200JSONResponse{
		Success: true,
	}
}

func (p *setStaffNotePresenter) Error(err error) api.SetJudgmentStaffNoteResponseObject {
	return httperror.From(err).Response()
}
