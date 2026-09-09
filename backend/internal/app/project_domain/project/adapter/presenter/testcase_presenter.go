package presenter

import jsonpresenter "judge-backend/internal/app/project_domain/project/adapter/presenter/json/testcase"

type TestcasePresenter struct {
	ListProjectTestcases jsonpresenter.ListProjectTestcasesPresenter
	GetTestcase          jsonpresenter.GetTestcasePresenter
	CreateTestcase       jsonpresenter.CreateTestcasePresenter
	UpdateTestcase       jsonpresenter.UpdateTestcasePresenter
	DeleteTestcase       jsonpresenter.DeleteTestcasePresenter
}

func NewTestcasePresenter() *TestcasePresenter {
	return &TestcasePresenter{
		ListProjectTestcases: jsonpresenter.NewListProjectTestcasesPresenter(),
		GetTestcase:          jsonpresenter.NewGetTestcasePresenter(),
		CreateTestcase:       jsonpresenter.NewCreateTestcasePresenter(),
		UpdateTestcase:       jsonpresenter.NewUpdateTestcasePresenter(),
		DeleteTestcase:       jsonpresenter.NewDeleteTestcasePresenter(),
	}
}
