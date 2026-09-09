package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/adapter/presenter/types"
	createimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/create_image"
)

type CreateImagePresenter struct{}

func NewCreateImagePresenter() types.CreateImagePresenterType {
	return &CreateImagePresenter{}
}

func (p *CreateImagePresenter) Output(output *createimage.Output) api.CreateImageResponseObject {
	return &api.CreateImage200JSONResponse{
		Success: true,
		Data:    output.Tag,
	}
}

func (p *CreateImagePresenter) Error(err error) api.CreateImageResponseObject {
	switch {
	case errors.Is(err, createimage.ErrorInvalidInput):
		fallthrough
	case errors.Is(err, createimage.ErrorTagIsRequired):
		fallthrough
	case errors.Is(err, createimage.ErrorProjectZipIsRequired):
		fallthrough
	case errors.Is(err, createimage.ErrorInvalidTag):
		fallthrough
	case errors.Is(err, createimage.ErrorInvalidMountPath):
		fallthrough
	case errors.Is(err, createimage.ErrorImageAlreadyExists):
		return &api.CreateImage400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
		}
	case errors.Is(err, createimage.ErrorFailedToSetup):
		fallthrough
	case errors.Is(err, createimage.ErrorInternalServer):
		fallthrough
	default:
		return &api.CreateImage500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
