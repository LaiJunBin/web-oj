package contestcontroller

import (
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/contest/adapter/presenter"
	"judge-backend/internal/app/contest_domain/contest/application"
	"judge-backend/internal/app/contest_domain/contest/application/usecase/public/common/getsessionttl"
	getpubliccontest "judge-backend/internal/app/contest_domain/contest/application/usecase/public/contest/get_public_contest"
	getpubliccontestschedule "judge-backend/internal/app/contest_domain/contest/application/usecase/public/contest/get_public_contest_schedule"
	listpubliccontests "judge-backend/internal/app/contest_domain/contest/application/usecase/public/contest/list_public_contests"
	getpublicproject "judge-backend/internal/app/contest_domain/contest/application/usecase/public/project/get_public_project"
	getpublicprojectscoreboard "judge-backend/internal/app/contest_domain/contest/application/usecase/public/project/get_public_project_scoreboard"
	streampublicprojectscoreboard "judge-backend/internal/app/contest_domain/contest/application/usecase/public/project/stream_public_project_scoreboard"
	streampublicprojecttime "judge-backend/internal/app/contest_domain/contest/application/usecase/public/project/stream_public_project_time"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/httperror"
	"judge-backend/internal/config"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"
)

type PublicController struct {
	app       *application.PublicApplication
	presenter *presenter.PublicPresenter
	config    *config.Config
	logger    sharedtypes.Logger
}

func NewPublicController(
	app *application.PublicApplication,
	presenter *presenter.PublicPresenter,
	config *config.Config,
	logger sharedtypes.Logger,
) *PublicController {
	return &PublicController{
		app:       app,
		presenter: presenter,
		config:    config,
		logger:    logger,
	}
}

func (ctl *PublicController) GetCsrfToken(ctx *gin.Context, request api.GetCsrfTokenRequestObject) (api.GetCsrfTokenResponseObject, error) {
	ctl.logger.Debugf("Starting to get CSRF token")

	sessions.Default(ctx).Options(sessions.Options{
		SameSite: http.SameSiteStrictMode,
		HttpOnly: true,
		// Match the auth cookies' Secure attribute so the CSRF session cookie is
		// not the one credential still sent in cleartext over HTTP.
		Secure: ctl.config.Auth.Cookie.Secure,
		Path:   "/",
		MaxAge: 60 * 30,
	})
	token := csrf.GetToken(ctx)

	input := &getsessionttl.Input{
		SessionID: sessions.Default(ctx).ID(),
	}

	output, err := ctl.app.GetSessionTtl(input)
	if err != nil {
		ctl.logger.Errorf("Failed to get CSRF token", "error", err)
		return httperror.From(err).Response(), nil
	}

	ctl.logger.Debugf("Successfully got CSRF token")
	return api.GetCsrfToken200JSONResponse{
		Success: true,
		Data: api.CommonCsrfSpecCsrfToken{
			Token: token,
			Ttl:   int32(output.TTL),
		},
	}, nil
}

func (ctl *PublicController) DeleteCsrfToken(ctx *gin.Context, request api.DeleteCsrfTokenRequestObject) (api.DeleteCsrfTokenResponseObject, error) {
	ctl.logger.Debugf("Starting to delete CSRF token")

	session := sessions.Default(ctx)
	session.Clear()
	session.Save()
	ctx.SetCookie("csrf", "", -1, "/", "", false, true)

	ctl.logger.Debugf("Successfully deleted CSRF token")
	return api.DeleteCsrfToken200JSONResponse{
		Success: true,
	}, nil
}

func (ctl *PublicController) GetPublicContestSchedule(ctx *gin.Context, request api.GetPublicContestScheduleRequestObject) (api.GetPublicContestScheduleResponseObject, error) {
	ctl.logger.Debugf("Starting to get public contest schedule", "contestId", request.ContestId)

	input := &getpubliccontestschedule.Input{
		ContestID: request.ContestId,
	}

	output, err := ctl.app.GetPublicContestSchedule(input)
	if err != nil {
		ctl.logger.Errorf("Failed to get public contest schedule", "contestId", request.ContestId, "error", err)
		return ctl.presenter.GetPublicContestSchedule.Error(err), nil
	}

	ctl.logger.Debugf("Successfully got public contest schedule", "contestId", request.ContestId)
	return ctl.presenter.GetPublicContestSchedule.Output(output), nil
}

func (ctl *PublicController) ListPublicContests(ctx *gin.Context, request api.ListPublicContestsRequestObject) (api.ListPublicContestsResponseObject, error) {
	ctl.logger.Debugf("Starting to list public contests")

	input := &listpubliccontests.Input{}

	output, err := ctl.app.ListPublicContests(input)
	if err != nil {
		ctl.logger.Errorf("Failed to list public contests", "error", err)
		return ctl.presenter.ListPublicContests.Error(err), nil
	}

	ctl.logger.Debugf("Successfully listed public contests")
	return ctl.presenter.ListPublicContests.Output(output), nil
}

