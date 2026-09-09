package presenter

import jsonpresenter "judge-backend/internal/app/project_domain/project/adapter/presenter/json/gradingitem"

type GradingItemPresenter struct {
	ListGradingItem      jsonpresenter.ListGradingItemPresenter
	GetGradingItem       jsonpresenter.GetGradingItemPresenter
	CreateGradingItem    jsonpresenter.CreateGradingItemPresenter
	UpdateGradingItem    jsonpresenter.UpdateGradingItemPresenter
	DeleteGradingItem    jsonpresenter.DeleteGradingItemPresenter
	MoveGradingItemOrder jsonpresenter.MoveGradingItemOrderPresenter
}

func NewGradingItemPresenter() *GradingItemPresenter {
	return &GradingItemPresenter{
		ListGradingItem:      jsonpresenter.NewListGradingItemPresenter(),
		GetGradingItem:       jsonpresenter.NewGetGradingItemPresenter(),
		CreateGradingItem:    jsonpresenter.NewCreateGradingItemPresenter(),
		UpdateGradingItem:    jsonpresenter.NewUpdateGradingItemPresenter(),
		DeleteGradingItem:    jsonpresenter.NewDeleteGradingItemPresenter(),
		MoveGradingItemOrder: jsonpresenter.NewMoveGradingItemOrderPresenter(),
	}
}
