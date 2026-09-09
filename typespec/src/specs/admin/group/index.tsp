import "./model.tsp";
import "@typespec/http";
import "@typespec/openapi";

using TypeSpec.OpenAPI;
using TypeSpec.Http;
using Shared;

@tag("Group")
@route("/groups")
namespace Admin.GroupSpec;

interface Interface
    extends CqrsResourcesInterface<
            "Group",
            Group,
            GroupDetail,
            CreateGroupRequest,
            UpdateGroupRequest
        > {
    @post
    @usePermission(AuthAdmin)
    @route("/{id}/users")
    @operationId("addUsersToGroup")
    @summary("Add users to a group")
    addUsers(
        @path id: int32,
        @body payload: ModifyGroupUsersRequest,
    ): APINoDataResponseOrAvailableErrors;

    @delete
    @usePermission(AuthAdmin)
    @route("/{id}/users")
    @operationId("removeUsersFromGroup")
    @summary("Remove users from a group")
    removeUsers(
        @path id: int32,
        @body payload: ModifyGroupUsersRequest,
    ): APINoDataResponseOrAvailableErrors;
}
