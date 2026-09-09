package contestcontroller

import (
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/contest/adapter/presenter"
	"judge-backend/internal/app/contest_domain/contest/application"
	cancelusersubmission "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/cancel_user_submission"
	downloaduserprojectasset "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/download_user_project_asset"
	getuserclarification "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/get_user_clarification"
	getusercontest "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/get_user_contest"
	getusercontestschedule "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/get_user_contest_schedule"
	getuserproject "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/get_user_project"
	getuserprojectscoreboard "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/get_user_project_scoreboard"
	getusersubmission "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/get_user_submission"
	listusercontests "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/list_user_contests"
	listuserprojectclarifications "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/list_user_project_clarifications"
	listuserprojectsubmissions "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/list_user_project_submissions"
	replyclarificationtoadmin "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/reply_clarification_to_admin"
	sendclarificationtoadmin "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/send_clarification_to_admin"
	streamuserprojectclarification "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/stream_user_project_clarification"
	streamuserprojectscoreboard "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/stream_user_project_scoreboard"
	streamuserprojecttime "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/stream_user_project_time"
	streamusersubmission "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/stream_user_submission"
	submituserproject "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/submit_user_project"
	sharedtypes "judge-backend/internal/app/shared/types"
	fileheader "judge-backend/internal/app/shared/types/file_header"
	formrequest "judge-backend/internal/app/shared/types/form_request"

	"github.com/gin-gonic/gin"
)

type ContestantController struct {
	app       *application.ContestantApplication
	presenter *presenter.ContestantPresenter
	logger    sharedtypes.Logger
}

func NewContestantController(
	app *application.ContestantApplication,
	presenter *presenter.ContestantPresenter,
	logger sharedtypes.Logger,
) *ContestantController {
	return &ContestantController{
		app:       app,
		presenter: presenter,
		logger:    logger,
	}
}

func (ctl *ContestantController) userID(ctx *gin.Context) int32 {
	return ctx.MustGet(api.UserIDScopes).(int32)
}

func (ctl *ContestantController) GetUserContestSchedule(ctx *gin.Context, request api.GetUserContestScheduleRequestObject) (api.GetUserContestScheduleResponseObject, error) {
	userID := ctl.userID(ctx)
	ctl.logger.Debugf("Starting to get user contest schedule", "userId", userID, "contestId", request.ContestId)

	input := &getusercontestschedule.Input{
		UserID:    userID,
		ContestID: request.ContestId,
	}

	output, err := ctl.app.GetUserContestSchedule(input)
	if err != nil {
		ctl.logger.Errorf("Failed to get user contest schedule", "contestId", request.ContestId, "error", err)
		return ctl.presenter.GetUserContestSchedule.Error(err), nil
	}

	ctl.logger.Debugf("Successfully got user contest schedule", "contestId", request.ContestId)
	return ctl.presenter.GetUserContestSchedule.Output(output), nil
}

func (ctl *ContestantController) ListUserContests(ctx *gin.Context, request api.ListUserContestsRequestObject) (api.ListUserContestsResponseObject, error) {
	userID := ctl.userID(ctx)
	ctl.logger.Debugf("Starting to list user contests", "userId", userID)

	input := &listusercontests.Input{
		UserID: userID,
	}

	output, err := ctl.app.ListUserContests(input)
	if err != nil {
		ctl.logger.Errorf("Failed to list user contests", "userId", userID, "error", err)
		return ctl.presenter.ListUserContests.Error(err), nil
	}

	ctl.logger.Debugf("Successfully listed user contests", "userId", userID)
	return ctl.presenter.ListUserContests.Output(output), nil
}

