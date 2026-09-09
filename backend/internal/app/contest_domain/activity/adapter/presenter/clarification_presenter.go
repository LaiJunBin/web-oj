package presenter

import jsonpresenter "judge-backend/internal/app/contest_domain/activity/adapter/presenter/json/clarification"

type ClarificationPresenter struct {
	ListClarifications            jsonpresenter.ListClarificationsPresenter
	GetClarification              jsonpresenter.GetClarificationPresenter
	SendClarification             jsonpresenter.SendClarificationPresenter
	ReplyClarification            jsonpresenter.ReplyClarificationPresenter
	MarkClarificationAsAnswered   jsonpresenter.MarkClarificationAsAnsweredPresenter
	MarkClarificationAsUnanswered jsonpresenter.MarkClarificationAsUnansweredPresenter
}

func NewClarificationPresenter() *ClarificationPresenter {
	return &ClarificationPresenter{
		ListClarifications:            jsonpresenter.NewListClarificationsPresenter(),
		GetClarification:              jsonpresenter.NewGetClarificationPresenter(),
		SendClarification:             jsonpresenter.NewSendClarificationPresenter(),
		ReplyClarification:            jsonpresenter.NewReplyClarificationPresenter(),
		MarkClarificationAsAnswered:   jsonpresenter.NewMarkClarificationAsAnsweredPresenter(),
		MarkClarificationAsUnanswered: jsonpresenter.NewMarkClarificationAsUnansweredPresenter(),
	}
}
