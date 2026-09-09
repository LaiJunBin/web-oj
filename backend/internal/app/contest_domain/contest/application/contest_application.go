package application

import (
	createcontest "judge-backend/internal/app/contest_domain/contest/application/usecase/create_contest"
	deletecontest "judge-backend/internal/app/contest_domain/contest/application/usecase/delete_contest"
	exportcontest "judge-backend/internal/app/contest_domain/contest/application/usecase/export_contest"
	getcontest "judge-backend/internal/app/contest_domain/contest/application/usecase/get_contest"
	importcontest "judge-backend/internal/app/contest_domain/contest/application/usecase/import_contest"
	listcontest "judge-backend/internal/app/contest_domain/contest/application/usecase/list_contest"
	updatecontest "judge-backend/internal/app/contest_domain/contest/application/usecase/update_contest"
)

type ContestApplication struct {
	listContest   listcontest.Usecase
	getContest    getcontest.Usecase
	createContest createcontest.Usecase
	updateContest updatecontest.Usecase
	deleteContest deletecontest.Usecase
	exportContest exportcontest.Usecase
	importContest importcontest.Usecase
}

func NewContestApplication(
	listContest listcontest.Usecase,
	getContest getcontest.Usecase,
	createContest createcontest.Usecase,
	updateContest updatecontest.Usecase,
	deleteContest deletecontest.Usecase,
	exportContest exportcontest.Usecase,
	importContest importcontest.Usecase,
) *ContestApplication {
	return &ContestApplication{
		listContest:   listContest,
		getContest:    getContest,
		createContest: createContest,
		updateContest: updateContest,
		deleteContest: deleteContest,
		exportContest: exportContest,
		importContest: importContest,
	}
}

func (a *ContestApplication) ListContest(input *listcontest.Input) (*listcontest.Output, error) {
	return a.listContest.Execute(input)
}

func (a *ContestApplication) GetContest(input *getcontest.Input) (*getcontest.Output, error) {
	return a.getContest.Execute(input)
}

func (a *ContestApplication) CreateContest(input *createcontest.Input) (*createcontest.Output, error) {
	return a.createContest.Execute(input)
}

func (a *ContestApplication) UpdateContest(input *updatecontest.Input) (*updatecontest.Output, error) {
	return a.updateContest.Execute(input)
}

func (a *ContestApplication) DeleteContest(input *deletecontest.Input) (*deletecontest.Output, error) {
	return a.deleteContest.Execute(input)
}

func (a *ContestApplication) ExportContest(input *exportcontest.Input) (*exportcontest.Output, error) {
	return a.exportContest.Execute(input)
}

func (a *ContestApplication) ImportContest(input *importcontest.Input) (*importcontest.Output, error) {
	return a.importContest.Execute(input)
}
