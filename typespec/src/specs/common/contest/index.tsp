import "./model.tsp";
import "@typespec/http";
import "@typespec/openapi";

using TypeSpec.OpenAPI;
using TypeSpec.Http;
using Shared;

@tag("PublicContest")
namespace Common.ContestSpec;

interface Interface {
    @get
    @route("/public/contests")
    @operationId("listPublicContests")
    listPublicContests(): APIListResponse<Contest>;

    @get
    @route("/public/contests/{contestId}")
    @operationId("getPublicContest")
    getPublicContest(contestId: int32): APIGetResponse<ContestDetail>;

    @get
    @route("/public/contests/{contestId}/schedule")
    @operationId("getPublicContestSchedule")
    getPublicContestSchedule(contestId: int32): APIGetResponse<ContestSchedule>;
}
