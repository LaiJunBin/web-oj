package application

import (
	approverejudge "judge-backend/internal/app/contest_domain/activity/application/usecase/rejudge/approve_rejudge"
	cancelrejudge "judge-backend/internal/app/contest_domain/activity/application/usecase/rejudge/cancel_rejudge"
	getrejudge "judge-backend/internal/app/contest_domain/activity/application/usecase/rejudge/get_rejudge"
	getrejudgesubmissioncomparison "judge-backend/internal/app/contest_domain/activity/application/usecase/rejudge/get_rejudge_submission_comparison"
	listrejudges "judge-backend/internal/app/contest_domain/activity/application/usecase/rejudge/list_rejudges"
	listrejudgesbycontestproject "judge-backend/internal/app/contest_domain/activity/application/usecase/rejudge/list_rejudges_by_contest_project"
	rejectrejudge "judge-backend/internal/app/contest_domain/activity/application/usecase/rejudge/reject_rejudge"
)

type RejudgeApplication struct {
	listRejudges                   listrejudges.Usecase
	listRejudgesByContestProject   listrejudgesbycontestproject.Usecase
	getRejudge                     getrejudge.Usecase
	getRejudgeSubmissionComparison getrejudgesubmissioncomparison.Usecase
	approveRejudge                 approverejudge.Usecase
	rejectRejudge                  rejectrejudge.Usecase
	cancelRejudge                  cancelrejudge.Usecase
}

func NewRejudgeApplication(
	listRejudges listrejudges.Usecase,
	listRejudgesByContestProject listrejudgesbycontestproject.Usecase,
	getRejudge getrejudge.Usecase,
	getRejudgeSubmissionComparison getrejudgesubmissioncomparison.Usecase,
	approveRejudge approverejudge.Usecase,
	rejectRejudge rejectrejudge.Usecase,
	cancelRejudge cancelrejudge.Usecase,
) *RejudgeApplication {
	return &RejudgeApplication{
		listRejudges:                   listRejudges,
		listRejudgesByContestProject:   listRejudgesByContestProject,
		getRejudge:                     getRejudge,
		getRejudgeSubmissionComparison: getRejudgeSubmissionComparison,
		approveRejudge:                 approveRejudge,
		rejectRejudge:                  rejectRejudge,
		cancelRejudge:                  cancelRejudge,
	}
}

func (a RejudgeApplication) ListRejudges(input *listrejudges.Input) (*listrejudges.Output, error) {
	return a.listRejudges.Execute(input)
}

func (a RejudgeApplication) ListRejudgesByContestProject(input *listrejudgesbycontestproject.Input) (*listrejudgesbycontestproject.Output, error) {
	return a.listRejudgesByContestProject.Execute(input)
}

func (a RejudgeApplication) GetRejudge(input *getrejudge.Input) (*getrejudge.Output, error) {
	return a.getRejudge.Execute(input)
}

func (a RejudgeApplication) GetRejudgeSubmissionComparison(input *getrejudgesubmissioncomparison.Input) (*getrejudgesubmissioncomparison.Output, error) {
	return a.getRejudgeSubmissionComparison.Execute(input)
}

func (a RejudgeApplication) ApproveRejudge(input *approverejudge.Input) (*approverejudge.Output, error) {
	return a.approveRejudge.Execute(input)
}

func (a RejudgeApplication) RejectRejudge(input *rejectrejudge.Input) (*rejectrejudge.Output, error) {
	return a.rejectRejudge.Execute(input)
}

func (a RejudgeApplication) CancelRejudge(input *cancelrejudge.Input) (*cancelrejudge.Output, error) {
	return a.cancelRejudge.Execute(input)
}
