package projectcontroller

import (
	"judge-backend/api"
	"judge-backend/internal/app/project_domain/project/adapter/presenter"
	"judge-backend/internal/app/project_domain/project/application"
	createlifecycle "judge-backend/internal/app/project_domain/project/application/usecase/lifecycle/create_lifecycle"
	deletelifecycle "judge-backend/internal/app/project_domain/project/application/usecase/lifecycle/delete_lifecycle"
	getlifecycle "judge-backend/internal/app/project_domain/project/application/usecase/lifecycle/get_lifecycle"
	updatelifecycle "judge-backend/internal/app/project_domain/project/application/usecase/lifecycle/update_lifecycle"

	"github.com/gin-gonic/gin"
)

type LifecycleController struct {
	app       *application.LifecycleApplication
	presenter *presenter.LifecyclePresenter
}

func NewLifecycleController(
	app *application.LifecycleApplication,
	presenter *presenter.LifecyclePresenter,
) *LifecycleController {
	return &LifecycleController{
		app:       app,
		presenter: presenter,
	}
}

func (ctl *LifecycleController) GetLifecycle(ctx *gin.Context, request api.GetLifecycleRequestObject) (api.GetLifecycleResponseObject, error) {
	input := &getlifecycle.Input{
		ID: request.Id,
	}

	output, err := ctl.app.GetLifecycle(input)
	if err != nil {
		return ctl.presenter.GetLifecycle.Error(err), nil
	}

	return ctl.presenter.GetLifecycle.Output(output), nil
}

func (ctl *LifecycleController) CreateLifecycle(ctx *gin.Context, request api.CreateLifecycleRequestObject) (api.CreateLifecycleResponseObject, error) {
	input := &createlifecycle.Input{
		ProjectID:            request.Id,
		Type:                 string(request.Body.Type),
		Name:                 request.Body.Name,
		Command:              request.Body.Command,
		ExecutionEnvironment: string(request.Body.RunsOn),
	}

	if request.Body.Description != nil {
		input.Description = *request.Body.Description
	}

	output, err := ctl.app.CreateLifecycle(input)
	if err != nil {
		return ctl.presenter.CreateLifecycle.Error(err), nil
	}

	return ctl.presenter.CreateLifecycle.Output(output), nil
}

func (ctl *LifecycleController) UpdateLifecycle(ctx *gin.Context, request api.UpdateLifecycleRequestObject) (api.UpdateLifecycleResponseObject, error) {
	input := updatelifecycle.Input{
		ID:          request.Id,
		Name:        request.Body.Name,
		Description: request.Body.Description,
		Command:     request.Body.Command,
	}

	if request.Body.Type != nil {
		t := string(*request.Body.Type)
		input.Type = &t
	}

	if request.Body.RunsOn != nil {
		e := string(*request.Body.RunsOn)
		input.ExecutionEnvironment = &e
	}

	output, err := ctl.app.UpdateLifecycle(&input)
	if err != nil {
		return ctl.presenter.UpdateLifecycle.Error(err), nil
	}

	return ctl.presenter.UpdateLifecycle.Output(output), nil
}

func (ctl *LifecycleController) DeleteLifecycle(ctx *gin.Context, request api.DeleteLifecycleRequestObject) (api.DeleteLifecycleResponseObject, error) {
	input := &deletelifecycle.Input{
		ID: request.Id,
	}

	output, err := ctl.app.DeleteLifecycle(input)
	if err != nil {
		return ctl.presenter.DeleteLifecycle.Error(err), nil
	}

	return ctl.presenter.DeleteLifecycle.Output(output), nil
}
