import "@typespec/http";
import "@typespec/openapi";

using TypeSpec.OpenAPI;
using TypeSpec.Http;
using Shared;
using Common.ContestSpec;

@tag("UserContest")
namespace User.ContestSpec;

interface Interface {
    @get
    @useAuth(AuthContestant)
    @route("/user/contests")
    @operationId("listUserContests")
    listUserContests(): APIListResponse<Contest>;

    @get
    @useAuth(AuthContestant)
    @route("/user/contests/{contestId}")
    @operationId("getUserContest")
    getUserContest(contestId: int32): APIGetResponse<ContestDetail>;

    @get
    @useAuth(AuthContestant)
    @route("/user/contests/{contestId}/schedule")
    @operationId("getUserContestSchedule")
    getUserContestSchedule(contestId: int32): APIGetResponse<ContestSchedule>;
}
