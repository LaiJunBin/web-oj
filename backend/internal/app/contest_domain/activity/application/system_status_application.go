package application

import (
	getsubmissionrate "judge-backend/internal/app/contest_domain/activity/application/usecase/system_status/get_submission_rate"
	listjudgmenterrors "judge-backend/internal/app/contest_domain/activity/application/usecase/system_status/list_judgment_errors"
)

type SystemStatusApplication struct {
	getSubmissionRate  getsubmissionrate.Usecase
	listJudgmentErrors listjudgmenterrors.Usecase
}

func NewSystemStatusApplication(
	getSubmissionRate getsubmissionrate.Usecase,
	listJudgmentErrors listjudgmenterrors.Usecase,
) *SystemStatusApplication {
	return &SystemStatusApplication{
		getSubmissionRate:  getSubmissionRate,
		listJudgmentErrors: listJudgmentErrors,
	}
}

func (a SystemStatusApplication) GetSubmissionRate(input *getsubmissionrate.Input) (*getsubmissionrate.Output, error) {
	return a.getSubmissionRate.Execute(input)
}

func (a SystemStatusApplication) ListJudgmentErrors(input *listjudgmenterrors.Input) (*listjudgmenterrors.Output, error) {
	return a.listJudgmentErrors.Execute(input)
}
