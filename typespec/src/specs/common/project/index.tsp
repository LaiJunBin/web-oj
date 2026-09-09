import "./model.tsp";
import "@typespec/http";
import "@typespec/openapi";

using TypeSpec.OpenAPI;
using TypeSpec.Http;
using Shared;

@tag("PublicProject")
namespace Common.ProjectSpec;

interface Interface {
    @get
    @route("/public/contest-projects/{id}")
    @operationId("getPublicProject")
    getPublicProject(@path id: int32): APIGetResponse<Project>;

    @get
    @route("/public/contest-projects/{id}/assets/{assetId}")
    @operationId("downloadPublicProjectAsset")
    DownloadPublicProjectAsset(
        @path id: int32,
        @path assetId: string,
    ): DownloadFileResponse;

    @get
    @route("/public/contest-projects/{id}/scoreboard")
    @operationId("getPublicProjectScoreboard")
    getPublicProjectScoreboard(@path id: int32): APIGetResponse<ProjectScoreboard>;

    @get
    @route("/public/contest-projects/{id}/stream-time")
    @operationId("streamPublicProjectTime")
    streamPublicProjectTime(@path id: int32): SSEResponse<123>;

    @get
    @route("/public/contest-projects/{id}/stream-scoreboard")
    @operationId("streamPublicProjectScoreboard")
    streamPublicProjectScoreboard(@path id: int32): SSEResponse<"[]">;
}
