package activity

import (
	contestprojectcontroller "judge-backend/internal/app/contest_domain/activity/adapter/controller"
	"judge-backend/internal/app/contest_domain/activity/adapter/presenter"
	systemstatusespresenter "judge-backend/internal/app/contest_domain/activity/adapter/presenter/event_stream/system_status"
	"judge-backend/internal/app/contest_domain/activity/application"
	getclarification "judge-backend/internal/app/contest_domain/activity/application/usecase/clarification/get_clarification"
	listclarifications "judge-backend/internal/app/contest_domain/activity/application/usecase/clarification/list_clarifications"
	replyclarification "judge-backend/internal/app/contest_domain/activity/application/usecase/clarification/reply_clarification"
	sendclarification "judge-backend/internal/app/contest_domain/activity/application/usecase/clarification/send_clarification"
	updateclarification "judge-backend/internal/app/contest_domain/activity/application/usecase/clarification/update_clarification"
	exportscoredetail "judge-backend/internal/app/contest_domain/activity/application/usecase/contest_project/export_score_detail"
	getcontestproject "judge-backend/internal/app/contest_domain/activity/application/usecase/contest_project/get_contest_project"
	getjudgmentanalytics "judge-backend/internal/app/contest_domain/activity/application/usecase/contest_project/get_judgment_analytics"
	getuserscores "judge-backend/internal/app/contest_domain/activity/application/usecase/contest_project/get_user_scores"
	rejudgecontestproject "judge-backend/internal/app/contest_domain/activity/application/usecase/contest_project/rejudge_contest_project"
	canceljudge "judge-backend/internal/app/contest_domain/activity/application/usecase/judgment/cancel_judge"
	setstaffnote "judge-backend/internal/app/contest_domain/activity/application/usecase/judgment/set_staff_note"
	approverejudge "judge-backend/internal/app/contest_domain/activity/application/usecase/rejudge/approve_rejudge"
	cancelrejudge "judge-backend/internal/app/contest_domain/activity/application/usecase/rejudge/cancel_rejudge"
	getrejudge "judge-backend/internal/app/contest_domain/activity/application/usecase/rejudge/get_rejudge"
	getrejudgesubmissioncomparison "judge-backend/internal/app/contest_domain/activity/application/usecase/rejudge/get_rejudge_submission_comparison"
	listrejudges "judge-backend/internal/app/contest_domain/activity/application/usecase/rejudge/list_rejudges"
	listrejudgesbycontestproject "judge-backend/internal/app/contest_domain/activity/application/usecase/rejudge/list_rejudges_by_contest_project"
	rejectrejudge "judge-backend/internal/app/contest_domain/activity/application/usecase/rejudge/reject_rejudge"
	getsubmission "judge-backend/internal/app/contest_domain/activity/application/usecase/submission/get_submission"
	getsubmissionsbycontestproject "judge-backend/internal/app/contest_domain/activity/application/usecase/submission/get_submissions_by_contest_project"
	getsubmissionsbyuserid "judge-backend/internal/app/contest_domain/activity/application/usecase/submission/get_submissions_by_user_id"
	ignoresubmission "judge-backend/internal/app/contest_domain/activity/application/usecase/submission/ignore_submission"
	rejudgesubmission "judge-backend/internal/app/contest_domain/activity/application/usecase/submission/rejudge_submission"
	streamsubmissionsbycontestproject "judge-backend/internal/app/contest_domain/activity/application/usecase/submission/stream_submissions_by_contest_project"
	unignoresubmission "judge-backend/internal/app/contest_domain/activity/application/usecase/submission/unignore_submission"
	getsubmissionrate "judge-backend/internal/app/contest_domain/activity/application/usecase/system_status/get_submission_rate"
	listjudgmenterrors "judge-backend/internal/app/contest_domain/activity/application/usecase/system_status/list_judgment_errors"
	"judge-backend/internal/app/contest_domain/activity/infrastructure/query"
	"judge-backend/internal/app/contest_domain/activity/infrastructure/repository"
	"judge-backend/internal/app/contest_domain/activity/infrastructure/service"
	"judge-backend/internal/app/shared/types/provider"
)

func NewProvider() *provider.Provider {
	return provider.New(provider.Config{
		Infrastructures: []interface{}{
			service.NewScoreDetailReportGenerator,
			query.NewContestProjectQuery,
			query.NewClarificationQuery,
			query.NewSubmissionQuery,
			query.NewRejudgeQuery,
			query.NewSystemStatusQuery,

			repository.NewClarificationRepository,
			repository.NewSubmissionRepository,
			repository.NewJudgmentRepository,
			repository.NewRejudgeRepository,
		},
		Controllers: []interface{}{
			contestprojectcontroller.NewContestProjectController,
			contestprojectcontroller.NewClarificationController,
			contestprojectcontroller.NewSubmissionController,
			contestprojectcontroller.NewJudgmentController,
			contestprojectcontroller.NewRejudgeController,
			contestprojectcontroller.NewSystemStatusController,
		},
		Presenters: []interface{}{
			presenter.NewContestProjectPresenter,
			presenter.NewClarificationPresenter,
			presenter.NewSubmissionPresenter,
			presenter.NewJudgmentPresenter,
			presenter.NewRejudgePresenter,
			systemstatusespresenter.NewStreamSubmissionRatePresenter,
			systemstatusespresenter.NewStreamJudgmentErrorsPresenter,
			presenter.NewSystemStatusPresenter,
		},
		Usecases: []interface{}{
			getcontestproject.NewUsecase,
			getuserscores.NewUsecase,
			rejudgecontestproject.NewUsecase,
			exportscoredetail.NewUsecase,
			getjudgmentanalytics.NewUsecase,

			listclarifications.NewUsecase,
			getclarification.NewUsecase,
			sendclarification.NewUsecase,
			replyclarification.NewUsecase,
			updateclarification.NewUsecase,

			getsubmissionsbycontestproject.NewUsecase,
			streamsubmissionsbycontestproject.NewUsecase,
			getsubmission.NewUsecase,
			getsubmissionsbyuserid.NewUsecase,
			rejudgesubmission.NewUsecase,
			ignoresubmission.NewUsecase,
			unignoresubmission.NewUsecase,

			canceljudge.NewUsecase,
			setstaffnote.NewUsecase,

			listrejudges.NewUsecase,
			listrejudgesbycontestproject.NewUsecase,
			getrejudge.NewUsecase,
			getrejudgesubmissioncomparison.NewUsecase,
			approverejudge.NewUsecase,
			rejectrejudge.NewUsecase,
			cancelrejudge.NewUsecase,

			getsubmissionrate.NewUsecase,
			listjudgmenterrors.NewUsecase,
		},
		Applications: []interface{}{
			application.NewContestProjectApplication,
			application.NewClarificationApplication,
			application.NewSubmissionApplication,
			application.NewJudgmentApplication,
			application.NewRejudgeApplication,
			application.NewSystemStatusApplication,
		},
		Services: []interface{}{
			service.NewSubmissionService,
			service.NewJudgmentService,
			service.NewClarificationService,
			service.NewRejudgeService,
		},
	})
}
