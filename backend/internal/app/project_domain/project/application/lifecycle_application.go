package application

import (
	createlifecycle "judge-backend/internal/app/project_domain/project/application/usecase/lifecycle/create_lifecycle"
	deletelifecycle "judge-backend/internal/app/project_domain/project/application/usecase/lifecycle/delete_lifecycle"
	getlifecycle "judge-backend/internal/app/project_domain/project/application/usecase/lifecycle/get_lifecycle"
	updatelifecycle "judge-backend/internal/app/project_domain/project/application/usecase/lifecycle/update_lifecycle"
)

type LifecycleApplication struct {
	getLifecycle    getlifecycle.Usecase
	createLifecycle createlifecycle.Usecase
	updateLifecycle updatelifecycle.Usecase
	deleteLifecycle deletelifecycle.Usecase
}

func NewLifecycleApplication(
	getLifecycle getlifecycle.Usecase,
	createLifecycle createlifecycle.Usecase,
	updateLifecycle updatelifecycle.Usecase,
	deleteLifecycle deletelifecycle.Usecase,
) *LifecycleApplication {
	return &LifecycleApplication{
		getLifecycle:    getLifecycle,
		createLifecycle: createLifecycle,
		updateLifecycle: updateLifecycle,
		deleteLifecycle: deleteLifecycle,
	}
}

func (a LifecycleApplication) GetLifecycle(input *getlifecycle.Input) (*getlifecycle.Output, error) {
	return a.getLifecycle.Execute(input)
}

func (a LifecycleApplication) CreateLifecycle(input *createlifecycle.Input) (*createlifecycle.Output, error) {
	return a.createLifecycle.Execute(input)
}

func (a LifecycleApplication) UpdateLifecycle(input *updatelifecycle.Input) (*updatelifecycle.Output, error) {
	return a.updateLifecycle.Execute(input)
}

func (a LifecycleApplication) DeleteLifecycle(input *deletelifecycle.Input) (*deletelifecycle.Output, error) {
	return a.deleteLifecycle.Execute(input)
}
