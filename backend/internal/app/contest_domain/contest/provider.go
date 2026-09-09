package contest

import (
	contestcontroller "judge-backend/internal/app/contest_domain/contest/adapter/controller"
	adapterhandler "judge-backend/internal/app/contest_domain/contest/adapter/handler"
	"judge-backend/internal/app/contest_domain/contest/adapter/presenter"
	"judge-backend/internal/app/contest_domain/contest/application"
	cancelusersubmission "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/cancel_user_submission"
	downloaduserprojectasset "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/download_user_project_asset"
	getuserclarification "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/get_user_clarification"
	getusercontest "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/get_user_contest"
	getusercontestschedule "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/get_user_contest_schedule"
	getuserproject "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/get_user_project"
	getuserprojectscoreboard "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/get_user_project_scoreboard"
	getusersubmission "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/get_user_submission"
	listusercontests "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/list_user_contests"
	listuserprojectclarifications "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/list_user_project_clarifications"
	listuserprojectsubmissions "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/list_user_project_submissions"
	replyclarificationtoadmin "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/reply_clarification_to_admin"
	sendclarificationtoadmin "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/send_clarification_to_admin"
	streamuserprojectclarification "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/stream_user_project_clarification"
	streamuserprojectscoreboard "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/stream_user_project_scoreboard"
	streamuserprojecttime "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/stream_user_project_time"
	streamusersubmission "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/stream_user_submission"
	submituserproject "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/submit_user_project"
	createcontest "judge-backend/internal/app/contest_domain/contest/application/usecase/create_contest"
	deletecontest "judge-backend/internal/app/contest_domain/contest/application/usecase/delete_contest"
	exportcontest "judge-backend/internal/app/contest_domain/contest/application/usecase/export_contest"
	getcontest "judge-backend/internal/app/contest_domain/contest/application/usecase/get_contest"
	importcontest "judge-backend/internal/app/contest_domain/contest/application/usecase/import_contest"
	listcontest "judge-backend/internal/app/contest_domain/contest/application/usecase/list_contest"
	addgroups "judge-backend/internal/app/contest_domain/contest/application/usecase/participant/add_groups"
	addusers "judge-backend/internal/app/contest_domain/contest/application/usecase/participant/add_users"
	listparticipants "judge-backend/internal/app/contest_domain/contest/application/usecase/participant/list_participants"
	removegroups "judge-backend/internal/app/contest_domain/contest/application/usecase/participant/remove_groups"
	removeusers "judge-backend/internal/app/contest_domain/contest/application/usecase/participant/remove_users"
	"judge-backend/internal/app/contest_domain/contest/application/usecase/public/common/getsessionttl"
	getpubliccontest "judge-backend/internal/app/contest_domain/contest/application/usecase/public/contest/get_public_contest"
	getpubliccontestschedule "judge-backend/internal/app/contest_domain/contest/application/usecase/public/contest/get_public_contest_schedule"
	listpubliccontests "judge-backend/internal/app/contest_domain/contest/application/usecase/public/contest/list_public_contests"
	getpublicproject "judge-backend/internal/app/contest_domain/contest/application/usecase/public/project/get_public_project"
	getpublicprojectscoreboard "judge-backend/internal/app/contest_domain/contest/application/usecase/public/project/get_public_project_scoreboard"
	streampublicprojectscoreboard "judge-backend/internal/app/contest_domain/contest/application/usecase/public/project/stream_public_project_scoreboard"
	streampublicprojecttime "judge-backend/internal/app/contest_domain/contest/application/usecase/public/project/stream_public_project_time"
	updatecontest "judge-backend/internal/app/contest_domain/contest/application/usecase/update_contest"
	infrahandler "judge-backend/internal/app/contest_domain/contest/infrastructure/handler"
	"judge-backend/internal/app/contest_domain/contest/infrastructure/parser"
	"judge-backend/internal/app/contest_domain/contest/infrastructure/provider"
	"judge-backend/internal/app/contest_domain/contest/infrastructure/query"
	"judge-backend/internal/app/contest_domain/contest/infrastructure/repository"
	"judge-backend/internal/app/contest_domain/contest/infrastructure/service"
	coreprovider "judge-backend/internal/app/shared/types/provider"
)