func (ctl *ContestantController) GetUserContest(ctx *gin.Context, request api.GetUserContestRequestObject) (api.GetUserContestResponseObject, error) {
	userID := ctl.userID(ctx)
	ctl.logger.Debugf("Starting to get user contest details", "userId", userID, "contestId", request.ContestId)

	input := &getusercontest.Input{
		UserID:    userID,
		ContestID: request.ContestId,
	}

	output, err := ctl.app.GetUserContest(input)
	if err != nil {
		ctl.logger.Errorf("Failed to get user contest details", "userId", userID, "contestId", request.ContestId, "error", err)
		return ctl.presenter.GetUserContest.Error(err), nil
	}

	ctl.logger.Debugf("Successfully got user contest details", "userId", userID, "contestId", request.ContestId)
	return ctl.presenter.GetUserContest.Output(output), nil
}

func (ctl *ContestantController) GetUserProject(ctx *gin.Context, request api.GetUserProjectRequestObject) (api.GetUserProjectResponseObject, error) {
	userID := ctl.userID(ctx)
	ctl.logger.Debugf("Starting to get user project", "userId", userID, "projectId", request.Id)

	input := &getuserproject.Input{
		UserID:           userID,
		ContestProjectID: request.Id,
	}

	output, err := ctl.app.GetUserProject(input)
	if err != nil {
		ctl.logger.Errorf("Failed to get user project", "userId", userID, "projectId", request.Id, "error", err)
		return ctl.presenter.GetUserProject.Error(err), nil
	}

	ctl.logger.Debugf("Successfully got user project", "userId", userID, "projectId", request.Id)
	return ctl.presenter.GetUserProject.Output(output), nil
}

func (ctl *ContestantController) DownloadUserProjectAsset(ctx *gin.Context, request api.DownloadUserProjectAssetRequestObject) (api.DownloadUserProjectAssetResponseObject, error) {
	userID := ctl.userID(ctx)
	ctl.logger.Debugf("Starting to download user project asset", "userId", userID, "projectId", request.Id, "assetId", request.AssetId)

	input := &downloaduserprojectasset.Input{
		UserID:           userID,
		ContestProjectID: request.Id,
		AssetID:          request.AssetId,
	}

	output, err := ctl.app.DownloadUserProjectAsset(input)
	if err != nil {
		ctl.logger.Errorf("Failed to download user project asset", "userId", userID, "projectId", request.Id, "assetId", request.AssetId, "error", err)
		return ctl.presenter.DownloadUserProjectAsset.Error(err), nil
	}

	ctl.logger.Debugf("Successfully downloaded user project asset", "userId", userID, "projectId", request.Id, "assetId", request.AssetId)
	return ctl.presenter.DownloadUserProjectAsset.Output(output), nil
}

func (ctl *ContestantController) ListUserProjectClarifications(ctx *gin.Context, request api.ListUserProjectClarificationsRequestObject) (api.ListUserProjectClarificationsResponseObject, error) {
	userID := ctl.userID(ctx)
	ctl.logger.Debugf("Starting to list user project clarifications", "userId", userID, "projectId", request.Id)

	input := &listuserprojectclarifications.Input{
		UserID:           userID,
		ContestProjectID: request.Id,
	}

	output, err := ctl.app.ListUserProjectClarifications(input)
	if err != nil {
		ctl.logger.Errorf("Failed to list user project clarifications", "userId", userID, "projectId", request.Id, "error", err)
		return ctl.presenter.ListUserProjectClarifications.Error(err), nil
	}

	ctl.logger.Debugf("Successfully listed user project clarifications", "userId", userID, "projectId", request.Id)
	return ctl.presenter.ListUserProjectClarifications.Output(output), nil
}

func (ctl *ContestantController) GetUserClarification(ctx *gin.Context, request api.GetUserClarificationRequestObject) (api.GetUserClarificationResponseObject, error) {
	userID := ctl.userID(ctx)
	ctl.logger.Debugf("Starting to get user clarification details", "userId", userID, "clarificationId", request.ClarificationId)

	input := &getuserclarification.Input{
		UserID:          userID,
		ClarificationID: request.ClarificationId,
	}

	output, err := ctl.app.GetUserClarification(input)
	if err != nil {
		ctl.logger.Errorf("Failed to get user clarification details", "userId", userID, "clarificationId", request.ClarificationId, "error", err)
		return ctl.presenter.GetUserClarification.Error(err), nil
	}

	ctl.logger.Debugf("Successfully got user clarification details", "userId", userID, "clarificationId", request.ClarificationId)
	return ctl.presenter.GetUserClarification.Output(output), nil
}

