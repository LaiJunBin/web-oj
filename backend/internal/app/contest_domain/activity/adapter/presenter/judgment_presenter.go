package presenter

import jsonpresenter "judge-backend/internal/app/contest_domain/activity/adapter/presenter/json/judgment"

type JudgmentPresenter struct {
	CancelJudge  jsonpresenter.CancelJudgePresenter
	SetStaffNote jsonpresenter.SetStaffNotePresenter
}

func NewJudgmentPresenter() *JudgmentPresenter {
	return &JudgmentPresenter{
		CancelJudge:  jsonpresenter.NewCancelJudgePresenter(),
		SetStaffNote: jsonpresenter.NewSetStaffNotePresenter(),
	}
}
