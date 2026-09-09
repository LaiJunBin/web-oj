package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/project_domain/project/adapter/presenter/types"
	gettestcase "judge-backend/internal/app/project_domain/project/application/usecase/testcase/get_testcase"
)

type GetTestcasePresenter types.Presenter[gettestcase.Output, api.GetTestcaseResponseObject]

type getTestcasePresenter struct{}

func NewGetTestcasePresenter() GetTestcasePresenter {
	return &getTestcasePresenter{}
}

func (p *getTestcasePresenter) Output(output *gettestcase.Output) api.GetTestcaseResponseObject {
	depends := make([]api.AdminProjectSpecGradingItemSpecTestcaseSpecDependTestcaseInfo, 0)
	for _, d := range output.Testcase.Dependencies {
		depends = append(depends, api.AdminProjectSpecGradingItemSpecTestcaseSpecDependTestcaseInfo{
			Id:   d.Testcase.ID,
			Name: d.Testcase.Name,
			GradingItem: api.AdminProjectSpecGradingItemSpecTestcaseSpecTestcaseGradingItemInfo{
				Id:    d.GradingItem.ID,
				Label: d.GradingItem.Label,
				Name:  d.GradingItem.Name,
			},
		})
	}

	return api.GetTestcase200JSONResponse{
		Success: true,
		Data: api.AdminProjectSpecGradingItemSpecTestcaseSpecTestcase{
			Id:        output.Testcase.TestcaseInfo.ID,
			Name:      output.Testcase.TestcaseInfo.Name,
			Command:   output.Testcase.Command,
			Score:     output.Testcase.TestcaseInfo.Score,
			IsSample:  output.Testcase.IsSample,
			Depends:   depends,
			CreatedAt: output.Testcase.TestcaseInfo.CreatedAt,
			UpdatedAt: output.Testcase.TestcaseInfo.UpdatedAt,
		},
	}
}

func (p *getTestcasePresenter) Error(err error) api.GetTestcaseResponseObject {
	switch {
	case errors.Is(err, gettestcase.ErrorInvalidID):
		return api.GetTestcase400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
		}
	case errors.Is(err, gettestcase.ErrorTestcaseNotFound):
		return api.GetTestcase404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	case errors.Is(err, gettestcase.ErrorInternalServer):
		fallthrough
	default:
		return api.GetTestcase500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
