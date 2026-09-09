package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/project_domain/project/adapter/presenter/types"
	deletetestcase "judge-backend/internal/app/project_domain/project/application/usecase/testcase/delete_testcase"
)

type DeleteTestcasePresenter types.Presenter[deletetestcase.Output, api.DeleteTestcaseResponseObject]

type deleteTestcasePresenter struct{}

func NewDeleteTestcasePresenter() DeleteTestcasePresenter {
	return &deleteTestcasePresenter{}
}

func (p *deleteTestcasePresenter) Output(output *deletetestcase.Output) api.DeleteTestcaseResponseObject {
	return api.DeleteTestcase200JSONResponse{
		Success: true,
	}
}

func (p *deleteTestcasePresenter) Error(err error) api.DeleteTestcaseResponseObject {
	switch {
	case errors.Is(err, deletetestcase.ErrorInvalidID):
		return api.DeleteTestcase400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
		}
	case errors.Is(err, deletetestcase.ErrorTestcaseNotFound):
		return api.DeleteTestcase404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	case errors.Is(err, deletetestcase.ErrorInternalServer):
		fallthrough
	default:
		return api.DeleteTestcase500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
