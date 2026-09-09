package application

import (
	cancelusersubmission "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/cancel_user_submission"
	getusercontest "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/get_user_contest"
	getusercontestschedule "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/get_user_contest_schedule"
	getuserproject "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/get_user_project"
	getuserprojectscoreboard "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/get_user_project_scoreboard"
	listusercontests "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/list_user_contests"
	streamuserprojectclarification "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/stream_user_project_clarification"
	streamuserprojectscoreboard "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/stream_user_project_scoreboard"
	streamuserprojecttime "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/stream_user_project_time"
	streamusersubmission "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/stream_user_submission"
	submituserproject "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/submit_user_project"

	getuserclarification "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/get_user_clarification"
	listuserprojectclarifications "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/list_user_project_clarifications"
	replyclarificationtoadmin "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/reply_clarification_to_admin"
	sendclarificationtoadmim "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/send_clarification_to_admin"

	getusersubmission "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/get_user_submission"
	listuserprojectsubmissions "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/list_user_project_submissions"

	downloaduserprojectasset "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/download_user_project_asset"
)

type ContestantApplication struct {
	getUserContestSchedule   getusercontestschedule.Usecase
	listUserContests         listusercontests.Usecase
	getUserContest           getusercontest.Usecase
	getUserProject           getuserproject.Usecase
	getUserProjectScoreboard getuserprojectscoreboard.Usecase

	listUserProjectClarifications listuserprojectclarifications.Usecase
	getUserClarification          getuserclarification.Usecase
	sendClarificationToAdmin      sendclarificationtoadmim.Usecase
	replyClarificationToAdmin     replyclarificationtoadmin.Usecase

	listUserProjectSubmissions listuserprojectsubmissions.Usecase
	getUserSubmission          getusersubmission.Usecase

	downloadUserProjectAsset downloaduserprojectasset.Usecase

	submitUserProject    submituserproject.Usecase
	cancelUserSubmission cancelusersubmission.Usecase

	streamUserProjectTime          streamuserprojecttime.Usecase
	streamUserProjectScoreboard    streamuserprojectscoreboard.Usecase
	streamUserProjectClarification streamuserprojectclarification.Usecase
	streamUserSubmission           streamusersubmission.Usecase
}

func NewContestantApplication(
	getUserContestSchedule getusercontestschedule.Usecase,
	listUserContests listusercontests.Usecase,
	getUserContest getusercontest.Usecase,
	getUserProject getuserproject.Usecase,
	getUserProjectScoreboard getuserprojectscoreboard.Usecase,

	listUserProjectClarifications listuserprojectclarifications.Usecase,
	getUserClarification getuserclarification.Usecase,
	sendClarificationToAdmin sendclarificationtoadmim.Usecase,
	replyClarificationToAdmin replyclarificationtoadmin.Usecase,

	listUserProjectSubmissions listuserprojectsubmissions.Usecase,
	getUserSubmission getusersubmission.Usecase,

	downloadUserProjectAsset downloaduserprojectasset.Usecase,

	submitUserProject submituserproject.Usecase,
	cancelUserSubmission cancelusersubmission.Usecase,

	streamUserProjectTime streamuserprojecttime.Usecase,
	streamUserProjectScoreboard streamuserprojectscoreboard.Usecase,
	streamUserProjectClarification streamuserprojectclarification.Usecase,
	streamUserSubmission streamusersubmission.Usecase,
) *ContestantApplication {
	return &ContestantApplication{
		getUserContestSchedule:   getUserContestSchedule,
		listUserContests:         listUserContests,
		getUserContest:           getUserContest,
		getUserProject:           getUserProject,
		getUserProjectScoreboard: getUserProjectScoreboard,

		listUserProjectClarifications: listUserProjectClarifications,
		getUserClarification:          getUserClarification,
		sendClarificationToAdmin:      sendClarificationToAdmin,
		replyClarificationToAdmin:     replyClarificationToAdmin,

		listUserProjectSubmissions: listUserProjectSubmissions,
		getUserSubmission:          getUserSubmission,

		downloadUserProjectAsset: downloadUserProjectAsset,

		submitUserProject:    submitUserProject,
		cancelUserSubmission: cancelUserSubmission,

		streamUserProjectTime:          streamUserProjectTime,
		streamUserProjectScoreboard:    streamUserProjectScoreboard,
		streamUserProjectClarification: streamUserProjectClarification,
		streamUserSubmission:           streamUserSubmission,
	}
}

