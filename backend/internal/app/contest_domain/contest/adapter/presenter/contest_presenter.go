package presenter

import (
	filepresenter "judge-backend/internal/app/contest_domain/contest/adapter/presenter/file/contest"
	jsonpresenter "judge-backend/internal/app/contest_domain/contest/adapter/presenter/json"
)

type ContestPresenter struct {
	ListContest   jsonpresenter.ListContestPresenter
	GetContest    jsonpresenter.GetContestPresenter
	CreateContest jsonpresenter.CreateContestPresenter
	UpdateContest jsonpresenter.JSONPresenter
	DeleteContest jsonpresenter.JSONPresenter
	ExportContest filepresenter.ExportContestPresenter
	ImportContest jsonpresenter.ImportContestPresenter
}

func NewContestPresenter() *ContestPresenter {
	return &ContestPresenter{
		ListContest:   jsonpresenter.NewListContestPresenter(),
		GetContest:    jsonpresenter.NewGetContestPresenter(),
		CreateContest: jsonpresenter.NewCreateContestPresenter(),
		UpdateContest: jsonpresenter.New(),
		DeleteContest: jsonpresenter.New(),
		ExportContest: filepresenter.NewExportContestPresenter(),
		ImportContest: jsonpresenter.NewImportContestPresenter(),
	}
}
