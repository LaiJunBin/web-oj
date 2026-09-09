package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/judge_domain/worker/adapter/presenter/types"
	getsystemdefaults "judge-backend/internal/app/judge_domain/worker/application/usecase/get_system_defaults"
	"judge-backend/internal/app/shared/types/httperror"
)

type GetSystemDefaultsPresenter types.Presenter[getsystemdefaults.Output, api.GetSystemDefaultsResponseObject]

type getSystemDefaultsPresenter struct {
}

func NewGetSystemDefaultsPresenter() GetSystemDefaultsPresenter {
	return &getSystemDefaultsPresenter{}
}

// Output maps the allowlist field by field. Written out rather than copied
// wholesale on purpose: the compiler then flags a new field here, which is
// the second place (after the usecase's Output) that a careless addition has
// to get past before it reaches a response.
func (p *getSystemDefaultsPresenter) Output(output *getsystemdefaults.Output) api.GetSystemDefaultsResponseObject {
	return api.GetSystemDefaults200JSONResponse{
		Success: true,
		Data: api.AdminJudgeWorkerSpecSystemDefaults{
			MaxJudgeDurationMinutes:   output.MaxJudgeDurationMinutes,
			SubmissionCooldownSeconds: output.SubmissionCooldownSeconds,
			PidsLimit:                 output.PidsLimit,
			MemoryLimitMB:             output.MemoryLimitMB,
		},
	}
}

func (p *getSystemDefaultsPresenter) Error(err error) api.GetSystemDefaultsResponseObject {
	return httperror.From(err).Response()
}
