package jsonpresenter

import (
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/activity/adapter/presenter/types"
	"judge-backend/internal/app/contest_domain/activity/application/query/dto"
	listrejudges "judge-backend/internal/app/contest_domain/activity/application/usecase/rejudge/list_rejudges"
	"judge-backend/internal/app/shared/types/httperror"
)

// rejudgeList renders the listing shape shared by the whole-instance listing and
// the per-contest-project one. They differ only in which rejudges they are
// given, so keeping one renderer is what stops the two responses from drifting
// apart field by field.
func rejudgeList(infos []*dto.RejudgeInfo) []api.AdminJudgmentSpecRejudge {
	rejudges := []api.AdminJudgmentSpecRejudge{}
	for _, rejudge := range infos {
		rejudges = append(rejudges, api.AdminJudgmentSpecRejudge{
			Id:                 rejudge.ID,
			ContestProjectId:   rejudge.ContestProjectID,
			Reason:             rejudge.Reason,
			EffectedJudgments:  rejudge.EffectedJudgments,
			CompletedJudgments: rejudge.CompletedJudgments,
			Status:             api.AdminJudgmentSpecRejudgeStatus(rejudge.Status),
			CreatedAt:          rejudge.CreatedAt,
		})
	}
	return rejudges
}

type ListRejudgesPresenter types.Presenter[listrejudges.Output, api.ListRejudgesResponseObject]

type listRejudgesPresenter struct {
}

func NewListRejudgesPresenter() ListRejudgesPresenter {
	return &listRejudgesPresenter{}
}

func (p *listRejudgesPresenter) Output(output *listrejudges.Output) api.ListRejudgesResponseObject {
	return api.ListRejudges200JSONResponse{
		Success: true,
		Data:    rejudgeList(output.Rejudges),
	}
}

func (p *listRejudgesPresenter) Error(err error) api.ListRejudgesResponseObject {
	return httperror.From(err).Response()
}
