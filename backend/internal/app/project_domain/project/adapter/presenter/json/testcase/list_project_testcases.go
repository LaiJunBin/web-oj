package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/project_domain/project/adapter/presenter/types"
	listprojecttestcases "judge-backend/internal/app/project_domain/project/application/usecase/testcase/list_project_testcases"
	"judge-backend/internal/app/shared/types/httperror"
)

type ListProjectTestcasesPresenter types.Presenter[listprojecttestcases.Output, api.ListProjectTestcasesResponseObject]

type listProjectTestcasesPresenter struct{}

func NewListProjectTestcasesPresenter() ListProjectTestcasesPresenter {
	return &listProjectTestcasesPresenter{}
}

func (p *listProjectTestcasesPresenter) Output(output *listprojecttestcases.Output) api.ListProjectTestcasesResponseObject {
	gradingItems := make([]api.AdminProjectSpecGradingItemSpecTestcaseSpecProjectTestcaseGradingItem, 0)
	for _, gradingItem := range output.GradingItems {
		testcases := make([]api.AdminProjectSpecGradingItemSpecTestcaseSpecProjectTestcase, 0)
		for _, testcase := range gradingItem.Testcases {
			testcases = append(testcases, api.AdminProjectSpecGradingItemSpecTestcaseSpecProjectTestcase{
				Id:   testcase.ID,
				Name: testcase.Name,
			})
		}

		gradingItems = append(gradingItems, api.AdminProjectSpecGradingItemSpecTestcaseSpecProjectTestcaseGradingItem{
			Id:        gradingItem.ID,
			Label:     gradingItem.Label,
			Name:      gradingItem.Name,
			Testcases: testcases,
		})
	}

	return api.ListProjectTestcases200JSONResponse{
		Success: true,
		Data:    gradingItems,
	}
}

func (p *listProjectTestcasesPresenter) Error(err error) api.ListProjectTestcasesResponseObject {
	return httperror.From(err).Response()
}
