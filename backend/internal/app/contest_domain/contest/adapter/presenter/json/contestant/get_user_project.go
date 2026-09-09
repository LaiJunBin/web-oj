package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/contest/adapter/presenter/types"
	getuserproject "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/get_user_project"
	"judge-backend/internal/app/shared/types/httperror"
)

type GetUserProjectPresenter types.Presenter[getuserproject.Output, api.GetUserProjectResponseObject]

type getUserProjectPresenter struct{}

func NewGetUserProjectPresenter() GetUserProjectPresenter {
	return &getUserProjectPresenter{}
}

func (p *getUserProjectPresenter) Output(output *getuserproject.Output) api.GetUserProjectResponseObject {
	gradingItems := make([]api.UserProjectSpecGradingItem, 0, len(output.Project.GradingItems))
	for _, gradingItem := range output.Project.GradingItems {
		publicTestcases := make([]api.UserProjectSpecTestcase, 0, len(gradingItem.PublicTestcases))
		for _, publicTestcase := range gradingItem.PublicTestcases {
			publicTestcases = append(publicTestcases, api.UserProjectSpecTestcase{
				Id:   publicTestcase.ID,
				Name: publicTestcase.Name,
			})
		}

		gradingItems = append(gradingItems, api.UserProjectSpecGradingItem{
			Id:              gradingItem.ID,
			Name:            gradingItem.Name,
			Label:           gradingItem.Label,
			Description:     gradingItem.Description,
			Score:           gradingItem.Score,
			OutputFormat:    api.SharedGradingItemOutputFormat(gradingItem.OutputFormat),
			PublicTestcases: publicTestcases,
		})
	}

	dynamicFields := make([]api.SharedProjectDynamicField, 0, len(output.Project.DynamicFields))
	for _, dynamicField := range output.Project.DynamicFields {
		dynamicFields = append(dynamicFields, api.SharedProjectDynamicField{
			Type:         dynamicField.Type,
			Label:        dynamicField.Label,
			Name:         dynamicField.Name,
			IsRequired:   dynamicField.IsRequired,
			DefaultValue: dynamicField.DefaultValue,
		})
	}

	assets := make([]api.SharedAssetInfo, 0, len(output.Project.Assets))
	for _, asset := range output.Project.Assets {
		assets = append(assets, api.SharedAssetInfo{
			Id:   asset.ID,
			Name: asset.Name,
		})
	}

	return api.GetUserProject200JSONResponse{
		Success: true,
		Data: api.UserProjectSpecProjectDetail{
			Id:            output.Project.ID,
			Name:          output.Project.Name,
			Description:   output.Project.Description,
			TimeLeft:      output.Project.TimeLeft,
			GradingItems:  gradingItems,
			DynamicFields: dynamicFields,
			Assets:        assets,
			IsParticipant: output.Project.IsParticipant,
		},
	}
}

func (p *getUserProjectPresenter) Error(err error) api.GetUserProjectResponseObject {
	return httperror.From(err).Response()
}
