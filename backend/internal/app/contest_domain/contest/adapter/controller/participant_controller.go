package contestcontroller

import (
	"errors"
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/contest/application"
	addgroups "judge-backend/internal/app/contest_domain/contest/application/usecase/participant/add_groups"
	addusers "judge-backend/internal/app/contest_domain/contest/application/usecase/participant/add_users"
	listparticipants "judge-backend/internal/app/contest_domain/contest/application/usecase/participant/list_participants"
	removegroups "judge-backend/internal/app/contest_domain/contest/application/usecase/participant/remove_groups"
	removeusers "judge-backend/internal/app/contest_domain/contest/application/usecase/participant/remove_users"

	"github.com/gin-gonic/gin"
)

type ParticipantController struct {
	app *application.ParticipantApplication
}

func NewParticipantController(app *application.ParticipantApplication) *ParticipantController {
	return &ParticipantController{app: app}
}

func (ctl *ParticipantController) ListContestParticipants(ctx *gin.Context, request api.ListContestParticipantsRequestObject) (api.ListContestParticipantsResponseObject, error) {
	out, err := ctl.app.ListParticipants(&listparticipants.Input{ContestID: request.Id})
	if err != nil {
		return nil, err
	}

	userIDs := make([]int32, 0, len(out.UserIDs))
	for _, id := range out.UserIDs {
		userIDs = append(userIDs, int32(id))
	}
	groupIDs := make([]int32, 0, len(out.GroupIDs))
	for _, id := range out.GroupIDs {
		groupIDs = append(groupIDs, int32(id))
	}

	return api.ListContestParticipants200JSONResponse{
		Data: api.AdminContestSpecContestParticipants{
			UserIds:  userIDs,
			GroupIds: groupIDs,
		},
		Success: true,
	}, nil
}

func (ctl *ParticipantController) AddUsersToContest(ctx *gin.Context, request api.AddUsersToContestRequestObject) (api.AddUsersToContestResponseObject, error) {
	userIDs := make([]int, 0, len(request.Body.UserIds))
	for _, id := range request.Body.UserIds {
		userIDs = append(userIDs, int(id))
	}
	if _, err := ctl.app.AddUsers(&addusers.Input{ContestID: request.Id, UserIDs: userIDs}); err != nil {
		if errors.Is(err, addusers.ErrorUserIDsRequired) {
			return api.AddUsersToContest400JSONResponse{Message: api.SharedAPIBadRequestResponseMessage(err.Error()), Success: false}, nil
		}
		return nil, err
	}
	return api.AddUsersToContest200JSONResponse{Success: true}, nil
}

func (ctl *ParticipantController) RemoveUsersFromContest(ctx *gin.Context, request api.RemoveUsersFromContestRequestObject) (api.RemoveUsersFromContestResponseObject, error) {
	userIDs := make([]int, 0, len(request.Body.UserIds))
	for _, id := range request.Body.UserIds {
		userIDs = append(userIDs, int(id))
	}
	if _, err := ctl.app.RemoveUsers(&removeusers.Input{ContestID: request.Id, UserIDs: userIDs}); err != nil {
		return nil, err
	}
	return api.RemoveUsersFromContest200JSONResponse{Success: true}, nil
}

func (ctl *ParticipantController) AddGroupsToContest(ctx *gin.Context, request api.AddGroupsToContestRequestObject) (api.AddGroupsToContestResponseObject, error) {
	groupIDs := make([]int, 0, len(request.Body.GroupIds))
	for _, id := range request.Body.GroupIds {
		groupIDs = append(groupIDs, int(id))
	}
	if _, err := ctl.app.AddGroups(&addgroups.Input{ContestID: request.Id, GroupIDs: groupIDs}); err != nil {
		if errors.Is(err, addgroups.ErrorGroupIDsRequired) {
			return api.AddGroupsToContest400JSONResponse{Message: api.SharedAPIBadRequestResponseMessage(err.Error()), Success: false}, nil
		}
		return nil, err
	}
	return api.AddGroupsToContest200JSONResponse{Success: true}, nil
}

func (ctl *ParticipantController) RemoveGroupsFromContest(ctx *gin.Context, request api.RemoveGroupsFromContestRequestObject) (api.RemoveGroupsFromContestResponseObject, error) {
	groupIDs := make([]int, 0, len(request.Body.GroupIds))
	for _, id := range request.Body.GroupIds {
		groupIDs = append(groupIDs, int(id))
	}
	if _, err := ctl.app.RemoveGroups(&removegroups.Input{ContestID: request.Id, GroupIDs: groupIDs}); err != nil {
		return nil, err
	}
	return api.RemoveGroupsFromContest200JSONResponse{Success: true}, nil
}
