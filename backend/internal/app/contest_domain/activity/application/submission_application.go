package application

import (
	getsubmission "judge-backend/internal/app/contest_domain/activity/application/usecase/submission/get_submission"
	getsubmissionsbycontestproject "judge-backend/internal/app/contest_domain/activity/application/usecase/submission/get_submissions_by_contest_project"
	getsubmissionsbyuserid "judge-backend/internal/app/contest_domain/activity/application/usecase/submission/get_submissions_by_user_id"
	ignoresubmission "judge-backend/internal/app/contest_domain/activity/application/usecase/submission/ignore_submission"
	rejudgesubmission "judge-backend/internal/app/contest_domain/activity/application/usecase/submission/rejudge_submission"
	streamsubmissionsbycontestproject "judge-backend/internal/app/contest_domain/activity/application/usecase/submission/stream_submissions_by_contest_project"
	unignoresubmission "judge-backend/internal/app/contest_domain/activity/application/usecase/submission/unignore_submission"
)

type SubmissionApplication struct {
	getSubmissionsByContestProject    getsubmissionsbycontestproject.Usecase
	streamSubmissionsByContestProject streamsubmissionsbycontestproject.Usecase
	getSubmission                     getsubmission.Usecase
	getSubmissionsByUserID            getsubmissionsbyuserid.Usecase
	rejudgeSubmission                 rejudgesubmission.Usecase
	ignoreSubmission                  ignoresubmission.Usecase
	unignoreSubmission                unignoresubmission.Usecase
}

func NewSubmissionApplication(
	getSubmissionsByContestProject getsubmissionsbycontestproject.Usecase,
	streamSubmissionsByContestProject streamsubmissionsbycontestproject.Usecase,
	getSubmission getsubmission.Usecase,
	getSubmissionsByUserID getsubmissionsbyuserid.Usecase,
	rejudgeSubmission rejudgesubmission.Usecase,
	ignoreSubmission ignoresubmission.Usecase,
	unignoreSubmission unignoresubmission.Usecase,
) *SubmissionApplication {
	return &SubmissionApplication{
		getSubmissionsByContestProject:    getSubmissionsByContestProject,
		streamSubmissionsByContestProject: streamSubmissionsByContestProject,
		getSubmission:                     getSubmission,
		getSubmissionsByUserID:            getSubmissionsByUserID,
		rejudgeSubmission:                 rejudgeSubmission,
		ignoreSubmission:                  ignoreSubmission,
		unignoreSubmission:                unignoreSubmission,
	}
}

func (a SubmissionApplication) GetSubmissionsByContestProject(input *getsubmissionsbycontestproject.Input) (*getsubmissionsbycontestproject.Output, error) {
	return a.getSubmissionsByContestProject.Execute(input)
}

func (a SubmissionApplication) StreamSubmissionsByContestProject(input *streamsubmissionsbycontestproject.Input) (*streamsubmissionsbycontestproject.Output, error) {
	return a.streamSubmissionsByContestProject.Execute(input)
}

func (a SubmissionApplication) GetSubmission(input *getsubmission.Input) (*getsubmission.Output, error) {
	return a.getSubmission.Execute(input)
}

func (a SubmissionApplication) GetSubmissionsByUserID(input *getsubmissionsbyuserid.Input) (*getsubmissionsbyuserid.Output, error) {
	return a.getSubmissionsByUserID.Execute(input)
}

func (a SubmissionApplication) RejudgeSubmission(input *rejudgesubmission.Input) (*rejudgesubmission.Output, error) {
	return a.rejudgeSubmission.Execute(input)
}

func (a SubmissionApplication) IgnoreSubmission(input *ignoresubmission.Input) (*ignoresubmission.Output, error) {
	return a.ignoreSubmission.Execute(input)
}

func (a SubmissionApplication) UnignoreSubmission(input *unignoresubmission.Input) (*unignoresubmission.Output, error) {
	return a.unignoreSubmission.Execute(input)
}
