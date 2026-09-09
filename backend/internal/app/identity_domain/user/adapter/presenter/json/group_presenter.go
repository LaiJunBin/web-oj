package jsonpresenter

import (
	"judge-backend/api"
	getgroup "judge-backend/internal/app/identity_domain/user/application/usecase/group/get_group"
	listgroups "judge-backend/internal/app/identity_domain/user/application/usecase/group/list_groups"
	"judge-backend/internal/app/identity_domain/user/domain/entity/group"
)

func toGroup(g *group.Group) api.AdminGroupSpecGroup {
	return api.AdminGroupSpecGroup{
		Id:          int32(g.GetID()),
		Name:        g.GetName(),
		Description: g.GetDescription(),
	}
}

func GetGroup(out *getgroup.Output) api.GetGroup200JSONResponse {
	userIDs := make([]int32, 0, len(out.UserIDs))
	for _, id := range out.UserIDs {
		userIDs = append(userIDs, int32(id))
	}
	return api.GetGroup200JSONResponse{
		Data: api.AdminGroupSpecGroupDetail{
			Id:          int32(out.Group.GetID()),
			Name:        out.Group.GetName(),
			Description: out.Group.GetDescription(),
			UserIds:     userIDs,
		},
		Success: true,
	}
}

func ListGroups(out *listgroups.Output) api.ListGroup200JSONResponse {
	groups := make([]api.AdminGroupSpecGroup, 0, len(out.Groups))
	for _, g := range out.Groups {
		groups = append(groups, toGroup(g))
	}
	return api.ListGroup200JSONResponse{
		Data:    groups,
		Success: true,
	}
}
