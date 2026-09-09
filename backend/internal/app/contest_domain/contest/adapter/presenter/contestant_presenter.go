package presenter

import (
	espresenter "judge-backend/internal/app/contest_domain/contest/adapter/presenter/event_stream/contestant"
	filepresenter "judge-backend/internal/app/contest_domain/contest/adapter/presenter/file/contestant"
	jsonpresenter "judge-backend/internal/app/contest_domain/contest/adapter/presenter/json/contestant"
	commonpresenter "judge-backend/internal/app/shared/adapter/presenter/common"
	sharedtypes "judge-backend/internal/app/shared/types"
)

type ContestantPresenter struct {
	GetUserContestSchedule              jsonpresenter.GetUserContestSchedulePresenter
	ListUserContests                    jsonpresenter.ListUserContestsPresenter
	GetUserContest                      jsonpresenter.GetUserContestPresenter
	GetUserProject                      jsonpresenter.GetUserProjectPresenter
	GetUserProjectScoreboard            jsonpresenter.GetUserProjectScoreboardPresenter
	ListUserProjectClarifications       jsonpresenter.ListUserProjectClarificationsPresenter
	GetUserClarification                jsonpresenter.GetUserClarificationPresenter
	SendUserProjectClarificationToAdmin *commonpresenter.JSONPresenter
	ReplyUserProjectClarificationToUser *commonpresenter.JSONPresenter
	ListUserProjectSubmissions          jsonpresenter.ListUserProjectSubmissionsPresenter
	GetUserSubmission                   jsonpresenter.GetUserSubmissionPresenter
	DownloadUserProjectAsset            filepresenter.DownloadUserProjectAssetPresenter
	SubmitUserProject                   jsonpresenter.SubmitUserProjectPresenter
	CancelUserSubmission                *commonpresenter.JSONPresenter
	StreamUserProjectTime               espresenter.StreamUserProjectTimePresenter
	StreamUserProjectScoreboard         espresenter.StreamUserProjectScoreboardPresenter
	StreamUserProjectClarification      espresenter.StreamUserProjectClarificationPresenter
	StreamUserSubmission                espresenter.StreamUserSubmissionPresenter
}

func NewContestantPresenter(
	eventBus sharedtypes.EventBus,
) *ContestantPresenter {
	return &ContestantPresenter{
		GetUserContestSchedule:              jsonpresenter.NewGetUserContestSchedulePresenter(),
		ListUserContests:                    jsonpresenter.NewListUserContestsPresenter(),
		GetUserContest:                      jsonpresenter.NewGetUserContestPresenter(),
		GetUserProject:                      jsonpresenter.NewGetUserProjectPresenter(),
		GetUserProjectScoreboard:            jsonpresenter.NewGetUserProjectScoreboardPresenter(),
		ListUserProjectClarifications:       jsonpresenter.NewListUserProjectClarificationsPresenter(),
		GetUserClarification:                jsonpresenter.NewGetUserClarificationPresenter(),
		SendUserProjectClarificationToAdmin: commonpresenter.NewJSONPresenter(),
		ReplyUserProjectClarificationToUser: commonpresenter.NewJSONPresenter(),
		ListUserProjectSubmissions:          jsonpresenter.NewListUserProjectSubmissionsPresenter(),
		GetUserSubmission:                   jsonpresenter.NewGetUserSubmissionPresenter(),
		DownloadUserProjectAsset:            filepresenter.NewDownloadUserProjectAssetPresenter(),
		SubmitUserProject:                   jsonpresenter.NewSubmitUserProjectPresenter(),
		CancelUserSubmission:                commonpresenter.NewJSONPresenter(),
		StreamUserProjectTime:               espresenter.NewStreamUserProjectTimePresenter(),
		StreamUserProjectScoreboard:         espresenter.NewStreamUserProjectScoreboardPresenter(eventBus),
		StreamUserProjectClarification:      espresenter.NewStreamUserProjectClarificationPresenter(eventBus),
		StreamUserSubmission:                espresenter.NewStreamUserSubmissionPresenter(eventBus),
	}
}
