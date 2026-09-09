package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/project_domain/project/adapter/presenter/types"
	movegradingitemorder "judge-backend/internal/app/project_domain/project/application/usecase/grading_item/move_grading_item_order"
)

type MoveGradingItemOrderPresenter types.Presenter[movegradingitemorder.Output, api.MoveGradingItemOrderResponseObject]

type moveGradingItemOrderPresenter struct{}

func NewMoveGradingItemOrderPresenter() MoveGradingItemOrderPresenter {
	return &moveGradingItemOrderPresenter{}
}

func (p *moveGradingItemOrderPresenter) Output(output *movegradingitemorder.Output) api.MoveGradingItemOrderResponseObject {
	return api.MoveGradingItemOrder200JSONResponse{
		Success: true,
	}
}

func (p *moveGradingItemOrderPresenter) Error(err error) api.MoveGradingItemOrderResponseObject {
	switch {
	case errors.Is(err, movegradingitemorder.ErrorInvalidID):
		fallthrough
	case errors.Is(err, movegradingitemorder.ErrorInvalidOrder):
		return api.MoveGradingItemOrder400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
		}
	case errors.Is(err, movegradingitemorder.ErrorGradingItemNotFound):
		return api.MoveGradingItemOrder404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	case errors.Is(err, movegradingitemorder.ErrorInternalServer):
		fallthrough
	default:
		return api.MoveGradingItemOrder500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
