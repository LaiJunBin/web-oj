package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/adapter/presenter/types"
	deleteimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/delete_image"
)

type DeleteImagePresenter struct{}

func NewDeleteImagePresenter() types.DeleteImagePresenterType {
	return &DeleteImagePresenter{}
}

func (p *DeleteImagePresenter) Output(output *deleteimage.Output) api.DeleteImageResponseObject {
	return api.DeleteImage200JSONResponse{
		Success: true,
	}
}

func (p *DeleteImagePresenter) Error(err error) api.DeleteImageResponseObject {
	switch {
	case errors.Is(err, deleteimage.ErrorImageNotFound):
		return &api.DeleteImage404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	case errors.Is(err, deleteimage.ErrorInternalServer):
		fallthrough
	default:
		return &api.DeleteImage500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
