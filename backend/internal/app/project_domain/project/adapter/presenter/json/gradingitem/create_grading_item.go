package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/project_domain/project/adapter/presenter/types"
	creategradingitem "judge-backend/internal/app/project_domain/project/application/usecase/grading_item/create_grading_item"
	"judge-backend/internal/app/project_domain/project/domain/entity/gradingitem"
)

type CreateGradingItemPresenter types.Presenter[creategradingitem.Output, api.CreateGradingItemResponseObject]

type createGradingItemPresenter struct{}

func NewCreateGradingItemPresenter() CreateGradingItemPresenter {
	return &createGradingItemPresenter{}
}

func (p *createGradingItemPresenter) Output(output *creategradingitem.Output) api.CreateGradingItemResponseObject {
	return api.CreateGradingItem200JSONResponse{
		Success: true,
		Data:    output.ID,
	}
}

func (p *createGradingItemPresenter) Error(err error) api.CreateGradingItemResponseObject {
	badRequestErrors := []error{
		gradingitem.ErrorInvalidScoringStrategy,
		gradingitem.ErrorInvalidOutputFormat,

		creategradingitem.ErrorInvalidProjectID,
		creategradingitem.ErrorEmptyName,
		creategradingitem.ErrorEmptyLabel,
		creategradingitem.ErrorInvalidScore,
		creategradingitem.ErrorEmptyScoringStrategy,
		creategradingitem.ErrorEmptyOutputFormat,
		creategradingitem.ErrorInvalidResultParser,
		creategradingitem.ErrorDependTestcaseNotFound,
	}

	for _, e := range badRequestErrors {
		if errors.Is(err, e) {
			return api.CreateGradingItem400JSONResponse{
				Success: false,
				Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
			}
		}
	}

	switch {
	case errors.Is(err, creategradingitem.ErrorProjectNotFound):
		return api.CreateGradingItem404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	case errors.Is(err, creategradingitem.ErrorInternalServer):
		fallthrough
	default:
		return api.CreateGradingItem500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
