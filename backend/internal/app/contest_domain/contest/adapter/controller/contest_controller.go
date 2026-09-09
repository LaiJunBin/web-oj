package contestcontroller

import (
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/contest/adapter/presenter"
	"judge-backend/internal/app/contest_domain/contest/application"
	createcontest "judge-backend/internal/app/contest_domain/contest/application/usecase/create_contest"
	deletecontest "judge-backend/internal/app/contest_domain/contest/application/usecase/delete_contest"
	exportcontest "judge-backend/internal/app/contest_domain/contest/application/usecase/export_contest"
	getcontest "judge-backend/internal/app/contest_domain/contest/application/usecase/get_contest"
	importcontest "judge-backend/internal/app/contest_domain/contest/application/usecase/import_contest"
	listcontest "judge-backend/internal/app/contest_domain/contest/application/usecase/list_contest"
	updatecontest "judge-backend/internal/app/contest_domain/contest/application/usecase/update_contest"
	fileheader "judge-backend/internal/app/shared/types/file_header"
	formrequest "judge-backend/internal/app/shared/types/form_request"

	"github.com/gin-gonic/gin"
)

type ContestController struct {
	app       *application.ContestApplication
	presenter *presenter.ContestPresenter
}

func NewContestController(
	app *application.ContestApplication,
	presenter *presenter.ContestPresenter,
) *ContestController {
	return &ContestController{
		app:       app,
		presenter: presenter,
	}
}

func (ctl *ContestController) ListContest(ctx *gin.Context, request api.ListContestRequestObject) (api.ListContestResponseObject, error) {
	input := &listcontest.Input{}

	output, err := ctl.app.ListContest(input)
	if err != nil {
		return ctl.presenter.ListContest.Error(err), nil
	}

	return ctl.presenter.ListContest.Output(output), nil
}

func (ctl *ContestController) GetContest(ctx *gin.Context, request api.GetContestRequestObject) (api.GetContestResponseObject, error) {
	input := &getcontest.Input{
		ID: request.Id,
	}

	output, err := ctl.app.GetContest(input)
	if err != nil {
		return ctl.presenter.GetContest.Error(err), nil
	}

	return ctl.presenter.GetContest.Output(output), nil
}

func (ctl *ContestController) CreateContest(ctx *gin.Context, request api.CreateContestRequestObject) (api.CreateContestResponseObject, error) {
	projects := make([]createcontest.CreateProjectConfig, 0)
	for _, project := range request.Body.Projects {
		projects = append(projects, createcontest.CreateProjectConfig{
			ProjectID:                 project.ProjectId,
			Name:                      project.Name,
			ActivationTime:            project.ActivateTime,
			DeactivationTime:          project.DeactivateTime,
			StartTime:                 project.StartTime,
			EndTime:                   project.EndTime,
			IsPublic:                  project.IsPublic,
			JudgePriorityStrategy:     project.JudgePriorityStrategy,
			MaxJudgeDurationMinutes:   project.MaxJudgeDurationMinutes,
			SubmissionCooldownSeconds: project.SubmissionCooldownSeconds,
			PidsLimit:                 project.PidsLimit,
			MemoryLimitMB:             project.MemoryLimitMB,
			ScoreboardCompact:         project.ScoreboardCompact != nil && *project.ScoreboardCompact,
			ScoreboardMaskNames:       project.ScoreboardMaskNames != nil && *project.ScoreboardMaskNames,
		})
	}

	input := &createcontest.Input{
		Name:           request.Body.Name,
		Description:    request.Body.Description,
		ActivateTime:   request.Body.ActivateTime,
		DeactivateTime: request.Body.DeactivateTime,
		IsPublic:       request.Body.IsPublic,
		ProjectConfigs: projects,
	}

	output, err := ctl.app.CreateContest(input)
	if err != nil {
		return ctl.presenter.CreateContest.Error(err), nil
	}

	return ctl.presenter.CreateContest.Output(output), nil
}

func (ctl *ContestController) DeleteContest(ctx *gin.Context, request api.DeleteContestRequestObject) (api.DeleteContestResponseObject, error) {
	input := &deletecontest.Input{
		ID: request.Id,
	}

	output, err := ctl.app.DeleteContest(input)
	if err != nil {
		return ctl.presenter.DeleteContest.Error(err), nil
	}

	return ctl.presenter.DeleteContest.Output(output), nil
}

func (ctl *ContestController) UpdateContest(ctx *gin.Context, request api.UpdateContestRequestObject) (api.UpdateContestResponseObject, error) {
	input := &updatecontest.Input{
		ID:             request.Id,
		Name:           request.Body.Name,
		Description:    request.Body.Description,
		ActivateTime:   request.Body.ActivateTime,
		DeactivateTime: request.Body.DeactivateTime,
		IsPublic:       request.Body.IsPublic,
	}

	if request.Body.Projects != nil {
		projects := make([]updatecontest.UpdateProjectConfig, 0)
		for _, project := range *request.Body.Projects {
			projects = append(projects, updatecontest.UpdateProjectConfig{
				ID:                        project.Id,
				Name:                      project.Name,
				ProjectID:                 project.ProjectId,
				ActivationTime:            project.ActivateTime,
				DeactivationTime:          project.DeactivateTime,
				StartTime:                 project.StartTime,
				EndTime:                   project.EndTime,
				IsPublic:                  project.IsPublic,
				JudgePriorityStrategy:     project.JudgePriorityStrategy,
				MaxJudgeDurationMinutes:   project.MaxJudgeDurationMinutes,
				SubmissionCooldownSeconds: project.SubmissionCooldownSeconds,
				PidsLimit:                 project.PidsLimit,
				MemoryLimitMB:             project.MemoryLimitMB,
				ScoreboardCompact:         project.ScoreboardCompact,
				ScoreboardMaskNames:       project.ScoreboardMaskNames,
			})
		}
		input.ProjectConfigs = &projects
	}

	output, err := ctl.app.UpdateContest(input)
	if err != nil {
		return ctl.presenter.UpdateContest.Error(err), nil
	}

	return ctl.presenter.UpdateContest.Output(output), nil
}

type ImportContestPayload struct {
	Bundle *fileheader.FileHeader `form:"bundle" validate:"required"`
}

func (ctl *ContestController) ExportContest(ctx *gin.Context, request api.ExportContestRequestObject) (api.ExportContestResponseObject, error) {
	output, err := ctl.app.ExportContest(&exportcontest.Input{ContestID: request.Id})
	if err != nil {
		return ctl.presenter.ExportContest.Error(err), nil
	}

	return ctl.presenter.ExportContest.Output(output), nil
}

func (ctl *ContestController) ImportContest(ctx *gin.Context, request api.ImportContestRequestObject) (api.ImportContestResponseObject, error) {
	payload := &ImportContestPayload{}
	if err := formrequest.New(request.Body).Bind(payload); err != nil {
		return ctl.presenter.ImportContest.Error(&importcontest.InvalidBundleError{
			Details: []string{"沒有收到 bundle 檔案"},
		}), nil
	}

	output, err := ctl.app.ImportContest(&importcontest.Input{Bundle: payload.Bundle.Buffer()})
	if err != nil {
		return ctl.presenter.ImportContest.Error(err), nil
	}

	return ctl.presenter.ImportContest.Output(output), nil
}