func (ctl *ContestantController) SendUserProjectClarificationToAdmin(ctx *gin.Context, request api.SendUserProjectClarificationToAdminRequestObject) (api.SendUserProjectClarificationToAdminResponseObject, error) {
	userID := ctl.userID(ctx)
	ctl.logger.Debugf("Starting to send user project clarification to admin", "userId", userID, "projectId", request.Id)

	input := &sendclarificationtoadmin.Input{
		SenderUserID:     userID,
		ContestProjectID: request.Id,
		Subject:          request.Body.Subject,
		Content:          request.Body.Content,
	}

	output, err := ctl.app.SendClarificationToAdmin(input)
	if err != nil {
		ctl.logger.Errorf("Failed to send user project clarification to admin", "userId", userID, "projectId", request.Id, "error", err)
		return ctl.presenter.SendUserProjectClarificationToAdmin.Error(err), nil
	}

	ctl.logger.Debugf("Successfully sent user project clarification to admin", "userId", userID, "projectId", request.Id)
	return ctl.presenter.SendUserProjectClarificationToAdmin.Output(output), nil
}

func (ctl *ContestantController) ReplyUserClarificationToAdmin(ctx *gin.Context, request api.ReplyUserClarificationToAdminRequestObject) (api.ReplyUserClarificationToAdminResponseObject, error) {
	userID := ctl.userID(ctx)
	ctl.logger.Debugf("Starting to reply user clarification to admin", "userId", userID, "clarificationId", request.ClarificationId)

	input := &replyclarificationtoadmin.Input{
		UserID:          userID,
		ClarificationID: request.ClarificationId,
		Content:         request.Body.Content,
	}

	output, err := ctl.app.ReplyClarificationToAdmin(input)
	if err != nil {
		ctl.logger.Errorf("Failed to reply user clarification to admin", "userId", userID, "clarificationId", request.ClarificationId, "error", err)
		return ctl.presenter.ReplyUserProjectClarificationToUser.Error(err), nil
	}

	ctl.logger.Debugf("Successfully replied user clarification to admin", "userId", userID, "clarificationId", request.ClarificationId)
	return ctl.presenter.ReplyUserProjectClarificationToUser.Output(output), nil
}

func (ctl *ContestantController) ListUserProjectSubmissions(ctx *gin.Context, request api.ListUserProjectSubmissionsRequestObject) (api.ListUserProjectSubmissionsResponseObject, error) {
	userID := ctl.userID(ctx)
	ctl.logger.Debugf("Starting to list user project submissions", "userId", userID, "projectId", request.Id)

	input := &listuserprojectsubmissions.Input{
		UserID:           userID,
		ContestProjectID: request.Id,
	}

	output, err := ctl.app.ListUserProjectSubmissions(input)
	if err != nil {
		ctl.logger.Errorf("Failed to list user project submissions", "userId", userID, "projectId", request.Id, "error", err)
		return ctl.presenter.ListUserProjectSubmissions.Error(err), nil
	}

	ctl.logger.Debugf("Successfully listed user project submissions", "userId", userID, "projectId", request.Id)
	return ctl.presenter.ListUserProjectSubmissions.Output(output), nil
}

