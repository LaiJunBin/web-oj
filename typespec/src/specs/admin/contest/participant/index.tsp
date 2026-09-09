import "@typespec/http";
import "@typespec/openapi";

using TypeSpec.OpenAPI;
using TypeSpec.Http;
using Shared;

@tag("ContestParticipants")
namespace Admin.ContestSpec.ParticipantSpec;

interface Interface {
    @get
    @usePermission(AuthRead)
    @route("/contests/{id}/participants")
    @operationId("listContestParticipants")
    @summary("List participants (users and groups) of a contest")
    list(@path id: int32): APIGetResponse<ContestParticipants>;

    @post
    @usePermission(AuthAdmin)
    @route("/contests/{id}/participants/users")
    @operationId("addUsersToContest")
    @summary("Add users to contest participants")
    addUsers(
        @path id: int32,
        @body payload: ModifyContestUsersRequest,
    ): APINoDataResponseOrAvailableErrors;

    @delete
    @usePermission(AuthAdmin)
    @route("/contests/{id}/participants/users")
    @operationId("removeUsersFromContest")
    @summary("Remove users from contest participants")
    removeUsers(
        @path id: int32,
        @body payload: ModifyContestUsersRequest,
    ): APINoDataResponseOrAvailableErrors;

    @post
    @usePermission(AuthAdmin)
    @route("/contests/{id}/participants/groups")
    @operationId("addGroupsToContest")
    @summary("Add groups to contest participants")
    addGroups(
        @path id: int32,
        @body payload: ModifyContestGroupsRequest,
    ): APINoDataResponseOrAvailableErrors;

    @delete
    @usePermission(AuthAdmin)
    @route("/contests/{id}/participants/groups")
    @operationId("removeGroupsFromContest")
    @summary("Remove groups from contest participants")
    removeGroups(
        @path id: int32,
        @body payload: ModifyContestGroupsRequest,
    ): APINoDataResponseOrAvailableErrors;
}
