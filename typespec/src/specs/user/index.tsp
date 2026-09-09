import "./model.tsp";
import "@typespec/http";
import "@typespec/openapi";

using TypeSpec.OpenAPI;
using TypeSpec.Http;
using Shared;
using Common.AuthSpec;

@tag("User")
namespace User.UserSpec;

model SubmitProjectRequest {
    project: bytes;
    gradingItemIDs: int32[];
    dynamicFields: string;
}

model SendProjectClarificationToAdminRequest {
    subject: string;
    content: string;
}

model ReplyClarificationToAdminRequest {
    content: string;
}

interface Interface {
    @post
    @usePermission(AuthContestant)
    @route("/user/contest-projects/{id}/submit")
    @operationId("submitUserProject")
    submitUserProject(
        @header contentType: "multipart/form-data",
        @path id: int32,
        @body payload: SubmitProjectRequest,
    ): APINoDataResponseOrAvailableErrors | APITooManyRequestsResponse;

    @post
    @usePermission(AuthContestant)
    @route("/user/submissions/{submissionId}/cancel")
    @operationId("cancelUserSubmission")
    cancelUserSubmission(
        @path submissionId: string,
    ): APINoDataResponseOrAvailableErrors;

    @get
    @usePermission(AuthContestant)
    @route("/user/contest-projects/{id}/submissions")
    @operationId("listUserProjectSubmissions")
    listUserProjectSubmissions(@path id: int32): APIGetResponse<Submission[]>;

    @get
    @usePermission(AuthContestant)
    @route("/user/submissions/{submissionId}")
    @operationId("getUserSubmission")
    getUserSubmission(
        @path submissionId: string,
    ): APIGetResponse<SubmissionDetail>;

    @get
    @usePermission(AuthContestant)
    @route("/user/contest-projects/{id}/clarifications")
    @operationId("listUserProjectClarifications")
    listUserProjectClarifications(
        @path id: int32,
    ): APIListResponse<Clarification>;

    @get
    @usePermission(AuthContestant)
    @route("/user/clarifications/{clarificationId}")
    @operationId("getUserClarification")
    getUserClarification(@path clarificationId: int32): APIGetResponse<ClarificationDetail>;

    @post
    @usePermission(AuthContestant)
    @route("/user/contest-projects/{id}/clarifications")
    @operationId("sendUserProjectClarificationToAdmin")
    sendUserProjectClarificationToAdmin(
        @path id: int32,
        @body payload: SendProjectClarificationToAdminRequest,
    ): APICreateResponse<{id: int32}>;

    @post
    @usePermission(AuthContestant)
    @route("/user/clarifications/{clarificationId}/reply")
    @operationId("replyUserClarificationToAdmin")
    replyUserClarificationToAdmin(
        @path clarificationId: int32,
        @body payload: ReplyClarificationToAdminRequest,
    ): APICreateResponse<{id: int32}>;

    @get
    @usePermission(AuthContestant)
    @route("/user/stream-submissions/{id}")
    @operationId("streamUserSubmission")
    streamUserSubmission(@path id: string): SSEResponse<"test">;

    @get
    @usePermission(AuthContestant)
    @route("/user/contest-projects/{id}/stream-clarification")
    @operationId("streamUserProjectClarification")
    streamUserProjectClarification(@path id: int32): SSEResponse<"{}">;
}