func (ctl *ContestantController) GetUserSubmission(ctx *gin.Context, request api.GetUserSubmissionRequestObject) (api.GetUserSubmissionResponseObject, error) {
	userID := ctl.userID(ctx)
	ctl.logger.Debugf("Starting to get user submission details", "userId", userID, "submissionId", request.SubmissionId)

	input := &getusersubmission.Input{
		UserID:       userID,
		SubmissionID: request.SubmissionId,
	}

	output, err := ctl.app.GetUserSubmission(input)
	if err != nil {
		ctl.logger.Errorf("Failed to get user submission details", "userId", userID, "submissionId", request.SubmissionId, "error", err)
		return ctl.presenter.GetUserSubmission.Error(err), nil
	}

	ctl.logger.Debugf("Successfully got user submission details", "userId", userID, "submissionId", request.SubmissionId)
	return ctl.presenter.GetUserSubmission.Output(output), nil
}

func (ctl *ContestantController) GetUserProjectScoreboard(ctx *gin.Context, request api.GetUserProjectScoreboardRequestObject) (api.GetUserProjectScoreboardResponseObject, error) {
	userID := ctl.userID(ctx)
	ctl.logger.Debugf("Starting to get user project scoreboard", "userId", userID, "projectId", request.Id)

	input := &getuserprojectscoreboard.Input{
		UserID:           userID,
		ContestProjectID: request.Id,
	}

	output, err := ctl.app.GetUserProjectScoreboard(input)
	if err != nil {
		ctl.logger.Errorf("Failed to get user project scoreboard", "projectId", request.Id, "error", err)
		return ctl.presenter.GetUserProjectScoreboard.Error(err), nil
	}

	ctl.logger.Debugf("Successfully got user project scoreboard", "projectId", request.Id)
	return ctl.presenter.GetUserProjectScoreboard.Output(output), nil
}

type SubmitUserProjectPayload struct {
	Project        *fileheader.FileHeader `form:"project" binding:"required"`
	GradingItemIDs []int32                `form:"gradingItemIDs"`
	DynamicFields  map[string]string      `form:"dynamicFields"`
}

func (ctl *ContestantController) SubmitUserProject(ctx *gin.Context, request api.SubmitUserProjectRequestObject) (api.SubmitUserProjectResponseObject, error) {
	userID := ctl.userID(ctx)
	ctl.logger.Debugf("Starting to submit user project", "userId", userID, "projectId", request.Id)

	payload := &SubmitUserProjectPayload{}
	if err := formrequest.New(request.Body).Bind(payload); err != nil {
		ctl.logger.Errorf("Failed to submit user project: invalid data", "userId", userID, "projectId", request.Id, "error", err)
		return ctl.presenter.SubmitUserProject.Error(submituserproject.ErrorInvalidData), nil
	}

	input := &submituserproject.Input{
		UserID:           userID,
		ContestProjectID: request.Id,
		Project:          payload.Project.Buffer(),
		GradingItemIDs:   payload.GradingItemIDs,
		DynamicFields:    payload.DynamicFields,
	}

	output, err := ctl.app.SubmitUserProject(input)
	if err != nil {
		ctl.logger.Errorf("Failed to submit user project", "userId", userID, "projectId", request.Id, "error", err)
		return ctl.presenter.SubmitUserProject.Error(err), nil
	}

	ctl.logger.Debugf("Successfully submitted user project", "userId", userID, "projectId", request.Id)
	return ctl.presenter.SubmitUserProject.Output(output), nil
}

func (ctl *ContestantController) CancelUserSubmission(ctx *gin.Context, request api.CancelUserSubmissionRequestObject) (api.CancelUserSubmissionResponseObject, error) {
	userID := ctl.userID(ctx)
	ctl.logger.Debugf("Starting to cancel user submission", "userId", userID, "submissionId", request.SubmissionId)

	input := &cancelusersubmission.Input{
		UserID:       userID,
		SubmissionID: request.SubmissionId,
	}

	output, err := ctl.app.CancelUserSubmission(input)
	if err != nil {
		ctl.logger.Errorf("Failed to cancel user submission", "userId", userID, "submissionId", request.SubmissionId, "error", err)
		return ctl.presenter.CancelUserSubmission.Error(err), nil
	}

	ctl.logger.Debugf("Successfully cancelled user submission", "userId", userID, "submissionId", request.SubmissionId)
	return ctl.presenter.CancelUserSubmission.Output(output), nil
}

