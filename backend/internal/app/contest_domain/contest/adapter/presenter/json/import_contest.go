package jsonpresenter

import (
	"errors"

	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/contest/adapter/presenter/types"
	importcontest "judge-backend/internal/app/contest_domain/contest/application/usecase/import_contest"
)

type ImportContestPresenter types.Presenter[importcontest.Output, api.ImportContestResponseObject]

type importContestPresenter struct{}

func NewImportContestPresenter() ImportContestPresenter { return &importContestPresenter{} }

func (p *importContestPresenter) Output(output *importcontest.Output) api.ImportContestResponseObject {
	return api.ImportContest200JSONResponse{
		Success: true,
		Data: api.AdminContestSpecImportContestResult{
			ContestId:          output.ContestID,
			Name:               output.Name,
			CreatedProjects:    output.CreatedProjects,
			UnresolvedProjects: output.UnresolvedProjects,
		},
	}
}

func (p *importContestPresenter) Error(err error) api.ImportContestResponseObject {
	var invalid *importcontest.InvalidBundleError
	if errors.As(err, &invalid) {
		return api.ImportContest400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
			Details: invalid.Details,
		}
	}

	return api.ImportContest500JSONResponse{
		Success: false,
		Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
	}
}
