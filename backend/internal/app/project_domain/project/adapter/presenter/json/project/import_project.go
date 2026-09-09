package jsonpresenter

import (
	"errors"

	"judge-backend/api"
	"judge-backend/internal/app/project_domain/project/adapter/presenter/types"
	importproject "judge-backend/internal/app/project_domain/project/application/usecase/project/import_project"
)

type ImportProjectPresenter types.Presenter[importproject.Output, api.ImportProjectResponseObject]

type importProjectPresenter struct{}

func NewImportProjectPresenter() ImportProjectPresenter {
	return &importProjectPresenter{}
}

func (p *importProjectPresenter) Output(output *importproject.Output) api.ImportProjectResponseObject {
	return api.ImportProject200JSONResponse{
		Success: true,
		Data: api.AdminProjectSpecImportProjectResult{
			ProjectId:        output.ProjectID,
			Name:             output.Name,
			ImageTag:         output.ImageTag,
			TestImageTag:     output.TestImageTag,
			MissingImageTags: output.MissingImageTags,
		},
	}
}

// Error returns every fault at once for a rejected bundle.
//
// The details are the whole point: a bundle can be hand-written, and an
// operator fixing one typo per upload would give up long before the file
// imported.
func (p *importProjectPresenter) Error(err error) api.ImportProjectResponseObject {
	var invalid *importproject.InvalidBundleError
	if errors.As(err, &invalid) {
		return api.ImportProject400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
			Details: invalid.Details,
		}
	}

	return api.ImportProject500JSONResponse{
		Success: false,
		Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
	}
}
