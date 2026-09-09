package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/adapter/presenter/types"
	getimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/get_image"
)

type GetImagePresenter struct{}

func NewGetImagePresenter() types.GetImagePresenterType {
	return &GetImagePresenter{}
}
func (p *GetImagePresenter) Output(output *getimage.Output) api.GetImageResponseObject {
	return &api.GetImage200JSONResponse{
		Success: true,
		Data: api.AdminImageSpecImage{
			Tag:         output.Tag,
			MountPath:   output.MountPath,
			Description: output.Description,
			IsReady:     output.IsReady,
			CreatedAt:   output.CreatedAt,
			UpdatedAt:   output.UpdatedAt,
		},
	}
}

func (p *GetImagePresenter) Error(err error) api.GetImageResponseObject {
	switch {
	case errors.Is(err, getimage.ErrorTagIsRequired):
		return &api.GetImage400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
		}
	case errors.Is(err, getimage.ErrorImageNotFound):
		return &api.GetImage404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	case errors.Is(err, getimage.ErrorInternalServer):
		fallthrough
	default:
		return &api.GetImage500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
