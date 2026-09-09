package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/project_domain/project/adapter/presenter/types"
	"judge-backend/internal/app/project_domain/project/application/query/dto"
	getproject "judge-backend/internal/app/project_domain/project/application/usecase/project/get_project"
)

type GetProjectPresenter types.Presenter[getproject.Output, api.GetProjectResponseObject]

type getProjectPresenter struct{}

func NewGetProjectPresenter() GetProjectPresenter {
	return &getProjectPresenter{}
}

func (p *getProjectPresenter) Output(output *getproject.Output) api.GetProjectResponseObject {
	image := p.transformImage(output.Project.Image)
	testImage := p.transformImage(output.Project.TestImage)
	assets := p.transformAssets(output.Project.Assets)
	dynamicFields := p.transformDynamicFields(output.Project.DynamicFields)
	lifecycle := p.transformLifecycle(output.Project.Lifecycle)
	testLifecycle := p.transformLifecycle(output.Project.TestLifecycle)

	return api.GetProject200JSONResponse{
		Success: true,
		Data: api.AdminProjectSpecProjectDetail{
			Id:            output.Project.ID,
			Name:          output.Project.Name,
			Description:   output.Project.Description,
			Image:         image,
			TestImage:     testImage,
			Env:           output.Project.Env,
			TestEnv:       output.Project.TestEnv,
			ResultParser:  output.Project.ResultParser,
			Assets:        assets,
			DynamicFields: dynamicFields,
			Lifecycle:     *lifecycle,
			TestLifecycle: *testLifecycle,
			CreatedAt:     output.Project.CreatedAt,
			UpdatedAt:     output.Project.UpdatedAt,
		},
	}
}

func (p *getProjectPresenter) Error(err error) api.GetProjectResponseObject {
	switch {
	case errors.Is(err, getproject.ErrorInvalidID):
		return api.GetProject400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
		}
	case errors.Is(err, getproject.ErrorProjectNotFound):
		return api.GetProject404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	case errors.Is(err, getproject.ErrorInternalServer):
		fallthrough
	default:
		return api.GetProject500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}

func (p *getProjectPresenter) transformImage(image *dto.Image) *api.AdminImageSpecImageInfo {
	if image == nil {
		return nil
	}
	return &api.AdminImageSpecImageInfo{
		Tag:         image.Tag,
		Description: image.Description,
		IsReady:     image.IsReady,
	}
}

func (p *getProjectPresenter) transformAssets(assets []*dto.Asset) []api.SharedAsset {
	result := make([]api.SharedAsset, 0, len(assets))
	for _, asset := range assets {
		result = append(result, api.SharedAsset{
			Id:           asset.ID,
			Name:         asset.Name,
			Size:         asset.Size,
			IsEnabled:    asset.Enabled,
			LastModified: asset.LastModified,
		})
	}
	return result
}

func (p *getProjectPresenter) transformDynamicFields(dynamicFields []*dto.DynamicField) []api.SharedProjectDynamicField {
	result := make([]api.SharedProjectDynamicField, 0, len(dynamicFields))
	for _, dynamicField := range dynamicFields {
		result = append(result, api.SharedProjectDynamicField{
			Type:         dynamicField.Type,
			Name:         dynamicField.Name,
			Label:        dynamicField.Label,
			DefaultValue: dynamicField.Default,
			IsRequired:   dynamicField.Required,
		})
	}
	return result
}

func (p *getProjectPresenter) transformLifecycle(lifecycle dto.Lifecycle) *api.AdminProjectSpecImageLifecycle {
	result := &api.AdminProjectSpecImageLifecycle{
		BeforeAll:  []api.AdminProjectSpecLifecycleSpecLifecycle{},
		BeforeEach: []api.AdminProjectSpecLifecycleSpecLifecycle{},
		AfterEach:  []api.AdminProjectSpecLifecycleSpecLifecycle{},
		AfterAll:   []api.AdminProjectSpecLifecycleSpecLifecycle{},
	}

	for _, command := range lifecycle.BeforeAll {
		result.BeforeAll = append(result.BeforeAll, api.AdminProjectSpecLifecycleSpecLifecycle{
			Id:          command.ID,
			Name:        command.Name,
			Description: command.Description,
			Command:     command.Text,
		})
	}

	for _, command := range lifecycle.BeforeEach {
		result.BeforeEach = append(result.BeforeEach, api.AdminProjectSpecLifecycleSpecLifecycle{
			Id:          command.ID,
			Name:        command.Name,
			Description: command.Description,
			Command:     command.Text,
		})
	}

	for _, command := range lifecycle.AfterEach {
		result.AfterEach = append(result.AfterEach, api.AdminProjectSpecLifecycleSpecLifecycle{
			Id:          command.ID,
			Name:        command.Name,
			Description: command.Description,
			Command:     command.Text,
		})
	}

	for _, command := range lifecycle.AfterAll {
		result.AfterAll = append(result.AfterAll, api.AdminProjectSpecLifecycleSpecLifecycle{
			Id:          command.ID,
			Name:        command.Name,
			Description: command.Description,
			Command:     command.Text,
		})
	}

	return result
}
