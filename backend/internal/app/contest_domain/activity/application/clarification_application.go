package application

import (
	getclarification "judge-backend/internal/app/contest_domain/activity/application/usecase/clarification/get_clarification"
	listclarifications "judge-backend/internal/app/contest_domain/activity/application/usecase/clarification/list_clarifications"
	replyclarification "judge-backend/internal/app/contest_domain/activity/application/usecase/clarification/reply_clarification"
	sendclarification "judge-backend/internal/app/contest_domain/activity/application/usecase/clarification/send_clarification"
	updateclarification "judge-backend/internal/app/contest_domain/activity/application/usecase/clarification/update_clarification"
)

type ClarificationApplication struct {
	listClarification   listclarifications.Usecase
	getClarification    getclarification.Usecase
	sendClarification   sendclarification.Usecase
	replyClarification  replyclarification.Usecase
	updateClarification updateclarification.Usecase
}

func NewClarificationApplication(
	listClarification listclarifications.Usecase,
	getClarification getclarification.Usecase,
	sendClarification sendclarification.Usecase,
	replyClarification replyclarification.Usecase,
	updateClarification updateclarification.Usecase,
) *ClarificationApplication {
	return &ClarificationApplication{
		listClarification:   listClarification,
		getClarification:    getClarification,
		sendClarification:   sendClarification,
		replyClarification:  replyClarification,
		updateClarification: updateClarification,
	}
}

func (a ClarificationApplication) ListClarifications(input *listclarifications.Input) (*listclarifications.Output, error) {
	return a.listClarification.Execute(input)
}

func (a ClarificationApplication) GetClarification(input *getclarification.Input) (*getclarification.Output, error) {
	return a.getClarification.Execute(input)
}

func (a ClarificationApplication) SendClarification(input *sendclarification.Input) (*sendclarification.Output, error) {
	return a.sendClarification.Execute(input)
}

func (a ClarificationApplication) ReplyClarification(input *replyclarification.Input) (*replyclarification.Output, error) {
	return a.replyClarification.Execute(input)
}

func (a ClarificationApplication) UpdateClarification(input *updateclarification.Input) (*updateclarification.Output, error) {
	return a.updateClarification.Execute(input)
}
