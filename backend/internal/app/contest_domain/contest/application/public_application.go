package application

import (
	"judge-backend/internal/app/contest_domain/contest/application/usecase/public/common/getsessionttl"
	getpubliccontest "judge-backend/internal/app/contest_domain/contest/application/usecase/public/contest/get_public_contest"
	getpubliccontestschedule "judge-backend/internal/app/contest_domain/contest/application/usecase/public/contest/get_public_contest_schedule"
	listpubliccontests "judge-backend/internal/app/contest_domain/contest/application/usecase/public/contest/list_public_contests"
	getpublicproject "judge-backend/internal/app/contest_domain/contest/application/usecase/public/project/get_public_project"
	getpublicprojectscoreboard "judge-backend/internal/app/contest_domain/contest/application/usecase/public/project/get_public_project_scoreboard"
	streampublicprojectscoreboard "judge-backend/internal/app/contest_domain/contest/application/usecase/public/project/stream_public_project_scoreboard"
	streampublicprojecttime "judge-backend/internal/app/contest_domain/contest/application/usecase/public/project/stream_public_project_time"
)

type PublicApplication struct {
	getSessionTtl                 getsessionttl.Usecase
	getPublicContestSchedule      getpubliccontestschedule.Usecase
	listPublicContests            listpubliccontests.Usecase
	getPublicContest              getpubliccontest.Usecase
	getPublicProject              getpublicproject.Usecase
	getPublicProjectScoreboard    getpublicprojectscoreboard.Usecase
	streamPublicProjectTime       streampublicprojecttime.Usecase
	streamPublicProjectScoreboard streampublicprojectscoreboard.Usecase
}

func NewPublicApplication(
	getSessionTtl getsessionttl.Usecase,
	getPublicContestSchedule getpubliccontestschedule.Usecase,
	listPublicContests listpubliccontests.Usecase,
	getPublicContest getpubliccontest.Usecase,
	getPublicProject getpublicproject.Usecase,
	getPublicProjectScoreboard getpublicprojectscoreboard.Usecase,
	streamPublicProjectTime streampublicprojecttime.Usecase,
	streamPublicProjectScoreboard streampublicprojectscoreboard.Usecase,
) *PublicApplication {
	return &PublicApplication{
		getSessionTtl:                 getSessionTtl,
		getPublicContestSchedule:      getPublicContestSchedule,
		listPublicContests:            listPublicContests,
		getPublicContest:              getPublicContest,
		getPublicProject:              getPublicProject,
		getPublicProjectScoreboard:    getPublicProjectScoreboard,
		streamPublicProjectTime:       streamPublicProjectTime,
		streamPublicProjectScoreboard: streamPublicProjectScoreboard,
	}
}

func (a *PublicApplication) GetSessionTtl(input *getsessionttl.Input) (*getsessionttl.Output, error) {
	return a.getSessionTtl.Execute(input)
}

func (a *PublicApplication) GetPublicContestSchedule(input *getpubliccontestschedule.Input) (*getpubliccontestschedule.Output, error) {
	return a.getPublicContestSchedule.Execute(input)
}

func (a *PublicApplication) ListPublicContests(input *listpubliccontests.Input) (*listpubliccontests.Output, error) {
	return a.listPublicContests.Execute(input)
}

func (a *PublicApplication) GetPublicContest(input *getpubliccontest.Input) (*getpubliccontest.Output, error) {
	return a.getPublicContest.Execute(input)
}

func (a *PublicApplication) GetPublicProject(input *getpublicproject.Input) (*getpublicproject.Output, error) {
	return a.getPublicProject.Execute(input)
}

func (a *PublicApplication) GetPublicProjectScoreboard(input *getpublicprojectscoreboard.Input) (*getpublicprojectscoreboard.Output, error) {
	return a.getPublicProjectScoreboard.Execute(input)
}

func (a *PublicApplication) StreamPublicProjectTime(input *streampublicprojecttime.Input) (*streampublicprojecttime.Output, error) {
	return a.streamPublicProjectTime.Execute(input)
}

func (a *PublicApplication) StreamPublicProjectScoreboard(input *streampublicprojectscoreboard.Input) (*streampublicprojectscoreboard.Output, error) {
	return a.streamPublicProjectScoreboard.Execute(input)
}
