package application

import (
	createtestcase "judge-backend/internal/app/project_domain/project/application/usecase/testcase/create_testcase"
	deletetestcase "judge-backend/internal/app/project_domain/project/application/usecase/testcase/delete_testcase"
	gettestcase "judge-backend/internal/app/project_domain/project/application/usecase/testcase/get_testcase"
	listprojecttestcases "judge-backend/internal/app/project_domain/project/application/usecase/testcase/list_project_testcases"
	updatetestcase "judge-backend/internal/app/project_domain/project/application/usecase/testcase/update_testcase"
)

type TestcaseApplication struct {
	listProjectTestcases listprojecttestcases.Usecase
	getTestcase          gettestcase.Usecase
	createTestcase       createtestcase.Usecase
	updateTestcase       updatetestcase.Usecase
	deleteTestcase       deletetestcase.Usecase
}

func NewTestcaseApplication(
	listProjectTestcases listprojecttestcases.Usecase,
	getTestcase gettestcase.Usecase,
	createTestcase createtestcase.Usecase,
	updateTestcase updatetestcase.Usecase,
	deleteTestcase deletetestcase.Usecase,
) *TestcaseApplication {
	return &TestcaseApplication{
		listProjectTestcases: listProjectTestcases,
		getTestcase:          getTestcase,
		createTestcase:       createTestcase,
		updateTestcase:       updateTestcase,
		deleteTestcase:       deleteTestcase,
	}
}

func (a TestcaseApplication) ListProjectTestcases(input *listprojecttestcases.Input) (*listprojecttestcases.Output, error) {
	return a.listProjectTestcases.Execute(input)
}

func (a TestcaseApplication) GetTestcase(input *gettestcase.Input) (*gettestcase.Output, error) {
	return a.getTestcase.Execute(input)
}

func (a TestcaseApplication) CreateTestcase(input *createtestcase.Input) (*createtestcase.Output, error) {
	return a.createTestcase.Execute(input)
}

func (a TestcaseApplication) UpdateTestcase(input *updatetestcase.Input) (*updatetestcase.Output, error) {
	return a.updateTestcase.Execute(input)
}

func (a TestcaseApplication) DeleteTestcase(input *deletetestcase.Input) (*deletetestcase.Output, error) {
	return a.deleteTestcase.Execute(input)
}
