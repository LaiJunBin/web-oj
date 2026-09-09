package repository

import "judge-backend/internal/app/project_domain/project/domain/entity/testcase"

type TestcaseRepository interface {
	Exists(id int) (bool, error)
	FindByID(id int) (*testcase.Testcase, error)
	Save(testcase *testcase.Testcase) error
	Delete(testcase *testcase.Testcase) error
}
