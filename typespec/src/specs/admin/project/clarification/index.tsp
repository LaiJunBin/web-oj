import "./model.tsp";
import "@typespec/http";
import "@typespec/openapi";

using TypeSpec.OpenAPI;
using TypeSpec.Http;
using Shared;

@tag("Clarifications")
namespace Admin.ProjectSpec.ClarificationSpec;

model CreateClarificationRequest {
    subject: string;
    content: string;
    receiverUserId?: int32;
    sendToAll: boolean;
}

model ReplyClarificationRequest {
    subject?: string;
    content: string;
    sendToAll: boolean;
}

interface Interface {
    @get
    @usePermission(AuthAdmin)
    @route("/contest-projects/{id}/clarifications")
    @operationId("listClarifications")
    list(@path id: int32): APIListResponse<ClarificationInfo>;

    @post
    @usePermission(AuthAdmin)
    @route("/contest-projects/{id}/clarifications")
    @operationId("sendClarification")
    create(
        @path id: int32,
        @body payload: CreateClarificationRequest,
    ): APINoDataCreateResponse;

    @get
    @usePermission(AuthAdmin)
    @route("/clarifications/{id}")
    @operationId("getClarification")
    get(@path id: int32): APIGetResponse<Clarification>;

    @post
    @usePermission(AuthAdmin)
    @route("/clarifications/{id}/reply")
    @operationId("replyClarificationToUser")
    reply(
        @path id: int32,
        @body payload: ReplyClarificationRequest,
    ): APINoDataCreateResponse;

    @put
    @usePermission(AuthAdmin)
    @route("/clarifications/{id}/mark-as-answered")
    @operationId("markClarificationAsAnswered")
    markAsAnswered(@path id: int32): APINoDataUpdateResponse;

    @put
    @usePermission(AuthAdmin)
    @route("/clarifications/{id}/mark-as-unanswered")
    @operationId("markClarificationAsUnanswered")
    markAsUnanswered(@path id: int32): APINoDataUpdateResponse;
}
