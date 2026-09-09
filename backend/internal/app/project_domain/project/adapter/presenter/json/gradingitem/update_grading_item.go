package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/project_domain/project/adapter/presenter/types"
	updategradingitem "judge-backend/internal/app/project_domain/project/application/usecase/grading_item/update_grading_item"
)

type UpdateGradingItemPresenter types.Presenter[updategradingitem.Output, api.UpdateGradingItemResponseObject]

type updateGradingItemPresenter struct{}

func NewUpdateGradingItemPresenter() UpdateGradingItemPresenter {
	return &updateGradingItemPresenter{}
}

func (p *updateGradingItemPresenter) Output(output *updategradingitem.Output) api.UpdateGradingItemResponseObject {
	return api.UpdateGradingItem200JSONResponse{
		Success: true,
	}
}

func (p *updateGradingItemPresenter) Error(err error) api.UpdateGradingItemResponseObject {
	badRequestErrors := []error{
		updategradingitem.ErrorInvalidID,
		updategradingitem.ErrorEmptyInput,
		updategradingitem.ErrorInvalidName,
		updategradingitem.ErrorEmptyLabel,
		updategradingitem.ErrorInvalidScore,
		updategradingitem.ErrorInvalidScoringStrategy,
		updategradingitem.ErrorInvalidOutputFormat,
		updategradingitem.ErrorInvalidResultParser,
		updategradingitem.ErrorDependTestcaseNotFound,
		updategradingitem.ErrorDependTestcaseIsSameGradingItem,
	}

	for _, e := range badRequestErrors {
		if errors.Is(err, e) {
			return api.UpdateGradingItem400JSONResponse{
				Success: false,
				Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
			}
		}
	}

	switch {
	case errors.Is(err, updategradingitem.ErrorGradingItemNotFound):
		return api.UpdateGradingItem404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	case errors.Is(err, updategradingitem.ErrorInternalServer):
		fallthrough
	default:
		return api.UpdateGradingItem500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
