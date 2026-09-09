package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/contest/adapter/presenter/types"
	getpublicproject "judge-backend/internal/app/contest_domain/contest/application/usecase/public/project/get_public_project"
	"judge-backend/internal/app/shared/types/httperror"
)

type GetPublicProjectPresenter types.Presenter[getpublicproject.Output, api.GetPublicProjectResponseObject]

type getPublicProjectPresenter struct{}

func NewGetPublicProjectPresenter() GetPublicProjectPresenter {
	return &getPublicProjectPresenter{}
}

func (p *getPublicProjectPresenter) Output(output *getpublicproject.Output) api.GetPublicProjectResponseObject {
	gradingItems := make([]api.CommonProjectSpecGradingItem, 0, len(output.Project.GradingItems))
	for _, gradingItem := range output.Project.GradingItems {
		publicTestcases := make([]api.CommonProjectSpecTestcase, 0, len(gradingItem.PublicTestcases))
		for _, publicTestcase := range gradingItem.PublicTestcases {
			publicTestcases = append(publicTestcases, api.CommonProjectSpecTestcase{
				Id:   publicTestcase.ID,
				Name: publicTestcase.Name,
			})
		}

		gradingItems = append(gradingItems, api.CommonProjectSpecGradingItem{
			Id:              gradingItem.ID,
			Name:            gradingItem.Name,
			Label:           gradingItem.Label,
			Description:     gradingItem.Description,
			Score:           gradingItem.Score,
			PublicTestcases: publicTestcases,
		})
	}

	assets := make([]api.SharedAssetInfo, 0, len(output.Project.Assets))
	for _, asset := range output.Project.Assets {
		assets = append(assets, api.SharedAssetInfo{
			Id:   asset.ID,
			Name: asset.Name,
		})
	}

	return api.GetPublicProject200JSONResponse{
		Success: true,
		Data: api.CommonProjectSpecProject{
			Id:           output.Project.ID,
			Name:         output.Project.Name,
			Description:  output.Project.Description,
			Status:       api.CommonProjectSpecProjectStatus(output.Project.Status),
			GradingItems: gradingItems,
			Assets:       assets,
			TimeLeft:     output.Project.TimeLeft,
		},
	}
}

func (p *getPublicProjectPresenter) Error(err error) api.GetPublicProjectResponseObject {
	return httperror.From(err).Response()
}