func NewProvider() *coreprovider.Provider {
	providers := coreprovider.New(coreprovider.Config{
		Infrastructures: []interface{}{
			query.NewContestQuery,
			query.NewProjectQuery,
			query.NewClarificationQuery,
			query.NewSubmissionQuery,
			query.NewContestProjectQuery,
			query.NewSessionQuery,
			query.NewContestExportQuery,
			query.NewAdminContestQuery,

			repository.NewClarificationRepository,
			repository.NewProjectRepository,
			repository.NewSubmissionRepository,
			repository.NewContestRepository,
			repository.NewParticipantRepository,

			parser.NewDynamicFieldParser,
		},
		Controllers: []interface{}{
			contestcontroller.NewContestantController,
			contestcontroller.NewPublicController,
			contestcontroller.NewContestController,
			contestcontroller.NewParticipantController,
		},
		Presenters: []interface{}{
			presenter.NewPublicPresenter,
			presenter.NewContestantPresenter,
			presenter.NewContestPresenter,
		},
		Usecases: []interface{}{
			getpubliccontestschedule.NewUsecase,
			getsessionttl.NewUsecase,
			listpubliccontests.NewUsecase,
			getpubliccontest.NewUsecase,
			getpublicproject.NewUsecase,
			getpublicprojectscoreboard.NewUsecase,
			streampublicprojecttime.NewUsecase,
			streampublicprojectscoreboard.NewUsecase,

			getusercontestschedule.NewUsecase,
			listusercontests.NewUsecase,
			getusercontest.NewUsecase,
			getuserproject.NewUsecase,
			getuserprojectscoreboard.NewUsecase,
			listuserprojectclarifications.NewUsecase,
			sendclarificationtoadmin.NewUsecase,
			replyclarificationtoadmin.NewUsecase,
			getuserclarification.NewUsecase,
			listuserprojectsubmissions.NewUsecase,
			getusersubmission.NewUsecase,
			downloaduserprojectasset.NewUsecase,
			submituserproject.NewUsecase,
			cancelusersubmission.NewUsecase,
			streamuserprojecttime.NewUsecase,
			streamuserprojectscoreboard.NewUsecase,
			streamuserprojectclarification.NewUsecase,
			streamusersubmission.NewUsecase,

			listcontest.NewUsecase,
			getcontest.NewUsecase,
			createcontest.NewUsecase,
			updatecontest.NewUsecase,
			deletecontest.NewUsecase,
			exportcontest.NewUsecase,
			importcontest.NewUsecase,

			addusers.New,
			removeusers.New,
			addgroups.New,
			removegroups.New,
			listparticipants.New,
		},
		Applications: []interface{}{
			application.NewContestantApplication,
			application.NewPublicApplication,
			application.NewContestApplication,
			application.NewParticipantApplication,
		},
		Services: []interface{}{
			service.NewProjectSubmitPublisher,
			service.NewSubmissionCooldown,
			service.NewSubmissionService,

			// Bundle ports: name lookup for the slots a contest bundle names,
			// and the collision suffix for an imported contest.
			service.NewProjectLibrary,
			service.NewContestNamer,
		},
		Handlers: []interface{}{
			// Adapter handlers: shared events → contest-internal events
			adapterhandler.RegisterJudgedHandler,
			adapterhandler.RegisterJudgeUpdatedHandler,
			// Infra handlers: contest-internal events
			infrahandler.RegisterJudgedHandler,
			infrahandler.RegisterJudgeUpdatedHandler,
		},
		Providers: []interface{}{
			provider.NewStoragePathProvider,
			provider.NewJudgePriorityStrategyProvider,
		},
	})

	return providers
}
