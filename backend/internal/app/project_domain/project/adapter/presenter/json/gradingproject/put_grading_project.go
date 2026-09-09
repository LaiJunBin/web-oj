package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/project_domain/project/adapter/presenter/types"
	gradingproject "judge-backend/internal/app/project_domain/project/application/service/grading_project"
	putgradingproject "judge-backend/internal/app/project_domain/project/application/usecase/grading_project/put_grading_project"
)

type PutGradingProjectPresenter types.Presenter[putgradingproject.Output, api.PutGradingProjectResponseObject]

type putGradingProjectPresenter struct{}

func NewPutGradingProjectPresenter() PutGradingProjectPresenter {
	return &putGradingProjectPresenter{}
}

func (p *putGradingProjectPresenter) Output(output *putgradingproject.Output) api.PutGradingProjectResponseObject {
	return api.PutGradingProject200JSONResponse{
		Success: true,
	}
}

func (p *putGradingProjectPresenter) Error(err error) api.PutGradingProjectResponseObject {
	switch {
	case errors.Is(err, putgradingproject.ErrorInvalidProjectID):
		fallthrough
	case errors.Is(err, putgradingproject.ErrorZipProjectRequired):
		fallthrough
	case errors.Is(err, gradingproject.ErrorInvalidZipProject):
		return api.PutGradingProject400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
		}
	case errors.Is(err, putgradingproject.ErrorProjectNotFound):
		return api.PutGradingProject404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	case errors.Is(err, putgradingproject.ErrorInternalServer):
		fallthrough
	default:
		return api.PutGradingProject500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
