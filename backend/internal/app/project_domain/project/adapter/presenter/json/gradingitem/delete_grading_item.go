package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/project_domain/project/adapter/presenter/types"
	deletegradingitem "judge-backend/internal/app/project_domain/project/application/usecase/grading_item/delete_grading_item"
)

type DeleteGradingItemPresenter types.Presenter[deletegradingitem.Output, api.DeleteGradingItemResponseObject]

type deleteGradingItemPresenter struct{}

func NewDeleteGradingItemPresenter() DeleteGradingItemPresenter {
	return &deleteGradingItemPresenter{}
}

func (p *deleteGradingItemPresenter) Output(output *deletegradingitem.Output) api.DeleteGradingItemResponseObject {
	return api.DeleteGradingItem200JSONResponse{
		Success: true,
	}
}

func (p *deleteGradingItemPresenter) Error(err error) api.DeleteGradingItemResponseObject {
	switch {
	case errors.Is(err, deletegradingitem.ErrorInvalidID):
		return api.DeleteGradingItem400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
		}
	case errors.Is(err, deletegradingitem.ErrorGradingItemNotFound):
		return api.DeleteGradingItem404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	case errors.Is(err, deletegradingitem.ErrorInternalServer):
		fallthrough
	default:
		return api.DeleteGradingItem500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
