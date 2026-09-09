package controller

import (
	"judge-backend/api"
	jsonpresenter "judge-backend/internal/app/identity_domain/user/adapter/presenter/json"
	"judge-backend/internal/app/identity_domain/user/application"
	adduserstogroup "judge-backend/internal/app/identity_domain/user/application/usecase/group/add_users_to_group"
	creategroup "judge-backend/internal/app/identity_domain/user/application/usecase/group/create_group"
	deletegroup "judge-backend/internal/app/identity_domain/user/application/usecase/group/delete_group"
	getgroup "judge-backend/internal/app/identity_domain/user/application/usecase/group/get_group"
	listgroups "judge-backend/internal/app/identity_domain/user/application/usecase/group/list_groups"
	removeusersfromgroup "judge-backend/internal/app/identity_domain/user/application/usecase/group/remove_users_from_group"
	updategroup "judge-backend/internal/app/identity_domain/user/application/usecase/group/update_group"
	"judge-backend/internal/app/shared/types/httperror"

	"github.com/gin-gonic/gin"
)

type GroupController struct {
	app *application.GroupApplication
}

func NewGroupController(app *application.GroupApplication) *GroupController {
	return &GroupController{
		app: app,
	}
}

func (c *GroupController) CreateGroup(ctx *gin.Context, request api.CreateGroupRequestObject) (api.CreateGroupResponseObject, error) {
	description := ""
	if request.Body.Description != nil {
		description = *request.Body.Description
	}
	out, err := c.app.CreateGroup(&creategroup.Input{
		Name:        request.Body.Name,
		Description: description,
	})
	if err != nil {
		return httperror.From(err).Response(), nil
	}
	return api.CreateGroup200JSONResponse{
		Data:    int32(out.Group.GetID()),
		Success: true,
	}, nil
}

func (c *GroupController) GetGroup(ctx *gin.Context, request api.GetGroupRequestObject) (api.GetGroupResponseObject, error) {
	out, err := c.app.GetGroup(&getgroup.Input{
		ID: int(request.Id),
	})
	if err != nil {
		return httperror.From(err).Response(), nil
	}
	return jsonpresenter.GetGroup(out), nil
}

func (c *GroupController) ListGroup(ctx *gin.Context, request api.ListGroupRequestObject) (api.ListGroupResponseObject, error) {
	out, err := c.app.ListGroups(&listgroups.Input{})
	if err != nil {
		return httperror.From(err).Response(), nil
	}
	return jsonpresenter.ListGroups(out), nil
}

func (c *GroupController) UpdateGroup(ctx *gin.Context, request api.UpdateGroupRequestObject) (api.UpdateGroupResponseObject, error) {
	_, err := c.app.UpdateGroup(&updategroup.Input{
		ID:          int(request.Id),
		Name:        request.Body.Name,
		Description: request.Body.Description,
	})
	if err != nil {
		return httperror.From(err).Response(), nil
	}
	return api.UpdateGroup200JSONResponse{Success: true}, nil
}

func (c *GroupController) DeleteGroup(ctx *gin.Context, request api.DeleteGroupRequestObject) (api.DeleteGroupResponseObject, error) {
	_, err := c.app.DeleteGroup(&deletegroup.Input{
		ID: int(request.Id),
	})
	if err != nil {
		return httperror.From(err).Response(), nil
	}
	return api.DeleteGroup200JSONResponse{Success: true}, nil
}

func (c *GroupController) AddUsersToGroup(ctx *gin.Context, request api.AddUsersToGroupRequestObject) (api.AddUsersToGroupResponseObject, error) {
	userIDs := make([]int, 0, len(request.Body.UserIds))
	for _, id := range request.Body.UserIds {
		userIDs = append(userIDs, int(id))
	}
	_, err := c.app.AddUsersToGroup(&adduserstogroup.Input{
		GroupID: int(request.Id),
		UserIDs: userIDs,
	})
	if err != nil {
		return httperror.From(err).Response(), nil
	}
	return api.AddUsersToGroup200JSONResponse{Success: true}, nil
}

func (c *GroupController) RemoveUsersFromGroup(ctx *gin.Context, request api.RemoveUsersFromGroupRequestObject) (api.RemoveUsersFromGroupResponseObject, error) {
	userIDs := make([]int, 0, len(request.Body.UserIds))
	for _, id := range request.Body.UserIds {
		userIDs = append(userIDs, int(id))
	}
	_, err := c.app.RemoveUsersFromGroup(&removeusersfromgroup.Input{
		GroupID: int(request.Id),
		UserIDs: userIDs,
	})
	if err != nil {
		return httperror.From(err).Response(), nil
	}
	return api.RemoveUsersFromGroup200JSONResponse{Success: true}, nil
}
