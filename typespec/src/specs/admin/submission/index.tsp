import "./model.tsp";
import "@typespec/http";
import "@typespec/openapi";

using TypeSpec.OpenAPI;
using TypeSpec.Http;
using Shared;

@tag("Submissions")
namespace Admin.SubmissionSpec;

model RejudgeRequest {
    reason: string;
}

interface Interface {
    @get
    @usePermission(AuthRead)
    @route("/submissions/{id}")
    @operationId("getSubmission")
    getSubmission(@path id: string): APIGetResponse<SubmissionDetail>;

    @get
    @usePermission(AuthRead)
    @route("/contest-projects/{id}/submissions")
    @operationId("getSubmissionsByContestProject")
    getSubmissionsByContestProject(@path id: int32): APIListResponse<Submission>;

    @get
    @usePermission(AuthRead)
    @route("/contest-projects/{id}/stream-submissions")
    @operationId("streamSubmissionsByContestProject")
    streamSubmissionsByContestProject(@path id: int32): SSEResponse<"submission">;

    @get
    @usePermission(AuthRead)
    @route("/users/{id}/submissions")
    @operationId("getSubmissionsByUser")
    getSubmissionsByUser(
        @path id: int32,
    ): APIListResponse<Submission>;

    // A time window and an optional filter, not a resource identity, so these
    // are query parameters rather than another path segment. Buckets come back
    // oldest first and every minute in the window is present, so a chart can
    // plot the series without filling gaps itself.
    @get
    @usePermission(AuthRead)
    @route("/submission-rate")
    @operationId("getSubmissionRate")
    getSubmissionRate(
        @query minutes?: int32,
        @query contestProjectId?: int32,
    ): APIGetResponse<SubmissionRatePoint[]>;

    // The same series pushed on a timer, with the identical envelope. Its tick
    // is slower than the other status streams on purpose: the buckets have
    // minute resolution, so pushing every couple of seconds would re-send the
    // same numbers several times over before any of them could change.
    @get
    @usePermission(AuthRead)
    @route("/submission-rate/stream")
    @operationId("streamSubmissionRate")
    streamSubmissionRate(
        @query minutes?: int32,
        @query contestProjectId?: int32,
    ): SSEResponse<"submissionRate">;

    @post
    @usePermission(AuthAdmin)
    @route("/submissions/{id}/rejudge")
    @operationId("rejudgeSubmission")
    rejudgeSubmission(@path id: string, @body payload: RejudgeRequest): APINoDataResponseOrAvailableErrors;

    @post
    @usePermission(AuthAdmin)
    @route("/submissions/{id}/ignore")
    @operationId("ignoreSubmission")
    ignoreSubmission(@path id: string): APINoDataResponseOrAvailableErrors;

    @delete
    @usePermission(AuthAdmin)
    @route("/submissions/{id}/ignore")
    @operationId("unignoreSubmission")
    unignoreSubmission(@path id: string): APINoDataResponseOrAvailableErrors;
}
