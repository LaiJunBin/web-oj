package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/project_domain/project/adapter/presenter/types"
	createtestcase "judge-backend/internal/app/project_domain/project/application/usecase/testcase/create_testcase"
)

type CreateTestcasePresenter types.Presenter[createtestcase.Output, api.CreateTestcaseResponseObject]

type createTestcasePresenter struct{}

func NewCreateTestcasePresenter() CreateTestcasePresenter {
	return &createTestcasePresenter{}
}

func (p *createTestcasePresenter) Output(output *createtestcase.Output) api.CreateTestcaseResponseObject {
	return api.CreateTestcase200JSONResponse{
		Success: true,
		Data:    output.ID,
	}
}

func (p *createTestcasePresenter) Error(err error) api.CreateTestcaseResponseObject {
	badRequestErrors := []error{
		createtestcase.ErrorInvalidGradingItemID,
		createtestcase.ErrorEmptyName,
		createtestcase.ErrorEmptyCommand,
		createtestcase.ErrorInvalidScore,
		createtestcase.ErrorDependTestcaseNotFound,
	}

	for _, e := range badRequestErrors {
		if errors.Is(err, e) {
			return api.CreateTestcase400JSONResponse{
				Success: false,
				Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
			}
		}
	}

	switch {
	case errors.Is(err, createtestcase.ErrorGradingItemNotFound):
		return api.CreateTestcase404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	case errors.Is(err, createtestcase.ErrorInternalServer):
		fallthrough
	default:
		return api.CreateTestcase500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
