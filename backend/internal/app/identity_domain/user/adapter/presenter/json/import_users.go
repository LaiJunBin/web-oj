package jsonpresenter

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/identity_domain/user/adapter/presenter/types"
	importusers "judge-backend/internal/app/identity_domain/user/application/usecase/import_users"
)

type ImportUsersPresenter struct{}

func NewImportUsersPresenter() types.ImportUsersPresenter {
	return &ImportUsersPresenter{}
}

func (p *ImportUsersPresenter) Output(output *importusers.Output) api.ImportUsersResponseObject {
	return &api.ImportUsers200JSONResponse{
		Success: true,
		Data:    int32(output.TotalImported),
	}
}

func (p *ImportUsersPresenter) Error(err error) api.ImportUsersResponseObject {
	var importUsersError *importusers.ImportUsersError
	switch {
	case errors.Is(err, importusers.ErrorInvalidJSON):
		return &api.ImportUsers400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
		}
	case errors.As(err, &importUsersError):
		return &api.ImportUsers400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
			Details: importUsersError.Details(),
		}
	case errors.Is(err, importusers.ErrorInternalServer):
		fallthrough
	default:
		return &api.ImportUsers500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
