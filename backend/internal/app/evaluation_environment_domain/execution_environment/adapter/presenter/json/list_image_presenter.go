package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/adapter/presenter/types"
	listimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/list_image"
)

type ListImagePresenter struct{}

func NewListImagePresenter() types.ListImagePresenterType {
	return &ListImagePresenter{}
}

func (p *ListImagePresenter) Output(output *listimage.Output) api.ListImageResponseObject {
	images := make([]api.AdminImageSpecImageInfo, 0, len(output.Images))
	for _, image := range output.Images {
		images = append(images, api.AdminImageSpecImageInfo{
			Tag:         image.Tag,
			Description: image.Description,
			IsReady:     image.IsReady,
			CreatedAt:   image.CreatedAt,
			UpdatedAt:   image.UpdatedAt,
		})
	}
	return &api.ListImage200JSONResponse{
		Success: true,
		Data:    images,
	}
}

func (p *ListImagePresenter) Error(err error) api.ListImageResponseObject {
	switch {
	case errors.Is(err, listimage.ErrorInternalServer):
		fallthrough
	default:
		return &api.ListImage500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