func (ctl *PublicController) GetPublicContest(ctx *gin.Context, request api.GetPublicContestRequestObject) (api.GetPublicContestResponseObject, error) {
	ctl.logger.Debugf("Starting to get public contest details", "contestId", request.ContestId)

	input := &getpubliccontest.Input{
		ContestID: request.ContestId,
	}

	output, err := ctl.app.GetPublicContest(input)
	if err != nil {
		ctl.logger.Errorf("Failed to get public contest details", "contestId", request.ContestId, "error", err)
		return ctl.presenter.GetPublicContest.Error(err), nil
	}

	ctl.logger.Debugf("Successfully got public contest details", "contestId", request.ContestId)
	return ctl.presenter.GetPublicContest.Output(output), nil
}

func (ctl *PublicController) GetPublicProject(ctx *gin.Context, request api.GetPublicProjectRequestObject) (api.GetPublicProjectResponseObject, error) {
	ctl.logger.Debugf("Starting to get public project", "projectId", request.Id)

	input := &getpublicproject.Input{
		ContestProjectID: request.Id,
	}

	output, err := ctl.app.GetPublicProject(input)
	if err != nil {
		ctl.logger.Errorf("Failed to get public project", "projectId", request.Id, "error", err)
		return ctl.presenter.GetPublicProject.Error(err), nil
	}

	ctl.logger.Debugf("Successfully got public project", "projectId", request.Id)
	return ctl.presenter.GetPublicProject.Output(output), nil
}

func (ctl *PublicController) GetPublicProjectScoreboard(ctx *gin.Context, request api.GetPublicProjectScoreboardRequestObject) (api.GetPublicProjectScoreboardResponseObject, error) {
	ctl.logger.Debugf("Starting to get public project scoreboard", "projectId", request.Id)

	input := &getpublicprojectscoreboard.Input{
		ContestProjectID: request.Id,
	}

	output, err := ctl.app.GetPublicProjectScoreboard(input)
	if err != nil {
		ctl.logger.Errorf("Failed to get public project scoreboard", "projectId", request.Id, "error", err)
		return ctl.presenter.GetPublicProjectScoreboard.Error(err), nil
	}

	ctl.logger.Debugf("Successfully got public project scoreboard", "projectId", request.Id)
	return ctl.presenter.GetPublicProjectScoreboard.Output(output), nil
}

func (ctl *PublicController) DownloadPublicProjectAsset(ctx *gin.Context, request api.DownloadPublicProjectAssetRequestObject) (api.DownloadPublicProjectAssetResponseObject, error) {
	ctl.logger.Debugf("Attempting to download public project asset without login", "projectId", request.Id, "assetId", request.AssetId)

	ctl.logger.Debugf("Download public project asset failed: authentication required", "projectId", request.Id, "assetId", request.AssetId)
	return api.DownloadPublicProjectAsset401JSONResponse{
		Success: false,
		Message: "Please login to download the asset",
	}, nil
}

func (ctl *PublicController) StreamPublicProjectScoreboard(ctx *gin.Context, request api.StreamPublicProjectScoreboardRequestObject) (api.StreamPublicProjectScoreboardResponseObject, error) {
	ctl.logger.Debugf("Starting to stream public project scoreboard", "projectId", request.Id)

	input := &streampublicprojectscoreboard.Input{
		ContestProjectID: request.Id,
	}

	output, err := ctl.app.StreamPublicProjectScoreboard(input)
	if err != nil {
		ctl.logger.Errorf("Failed to stream public project scoreboard", "projectId", request.Id, "error", err)
		return ctl.presenter.StreamPublicProjectScoreboard.Error(err), nil
	}

	ctl.logger.Debugf("Successfully streamed public project scoreboard", "projectId", request.Id)
	return ctl.presenter.StreamPublicProjectScoreboard.Output(ctx.Request.Context(), output), nil
}

func (ctl *PublicController) StreamPublicProjectTime(ctx *gin.Context, request api.StreamPublicProjectTimeRequestObject) (api.StreamPublicProjectTimeResponseObject, error) {
	ctl.logger.Debugf("Starting to stream public project time", "projectId", request.Id)

	input := &streampublicprojecttime.Input{
		ContestProjectID: request.Id,
	}

	output, err := ctl.app.StreamPublicProjectTime(input)
	if err != nil {
		ctl.logger.Errorf("Failed to stream public project time", "projectId", request.Id, "error", err)
		return ctl.presenter.StreamPublicProjectTime.Error(err), nil
	}

	ctl.logger.Debugf("Successfully streamed public project time", "projectId", request.Id)
	return ctl.presenter.StreamPublicProjectTime.Output(ctx.Request.Context(), output), nil
}
