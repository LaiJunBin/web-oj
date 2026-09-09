package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/adapter/presenter/types"
	updateimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/update_image"
)

type UpdateImagePresenter struct{}

func NewUpdateImagePresenter() types.UpdateImagePresenterType {
	return &UpdateImagePresenter{}
}

func (p *UpdateImagePresenter) Output(output *updateimage.Output) api.UpdateImageResponseObject {
	return &api.UpdateImage200JSONResponse{
		Success: true,
	}
}

func (p *UpdateImagePresenter) Error(err error) api.UpdateImageResponseObject {
	switch {
	case errors.Is(err, updateimage.ErrorInvalidInput):
		fallthrough
	case errors.Is(err, updateimage.ErrorInvalidTag):
		fallthrough
	case errors.Is(err, updateimage.ErrorCurrentBuildInProgress):
		return &api.UpdateImage400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
		}
	case errors.Is(err, updateimage.ErrorImageNotFound):
		return &api.UpdateImage404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	case errors.Is(err, updateimage.ErrorFailedToSetup):
		fallthrough
	case errors.Is(err, updateimage.ErrorInternalServer):
		fallthrough
	default:
		return &api.UpdateImage500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
