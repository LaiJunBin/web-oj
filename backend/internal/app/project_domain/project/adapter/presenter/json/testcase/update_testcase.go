package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/project_domain/project/adapter/presenter/types"
	updatetestcase "judge-backend/internal/app/project_domain/project/application/usecase/testcase/update_testcase"
	"judge-backend/internal/app/project_domain/project/domain/entity/testcase"
)

type UpdateTestcasePresenter types.Presenter[updatetestcase.Output, api.UpdateTestcaseResponseObject]

type updateTestcasePresenter struct{}

func NewUpdateTestcasePresenter() UpdateTestcasePresenter {
	return &updateTestcasePresenter{}
}

func (p *updateTestcasePresenter) Output(output *updatetestcase.Output) api.UpdateTestcaseResponseObject {
	return api.UpdateTestcase200JSONResponse{
		Success: true,
	}
}

func (p *updateTestcasePresenter) Error(err error) api.UpdateTestcaseResponseObject {
	badRequestErrors := []error{
		updatetestcase.ErrorInvalidID,
		updatetestcase.ErrorEmptyInput,
		updatetestcase.ErrorInvalidName,
		updatetestcase.ErrorInvalidScore,
		updatetestcase.ErrorDependTestcaseNotFound,
		testcase.ErrorDependSelf,
	}

	for _, e := range badRequestErrors {
		if errors.Is(err, e) {
			return api.UpdateTestcase400JSONResponse{
				Success: false,
				Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
			}
		}
	}

	switch {
	case errors.Is(err, updatetestcase.ErrorTestcaseNotFound):
		return api.UpdateTestcase404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	case errors.Is(err, updatetestcase.ErrorInternalServer):
		fallthrough
	default:
		return api.UpdateTestcase500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
