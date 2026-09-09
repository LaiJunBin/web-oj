package application

import (
	creategradingitem "judge-backend/internal/app/project_domain/project/application/usecase/grading_item/create_grading_item"
	deletegradingitem "judge-backend/internal/app/project_domain/project/application/usecase/grading_item/delete_grading_item"
	getgradingitem "judge-backend/internal/app/project_domain/project/application/usecase/grading_item/get_grading_item"
	listgradingitem "judge-backend/internal/app/project_domain/project/application/usecase/grading_item/list_grading_item"
	movegradingitemorder "judge-backend/internal/app/project_domain/project/application/usecase/grading_item/move_grading_item_order"
	updategradingitem "judge-backend/internal/app/project_domain/project/application/usecase/grading_item/update_grading_item"
)

type GradingItemApplication struct {
	getGradingItem       getgradingitem.Usecase
	listGradingItem      listgradingitem.Usecase
	createGradingItem    creategradingitem.Usecase
	updateGradingItem    updategradingitem.Usecase
	deleteGradingItem    deletegradingitem.Usecase
	moveGradingItemOrder movegradingitemorder.Usecase
}

func NewGradingItemApplication(
	getGradingItem getgradingitem.Usecase,
	listGradingItem listgradingitem.Usecase,
	createGradingItem creategradingitem.Usecase,
	updateGradingItem updategradingitem.Usecase,
	deleteGradingItem deletegradingitem.Usecase,
	moveGradingItemOrder movegradingitemorder.Usecase,
) *GradingItemApplication {
	return &GradingItemApplication{
		getGradingItem:       getGradingItem,
		listGradingItem:      listGradingItem,
		createGradingItem:    createGradingItem,
		updateGradingItem:    updateGradingItem,
		deleteGradingItem:    deleteGradingItem,
		moveGradingItemOrder: moveGradingItemOrder,
	}
}

func (a GradingItemApplication) ListGradingItem(input *listgradingitem.Input) (*listgradingitem.Output, error) {
	return a.listGradingItem.Execute(input)
}

func (a GradingItemApplication) GetGradingItem(input *getgradingitem.Input) (*getgradingitem.Output, error) {
	return a.getGradingItem.Execute(input)
}

func (a GradingItemApplication) CreateGradingItem(input *creategradingitem.Input) (*creategradingitem.Output, error) {
	return a.createGradingItem.Execute(input)
}

func (a GradingItemApplication) UpdateGradingItem(input *updategradingitem.Input) (*updategradingitem.Output, error) {
	return a.updateGradingItem.Execute(input)
}

func (a GradingItemApplication) DeleteGradingItem(input *deletegradingitem.Input) (*deletegradingitem.Output, error) {
	return a.deleteGradingItem.Execute(input)
}

func (a GradingItemApplication) MoveGradingItemOrder(input *movegradingitemorder.Input) (*movegradingitemorder.Output, error) {
	return a.moveGradingItemOrder.Execute(input)
}
