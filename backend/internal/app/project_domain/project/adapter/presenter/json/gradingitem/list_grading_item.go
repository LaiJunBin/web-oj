package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/project_domain/project/adapter/presenter/types"
	listgradingitem "judge-backend/internal/app/project_domain/project/application/usecase/grading_item/list_grading_item"
)

type ListGradingItemPresenter types.Presenter[listgradingitem.Output, api.ListGradingItemResponseObject]

type listGradingItemPresenter struct{}

func NewListGradingItemPresenter() ListGradingItemPresenter {
	return &listGradingItemPresenter{}
}

func (p *listGradingItemPresenter) Output(output *listgradingitem.Output) api.ListGradingItemResponseObject {
	gradingItems := make([]api.AdminProjectSpecGradingItemSpecGradingItem, len(output.GradingItems))
	for i, gradingItem := range output.GradingItems {
		gradingItems[i] = api.AdminProjectSpecGradingItemSpecGradingItem{
			Id:             gradingItem.ID,
			Name:           gradingItem.Name,
			Label:          gradingItem.Label,
			Description:    gradingItem.Description,
			Order:          int32(gradingItem.Order),
			Score:          gradingItem.Score,
			TestcaseLength: gradingItem.TestcaseLen,
			CreatedAt:      gradingItem.CreatedAt,
			UpdatedAt:      gradingItem.UpdatedAt,
		}
	}
	return api.ListGradingItem200JSONResponse{
		Success: true,
		Data:    gradingItems,
	}
}

func (p *listGradingItemPresenter) Error(err error) api.ListGradingItemResponseObject {
	switch {
	case errors.Is(err, listgradingitem.ErrorInvalidProjectID):
		return api.ListGradingItem400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
		}
	case errors.Is(err, listgradingitem.ErrorProjectNotFound):
		return api.ListGradingItem404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	case errors.Is(err, listgradingitem.ErrorInternalServer):
		fallthrough
	default:
		return api.ListGradingItem500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