func (a *ContestantApplication) GetUserContestSchedule(input *getusercontestschedule.Input) (*getusercontestschedule.Output, error) {
	return a.getUserContestSchedule.Execute(input)
}

func (a *ContestantApplication) ListUserContests(input *listusercontests.Input) (*listusercontests.Output, error) {
	return a.listUserContests.Execute(input)
}

func (a *ContestantApplication) GetUserContest(input *getusercontest.Input) (*getusercontest.Output, error) {
	return a.getUserContest.Execute(input)
}

func (a *ContestantApplication) GetUserProject(input *getuserproject.Input) (*getuserproject.Output, error) {
	return a.getUserProject.Execute(input)
}

func (a *ContestantApplication) GetUserProjectScoreboard(input *getuserprojectscoreboard.Input) (*getuserprojectscoreboard.Output, error) {
	return a.getUserProjectScoreboard.Execute(input)
}

func (a *ContestantApplication) ListUserProjectClarifications(input *listuserprojectclarifications.Input) (*listuserprojectclarifications.Output, error) {
	return a.listUserProjectClarifications.Execute(input)
}

func (a *ContestantApplication) GetUserClarification(input *getuserclarification.Input) (*getuserclarification.Output, error) {
	return a.getUserClarification.Execute(input)
}

func (a *ContestantApplication) SendClarificationToAdmin(input *sendclarificationtoadmim.Input) (*sendclarificationtoadmim.Output, error) {
	return a.sendClarificationToAdmin.Execute(input)
}

func (a *ContestantApplication) ReplyClarificationToAdmin(input *replyclarificationtoadmin.Input) (*replyclarificationtoadmin.Output, error) {
	return a.replyClarificationToAdmin.Execute(input)
}

func (a *ContestantApplication) ListUserProjectSubmissions(input *listuserprojectsubmissions.Input) (*listuserprojectsubmissions.Output, error) {
	return a.listUserProjectSubmissions.Execute(input)
}

func (a *ContestantApplication) GetUserSubmission(input *getusersubmission.Input) (*getusersubmission.Output, error) {
	return a.getUserSubmission.Execute(input)
}

func (a *ContestantApplication) DownloadUserProjectAsset(input *downloaduserprojectasset.Input) (*downloaduserprojectasset.Output, error) {
	return a.downloadUserProjectAsset.Execute(input)
}

func (a *ContestantApplication) SubmitUserProject(input *submituserproject.Input) (*submituserproject.Output, error) {
	return a.submitUserProject.Execute(input)
}

func (a *ContestantApplication) CancelUserSubmission(input *cancelusersubmission.Input) (*cancelusersubmission.Output, error) {
	return a.cancelUserSubmission.Execute(input)
}

func (a *ContestantApplication) StreamUserProjectTime(input *streamuserprojecttime.Input) (*streamuserprojecttime.Output, error) {
	return a.streamUserProjectTime.Execute(input)
}

func (a *ContestantApplication) StreamUserProjectScoreboard(input *streamuserprojectscoreboard.Input) (*streamuserprojectscoreboard.Output, error) {
	return a.streamUserProjectScoreboard.Execute(input)
}

func (a *ContestantApplication) StreamUserProjectClarification(input *streamuserprojectclarification.Input) (*streamuserprojectclarification.Output, error) {
	return a.streamUserProjectClarification.Execute(input)
}

func (a *ContestantApplication) StreamUserSubmission(input *streamusersubmission.Input) (*streamusersubmission.Output, error) {
	return a.streamUserSubmission.Execute(input)
}
