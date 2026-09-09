package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/project_domain/project/adapter/presenter/types"
	getgradingitem "judge-backend/internal/app/project_domain/project/application/usecase/grading_item/get_grading_item"
)

type GetGradingItemPresenter types.Presenter[getgradingitem.Output, api.GetGradingItemResponseObject]

type getGradingItemPresenter struct{}

func NewGetGradingItemPresenter() GetGradingItemPresenter {
	return &getGradingItemPresenter{}
}

func (p *getGradingItemPresenter) Output(output *getgradingitem.Output) api.GetGradingItemResponseObject {
	testcases := make([]api.AdminProjectSpecGradingItemSpecTestcaseSpecTestcaseInfo, 0)
	for _, testcase := range output.GradingItem.Testcases {
		testcases = append(testcases, api.AdminProjectSpecGradingItemSpecTestcaseSpecTestcaseInfo{
			Id:        testcase.TestcaseInfo.ID,
			Name:      testcase.TestcaseInfo.Name,
			Score:     testcase.TestcaseInfo.Score,
			Command:   testcase.Command,
			IsSample:  testcase.IsSample,
			CreatedAt: testcase.TestcaseInfo.CreatedAt,
			UpdatedAt: testcase.TestcaseInfo.UpdatedAt,
		})
	}

	depends := make([]api.AdminProjectSpecGradingItemSpecTestcaseSpecDependTestcaseInfo, 0)
	for _, depend := range output.GradingItem.Dependencies {
		depends = append(depends, api.AdminProjectSpecGradingItemSpecTestcaseSpecDependTestcaseInfo{
			Id:   depend.Testcase.ID,
			Name: depend.Testcase.Name,
			GradingItem: api.AdminProjectSpecGradingItemSpecTestcaseSpecTestcaseGradingItemInfo{
				Id:    depend.GradingItem.ID,
				Label: depend.GradingItem.Label,
				Name:  depend.GradingItem.Name,
			},
		})
	}

	return api.GetGradingItem200JSONResponse{
		Success: true,
		Data: api.AdminProjectSpecGradingItemSpecGradingItemDetail{
			Id:              output.GradingItem.ID,
			Name:            output.GradingItem.Name,
			Description:     output.GradingItem.Description,
			Score:           output.GradingItem.Score,
			Order:           int32(output.GradingItem.Order),
			OutputFormat:    output.GradingItem.OutputFormat,
			ScoringStrategy: output.GradingItem.ScoringStrategy,
			ResultParser:    output.GradingItem.ResultParser,
			Label:           output.GradingItem.Label,
			Testcases:       testcases,
			Depends:         depends,
			CreatedAt:       output.GradingItem.CreatedAt,
			UpdatedAt:       output.GradingItem.UpdatedAt,
		},
	}
}

func (p *getGradingItemPresenter) Error(err error) api.GetGradingItemResponseObject {
	switch {
	case errors.Is(err, getgradingitem.ErrorInvalidID):
		return api.GetGradingItem400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
		}
	case errors.Is(err, getgradingitem.ErrorGradingItemNotFound):
		return api.GetGradingItem404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	case errors.Is(err, getgradingitem.ErrorInternalServer):
		fallthrough
	default:
		return api.GetGradingItem500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
