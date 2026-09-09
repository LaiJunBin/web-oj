import "./model.tsp";
import "@typespec/http";
import "@typespec/openapi";

using TypeSpec.OpenAPI;
using TypeSpec.Http;
using Shared;

@tag("UserProject")
namespace User.ProjectSpec;

interface Interface {
    @get
    @useAuth(AuthContestant)
    @route("/user/contest-projects/{id}")
    @operationId("getUserProject")
    getUserProject(@path id: int32): APIGetResponse<ProjectDetail>;

    @get
    @useAuth(AuthContestant)
    @route("/user/contest-projects/{id}/assets/{assetId}")
    @operationId("downloadUserProjectAsset")
    DownloadUserProjectAsset(
        @path id: int32,
        @path assetId: string,
    ): DownloadFileResponse;

    @get
    @useAuth(AuthContestant)
    @route("/user/contest-projects/{id}/scoreboard")
    @operationId("getUserProjectScoreboard")
    getUserProjectScoreboard(@path id: int32): APIGetResponse<ProjectScoreboard>;

    @get
    @useAuth(AuthContestant)
    @route("/user/contest-projects/{id}/stream-time")
    @operationId("streamUserProjectTime")
    streamUserProjectTime(@path id: int32): SSEResponse<123>;

    @get
    @useAuth(AuthContestant)
    @route("/user/contest-projects/{id}/stream-scoreboard")
    @operationId("streamUserProjectScoreboard")
    streamUserProjectScoreboard(@path id: int32): SSEResponse<"[]">;
}
