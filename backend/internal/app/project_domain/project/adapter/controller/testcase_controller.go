package projectcontroller

import (
	"judge-backend/api"
	"judge-backend/internal/app/project_domain/project/adapter/presenter"
	"judge-backend/internal/app/project_domain/project/application"
	createtestcase "judge-backend/internal/app/project_domain/project/application/usecase/testcase/create_testcase"
	deletetestcase "judge-backend/internal/app/project_domain/project/application/usecase/testcase/delete_testcase"
	gettestcase "judge-backend/internal/app/project_domain/project/application/usecase/testcase/get_testcase"
	listprojecttestcases "judge-backend/internal/app/project_domain/project/application/usecase/testcase/list_project_testcases"
	updatetestcase "judge-backend/internal/app/project_domain/project/application/usecase/testcase/update_testcase"

	"github.com/gin-gonic/gin"
)

type TestcaseController struct {
	app       *application.TestcaseApplication
	presenter *presenter.TestcasePresenter
}

func NewTestcaseController(
	app *application.TestcaseApplication,
	presenter *presenter.TestcasePresenter,
) *TestcaseController {
	return &TestcaseController{
		app:       app,
		presenter: presenter,
	}
}

func (s *TestcaseController) ListProjectTestcases(ctx *gin.Context, request api.ListProjectTestcasesRequestObject) (api.ListProjectTestcasesResponseObject, error) {
	input := &listprojecttestcases.Input{
		ProjectID: request.Id,
	}

	output, err := s.app.ListProjectTestcases(input)
	if err != nil {
		return s.presenter.ListProjectTestcases.Error(err), nil
	}

	return s.presenter.ListProjectTestcases.Output(output), nil
}

func (ctl *TestcaseController) GetTestcase(ctx *gin.Context, request api.GetTestcaseRequestObject) (api.GetTestcaseResponseObject, error) {
	input := &gettestcase.Input{
		ID: request.Id,
	}

	output, err := ctl.app.GetTestcase(input)
	if err != nil {
		return ctl.presenter.GetTestcase.Error(err), nil
	}

	return ctl.presenter.GetTestcase.Output(output), nil
}

func (ctl *TestcaseController) CreateTestcase(ctx *gin.Context, request api.CreateTestcaseRequestObject) (api.CreateTestcaseResponseObject, error) {
	input := &createtestcase.Input{
		GradingItemID: request.Id,
		Name:          request.Body.Name,
		Command:       request.Body.Command,
		Score:         request.Body.Score,
		IsSample:      request.Body.IsSample,
		Depends:       request.Body.Depends,
	}

	output, err := ctl.app.CreateTestcase(input)
	if err != nil {
		return ctl.presenter.CreateTestcase.Error(err), nil
	}

	return ctl.presenter.CreateTestcase.Output(output), nil
}

func (ctl *TestcaseController) UpdateTestcase(ctx *gin.Context, request api.UpdateTestcaseRequestObject) (api.UpdateTestcaseResponseObject, error) {
	input := &updatetestcase.Input{
		ID:       request.Id,
		Name:     request.Body.Name,
		Command:  request.Body.Command,
		Score:    request.Body.Score,
		IsSample: request.Body.IsSample,
		Depends:  request.Body.Depends,
	}

	output, err := ctl.app.UpdateTestcase(input)
	if err != nil {
		return ctl.presenter.UpdateTestcase.Error(err), nil
	}

	return ctl.presenter.UpdateTestcase.Output(output), nil
}

func (ctl *TestcaseController) DeleteTestcase(ctx *gin.Context, request api.DeleteTestcaseRequestObject) (api.DeleteTestcaseResponseObject, error) {
	input := &deletetestcase.Input{
		ID: request.Id,
	}

	output, err := ctl.app.DeleteTestcase(input)
	if err != nil {
		return ctl.presenter.DeleteTestcase.Error(err), nil
	}

	return ctl.presenter.DeleteTestcase.Output(output), nil
}
