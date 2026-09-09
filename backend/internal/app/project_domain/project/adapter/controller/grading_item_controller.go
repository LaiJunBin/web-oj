package projectcontroller

import (
	"judge-backend/api"
	"judge-backend/internal/app/project_domain/project/adapter/presenter"
	"judge-backend/internal/app/project_domain/project/application"
	creategradingitem "judge-backend/internal/app/project_domain/project/application/usecase/grading_item/create_grading_item"
	deletegradingitem "judge-backend/internal/app/project_domain/project/application/usecase/grading_item/delete_grading_item"
	getgradingitem "judge-backend/internal/app/project_domain/project/application/usecase/grading_item/get_grading_item"
	listgradingitem "judge-backend/internal/app/project_domain/project/application/usecase/grading_item/list_grading_item"
	movegradingitemorder "judge-backend/internal/app/project_domain/project/application/usecase/grading_item/move_grading_item_order"
	updategradingitem "judge-backend/internal/app/project_domain/project/application/usecase/grading_item/update_grading_item"

	"github.com/gin-gonic/gin"
)

type GradingItemController struct {
	app       *application.GradingItemApplication
	presenter *presenter.GradingItemPresenter
}

func NewGradingItemController(
	app *application.GradingItemApplication,
	presenter *presenter.GradingItemPresenter,
) *GradingItemController {
	return &GradingItemController{
		app:       app,
		presenter: presenter,
	}
}

func (ctl *GradingItemController) ListGradingItem(ctx *gin.Context, request api.ListGradingItemRequestObject) (api.ListGradingItemResponseObject, error) {
	input := &listgradingitem.Input{
		ProjectID: request.Id,
	}

	output, err := ctl.app.ListGradingItem(input)
	if err != nil {
		return ctl.presenter.ListGradingItem.Error(err), nil
	}

	return ctl.presenter.ListGradingItem.Output(output), nil
}

func (ctl *GradingItemController) GetGradingItem(ctx *gin.Context, request api.GetGradingItemRequestObject) (api.GetGradingItemResponseObject, error) {
	input := &getgradingitem.Input{
		ID: request.Id,
	}

	output, err := ctl.app.GetGradingItem(input)
	if err != nil {
		return ctl.presenter.GetGradingItem.Error(err), nil
	}

	return ctl.presenter.GetGradingItem.Output(output), nil
}

func (ctl *GradingItemController) CreateGradingItem(ctx *gin.Context, request api.CreateGradingItemRequestObject) (api.CreateGradingItemResponseObject, error) {
	input := &creategradingitem.Input{
		ProjectID:       request.Id,
		Name:            request.Body.Name,
		Score:           request.Body.Score,
		ScoringStrategy: string(request.Body.ScoringStrategy),
		OutputFormat:    string(request.Body.OutputFormat),
		DependTestcases: request.Body.Depends,
	}

	if request.Body.Description != nil {
		input.Description = *request.Body.Description
	}

	if request.Body.ResultParser != nil {
		input.ResultParser = *request.Body.ResultParser
	}

	input.Label = request.Body.Label

	output, err := ctl.app.CreateGradingItem(input)
	if err != nil {
		return ctl.presenter.CreateGradingItem.Error(err), nil
	}

	return ctl.presenter.CreateGradingItem.Output(output), nil
}

func (ctl *GradingItemController) UpdateGradingItem(ctx *gin.Context, request api.UpdateGradingItemRequestObject) (api.UpdateGradingItemResponseObject, error) {
	input := &updategradingitem.Input{
		ID:              request.Id,
		Name:            request.Body.Name,
		Description:     request.Body.Description,
		Score:           request.Body.Score,
		DependTestcases: request.Body.Depends,
		ResultParser:    request.Body.ResultParser,
		Label:           request.Body.Label,
	}

	if request.Body.ScoringStrategy != nil {
		s := string(*request.Body.ScoringStrategy)
		input.ScoringStrategy = &s
	}

	if request.Body.OutputFormat != nil {
		o := string(*request.Body.OutputFormat)
		input.OutputFormat = &o
	}

	output, err := ctl.app.UpdateGradingItem(input)
	if err != nil {
		return ctl.presenter.UpdateGradingItem.Error(err), nil
	}

	return ctl.presenter.UpdateGradingItem.Output(output), nil
}

func (ctl *GradingItemController) DeleteGradingItem(ctx *gin.Context, request api.DeleteGradingItemRequestObject) (api.DeleteGradingItemResponseObject, error) {
	input := &deletegradingitem.Input{
		ID: request.Id,
	}

	output, err := ctl.app.DeleteGradingItem(input)
	if err != nil {
		return ctl.presenter.DeleteGradingItem.Error(err), nil
	}

	return ctl.presenter.DeleteGradingItem.Output(output), nil
}

func (ctl *GradingItemController) MoveGradingItemOrder(ctx *gin.Context, request api.MoveGradingItemOrderRequestObject) (api.MoveGradingItemOrderResponseObject, error) {
	input := &movegradingitemorder.Input{
		ID:    request.Id,
		Order: request.Body.Order,
	}

	output, err := ctl.app.MoveGradingItemOrder(input)
	if err != nil {
		return ctl.presenter.MoveGradingItemOrder.Error(err), nil
	}

	return ctl.presenter.MoveGradingItemOrder.Output(output), nil
}
