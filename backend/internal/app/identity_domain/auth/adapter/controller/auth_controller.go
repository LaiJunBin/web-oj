package controller

import (
	"judge-backend/api"
	"judge-backend/internal/app/identity_domain/auth/adapter/cookie"
	"judge-backend/internal/app/identity_domain/auth/adapter/presenter"
	"judge-backend/internal/app/identity_domain/auth/application"
	"judge-backend/internal/app/identity_domain/auth/application/usecase"
	sharedtypes "judge-backend/internal/app/shared/types"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	presenter *presenter.AuthPresenter
	app       *application.AuthApplication
	cookie    *cookie.Codec
	logger    sharedtypes.Logger
}

func NewAuthController(
	presenter *presenter.AuthPresenter,
	app *application.AuthApplication,
	cookie *cookie.Codec,
	logger sharedtypes.Logger,
) *AuthController {
	return &AuthController{
		presenter: presenter,
		app:       app,
		cookie:    cookie,
		logger:    logger,
	}
}

func (ctl *AuthController) Login(ctx *gin.Context, request api.LoginRequestObject) (api.LoginResponseObject, error) {
	ctl.logger.Infof("Login request received for username: %s", request.Body.Username)

	input := &usecase.LoginInput{
		Username: request.Body.Username,
		Password: request.Body.Password,
	}

	output, err := ctl.app.Login(input)
	if err != nil {
		ctl.logger.Errorf("Login failed for username: %s, error: %v", request.Body.Username, err)
		return ctl.presenter.Login.Error(err), nil
	}

	ctl.logger.Infof("Login successful for username: %s", request.Body.Username)

	ctl.cookie.SetAccessToken(ctx, output.AccessToken.Value, int(output.AccessToken.MaxAge))
	ctl.cookie.SetRefreshToken(ctx, output.RefreshToken.Value, int(output.RefreshToken.MaxAge))

	return ctl.presenter.Login.Output(output), nil
}

func (ctl *AuthController) Logout(ctx *gin.Context, request api.LogoutRequestObject) (api.LogoutResponseObject, error) {
	ctl.logger.Infof("Logout request received")

	refreshToken := ctx.MustGet("refresh_token").(string)
	input := &usecase.LogoutInput{
		RefreshToken: refreshToken,
	}

	ctl.cookie.Clear(ctx)

	output, err := ctl.app.Logout(input)
	if err != nil {
		ctl.logger.Errorf("Logout failed, error: %v", err)
		return ctl.presenter.Logout.Error(err), nil
	}

	ctl.logger.Infof("Logout successful")

	return ctl.presenter.Logout.Output(output), nil
}

func (ctl *AuthController) GetProfile(ctx *gin.Context, request api.GetProfileRequestObject) (api.GetProfileResponseObject, error) {
	ctl.logger.Debugf("GetProfile request received")

	accessToken := ctx.MustGet("access_token").(string)
	input := &usecase.GetProfileInput{
		AccessToken: accessToken,
	}

	output, err := ctl.app.GetProfile(input)
	if err != nil {
		ctl.logger.Errorf("GetProfile failed, error: %v", err)
		return ctl.presenter.GetProfile.Error(err), nil
	}

	ctl.logger.Debugf("GetProfile successful")

	return ctl.presenter.GetProfile.Output(output), nil
}

func (ctl *AuthController) RefreshToken(ctx *gin.Context, request api.RefreshTokenRequestObject) (api.RefreshTokenResponseObject, error) {
	ctl.logger.Debugf("RefreshToken request received")

	refreshToken := ctx.MustGet("refresh_token").(string)
	input := &usecase.RefreshTokenInput{
		RefreshToken: refreshToken,
	}

	output, err := ctl.app.RefreshToken(input)
	if err != nil {
		ctl.logger.Errorf("RefreshToken failed, error: %v", err)
		ctl.cookie.Clear(ctx)
		return ctl.presenter.RefreshToken.Error(err), nil
	}

	ctl.logger.Debugf("RefreshToken successful")

	ctl.cookie.SetAccessToken(ctx, output.AccessToken.Value, int(output.AccessToken.MaxAge))
	ctl.cookie.SetRefreshToken(ctx, output.RefreshToken.Value, int(output.RefreshToken.MaxAge))

	return ctl.presenter.RefreshToken.Output(output), nil
}