func (ctl *ContestantController) StreamUserProjectTime(ctx *gin.Context, request api.StreamUserProjectTimeRequestObject) (api.StreamUserProjectTimeResponseObject, error) {
	userID := ctl.userID(ctx)
	ctl.logger.Debugf("Starting to stream user project time", "userId", userID, "projectId", request.Id)

	input := &streamuserprojecttime.Input{
		UserID:           userID,
		ContestProjectID: request.Id,
	}

	output, err := ctl.app.StreamUserProjectTime(input)
	if err != nil {
		ctl.logger.Errorf("Failed to stream user project time", "projectId", request.Id, "error", err)
		return ctl.presenter.StreamUserProjectTime.Error(err), nil
	}

	ctl.logger.Debugf("Successfully streamed user project time", "projectId", request.Id)
	return ctl.presenter.StreamUserProjectTime.Output(ctx.Request.Context(), output), nil
}

func (ctl *ContestantController) StreamUserProjectScoreboard(ctx *gin.Context, request api.StreamUserProjectScoreboardRequestObject) (api.StreamUserProjectScoreboardResponseObject, error) {
	userID := ctl.userID(ctx)
	ctl.logger.Debugf("Starting to stream user project scoreboard", "userId", userID, "projectId", request.Id)

	input := &streamuserprojectscoreboard.Input{
		UserID:           userID,
		ContestProjectID: request.Id,
	}

	output, err := ctl.app.StreamUserProjectScoreboard(input)
	if err != nil {
		ctl.logger.Errorf("Failed to stream user project scoreboard", "projectId", request.Id, "error", err)
		return ctl.presenter.StreamUserProjectScoreboard.Error(err), nil
	}

	ctl.logger.Debugf("Successfully streamed user project scoreboard", "projectId", request.Id)
	return ctl.presenter.StreamUserProjectScoreboard.Output(ctx.Request.Context(), output), nil
}

func (ctl *ContestantController) StreamUserProjectClarification(ctx *gin.Context, request api.StreamUserProjectClarificationRequestObject) (api.StreamUserProjectClarificationResponseObject, error) {
	userID := ctl.userID(ctx)
	ctl.logger.Debugf("Starting to stream user project clarification", "userId", userID, "projectId", request.Id)

	input := &streamuserprojectclarification.Input{
		UserID:           userID,
		ContestProjectID: request.Id,
	}

	output, err := ctl.app.StreamUserProjectClarification(input)
	if err != nil {
		ctl.logger.Errorf("Failed to stream user project clarification", "userId", userID, "projectId", request.Id, "error", err)
		return ctl.presenter.StreamUserProjectClarification.Error(err), nil
	}

	ctl.logger.Debugf("Successfully streamed user project clarification", "userId", userID, "projectId", request.Id)
	return ctl.presenter.StreamUserProjectClarification.Output(ctx.Request.Context(), output), nil
}

func (ctl *ContestantController) StreamUserSubmission(ctx *gin.Context, request api.StreamUserSubmissionRequestObject) (api.StreamUserSubmissionResponseObject, error) {
	userID := ctl.userID(ctx)
	ctl.logger.Debugf("Starting to stream user submission", "userId", userID, "submissionId", request.Id)

	input := &streamusersubmission.Input{
		UserID:       userID,
		SubmissionID: request.Id,
	}

	output, err := ctl.app.StreamUserSubmission(input)
	if err != nil {
		ctl.logger.Errorf("Failed to stream user submission", "userId", userID, "submissionId", request.Id, "error", err)
		return ctl.presenter.StreamUserSubmission.Error(err), nil
	}

	ctl.logger.Debugf("Successfully streamed user submission", "userId", userID, "submissionId", request.Id)
	return ctl.presenter.StreamUserSubmission.Output(ctx.Request.Context(), output), nil
}
