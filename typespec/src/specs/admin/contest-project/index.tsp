import "@typespec/http";
import "@typespec/openapi";
import "./model.tsp";

using TypeSpec.OpenAPI;
using TypeSpec.Http;
using Shared;

@route("/contest-projects")
@tag("ContestProject")
namespace Admin.ContestProjectSpec;

model RejudgeContestProjectPayload {
    gradingItemIds?: int32[];
    reason: string;
}

interface Interface {
    @get
    @usePermission(AuthRead)
    @route("/{id}")
    @operationId("getContestProject")
    getContestProject(@path id: int32): APIGetResponse<ContestProject>;

    @post
    @usePermission(AuthAdmin)
    @route("/{id}/rejudge")
    @operationId("rejudgeContestProject")
    rejudgeContestProject(@path id: int32, @body payload: RejudgeContestProjectPayload): APINoDataResponseOrAvailableErrors;

    @get
    @usePermission(AuthRead)
    @route("/{id}/user-scores")
    @operationId("getUserScores")
    getUserScores(@path id: int32): APIGetResponse<UserScore[]>;

    @get
    @usePermission(AuthRead)
    @route("/{id}/score-detail/export")
    @operationId("exportContestProjectScoreDetail")
    exportScoreDetail(@path id: int32): DownloadFileResponse;

    @get
    @usePermission(AuthRead)
    @route("/{id}/judgment-analytics")
    @operationId("getContestProjectJudgmentAnalytics")
    getJudgmentAnalytics(@path id: int32): APIGetResponse<JudgmentAnalytics>;
}


