package presenter

import (
	espresenter "judge-backend/internal/app/contest_domain/contest/adapter/presenter/event_stream/public"
	jsonpresenter "judge-backend/internal/app/contest_domain/contest/adapter/presenter/json/public"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type PublicPresenter struct {
	GetPublicContestSchedule      jsonpresenter.GetPublicContestSchedulePresenter
	ListPublicContests            jsonpresenter.ListPublicContestsPresenter
	GetPublicContest              jsonpresenter.GetPublicContestPresenter
	GetPublicProject              jsonpresenter.GetPublicProjectPresenter
	GetPublicProjectScoreboard    jsonpresenter.GetPublicProjectScoreboardPresenter
	StreamPublicProjectTime       espresenter.StreamPublicProjectTimePresenter
	StreamPublicProjectScoreboard espresenter.StreamPublicProjectScoreboardPresenter
}

func NewPublicPresenter(
	eventBus sharedtypes.EventBus,
) *PublicPresenter {
	return &PublicPresenter{
		GetPublicContestSchedule:      jsonpresenter.NewGetPublicContestSchedulePresenter(),
		ListPublicContests:            jsonpresenter.NewListPublicContestsPresenter(),
		GetPublicContest:              jsonpresenter.NewGetPublicContestPresenter(),
		GetPublicProject:              jsonpresenter.NewGetPublicProjectPresenter(),
		GetPublicProjectScoreboard:    jsonpresenter.NewGetPublicProjectScoreboardPresenter(),
		StreamPublicProjectTime:       espresenter.NewStreamPublicProjectTimePresenter(),
		StreamPublicProjectScoreboard: espresenter.NewStreamPublicProjectScoreboardPresenter(eventBus),
	}
}
