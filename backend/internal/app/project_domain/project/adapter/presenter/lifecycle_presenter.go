package presenter

import (
	jsonpresenter "judge-backend/internal/app/project_domain/project/adapter/presenter/json/lifecycle"
)

type LifecyclePresenter struct {
	GetLifecycle    jsonpresenter.GetLifecyclePresenter
	CreateLifecycle jsonpresenter.CreateLifecyclePresenter
	UpdateLifecycle jsonpresenter.UpdateLifecyclePresenter
	DeleteLifecycle jsonpresenter.DeleteLifecyclePresenter
}

func NewLifecyclePresenter() *LifecyclePresenter {
	return &LifecyclePresenter{
		GetLifecycle:    jsonpresenter.NewGetLifecyclePresenter(),
		CreateLifecycle: jsonpresenter.NewCreateLifecyclePresenter(),
		UpdateLifecycle: jsonpresenter.NewUpdateLifecyclePresenter(),
		DeleteLifecycle: jsonpresenter.NewDeleteLifecyclePresenter(),
	}
}
