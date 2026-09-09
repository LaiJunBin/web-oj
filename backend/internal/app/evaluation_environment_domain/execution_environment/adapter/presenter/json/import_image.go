package jsonpresenter

import (
	"errors"

	"judge-backend/api"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/adapter/presenter/types"
	importimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/import_image"
)

type importImagePresenter struct{}

func NewImportImagePresenter() types.ImportImagePresenterType { return &importImagePresenter{} }

func (p *importImagePresenter) Output(output *importimage.Output) api.ImportImageResponseObject {
	return api.ImportImage200JSONResponse{
		Success: true,
		Data: api.AdminImageSpecImportImageResult{
			Tag:         output.Tag,
			Mode:        output.Mode,
			BuildQueued: output.BuildQueued,
		},
	}
}

// Error keeps a conflict apart from a malformed bundle.
//
// They need different things from the operator: a bad file has to be fixed,
// a conflict has to be decided. Collapsing both into 400 would leave the
// frontend guessing which one it is by reading message text.
func (p *importImagePresenter) Error(err error) api.ImportImageResponseObject {
	var conflict *importimage.ConflictError
	if errors.As(err, &conflict) {
		return api.ImportImage409JSONResponse{
			Success: false,
			Message: err.Error(),
			Details: conflict.Details,
		}
	}

	var invalid *importimage.InvalidBundleError
	if errors.As(err, &invalid) {
		return api.ImportImage400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
			Details: invalid.Details,
		}
	}

	return api.ImportImage500JSONResponse{
		Success: false,
		Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
	}
}
