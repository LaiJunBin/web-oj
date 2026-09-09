import "./model.tsp";
import "@typespec/http";
import "@typespec/openapi";

using TypeSpec.OpenAPI;
using TypeSpec.Http;
using Shared;

@tag("Judgments")
namespace Admin.JudgmentSpec;

model CancelJudgeRequest {
    reason: string;
}

model SetJudgmentStaffNoteRequest {
    @doc("The note as it should read afterwards. An empty string clears it. It is shown to the contestant on their submission, alongside — never replacing — the system's reason, so it must not carry commands, paths or internal details.")
    staffNote: string;
}

// EDITING THIS FILE: never write the word that TypeSpec uses for a declared
// scope (n-a-m-e-s-p-a-c-e) anywhere below this line, comments included.
//
// build-utils/utils.js scrapes each spec's interface-level route with
// /(@route.*)[\s\S]*<that word>/. This file has no interface-level route — its
// operations carry full paths — and the only reason the scrape does not pick up
// the first operation's @route by mistake is that the word never appears after
// it. Introducing one makes the generator emit @route("/rejudges") onto
// JudgmentSpecImpl in main.tsp, and every route below is then produced doubled:
// /rejudges/rejudges, /rejudges/judgments/{id}/cancel, and so on. The compiler
// reports nothing; only the emitted openapi.yaml shows it.
interface Interface {
    @get
    @usePermission(AuthRead)
    @route("/rejudges")
    @operationId("listRejudges")
    listRejudges(): APIGetResponse<Rejudge[]>;

    // Lives in this spec rather than under ContestProjectSpec because the model
    // it returns is defined here — the same arrangement SubmissionSpec uses for
    // /contest-projects/{id}/submissions.
    @get
    @usePermission(AuthRead)
    @route("/contest-projects/{id}/rejudges")
    @operationId("listRejudgesByContestProject")
    listRejudgesByContestProject(@path id: int32): APIGetResponse<Rejudge[]>;

    @get
    @usePermission(AuthRead)
    @route("/rejudges/{id}")
    @operationId("getRejudge")
    getRejudge(@path id: string): APIGetResponse<RejudgeDetail>;

    // The rejudge path parameter must stay named `id` to match the sibling
    // /rejudges/{id} routes: the Gin router this spec generates onto refuses two
    // different wildcard names in the same path position, and panics at startup
    // rather than at request time.
    @get
    @usePermission(AuthRead)
    @route("/rejudges/{id}/submissions/{submissionId}/comparison")
    @operationId("getRejudgeSubmissionComparison")
    getRejudgeSubmissionComparison(
        @path id: string,
        @path submissionId: string,
    ): APIGetResponse<SubmissionComparisonDetail>;

    @post
    @usePermission(AuthAdmin)
    @route("/rejudges/{id}/cancel")
    @operationId("cancelRejudge")
    cancelRejudge(@path id: string): APINoDataResponseOrAvailableErrors;

    @post
    @usePermission(AuthAdmin)
    @route("/rejudge/{id}/approve")
    @operationId("approveRejudge")
    approveRejudge(@path id: string): APINoDataResponseOrAvailableErrors;

    @post
    @usePermission(AuthAdmin)
    @route("/rejudge/{id}/reject")
    @operationId("rejectRejudge")
    rejectRejudge(@path id: string): APINoDataResponseOrAvailableErrors;

    @post
    @usePermission(AuthAdmin)
    @route("/judgments/{id}/cancel")
    @operationId("cancelJudge")
    cancelJudge(@path id: string, @body payload: CancelJudgeRequest): APINoDataResponseOrAvailableErrors;

    // A staff note is the judge's own explanation to the contestant, written
    // after looking into a judgment (typically an Error). Last write wins and
    // there is no history: one contest, one judging desk. It never overwrites
    // the system's reason — the two are shown side by side.
    @put
    @usePermission(AuthAdmin)
    @route("/judgments/{id}/staff-note")
    @operationId("setJudgmentStaffNote")
    setJudgmentStaffNote(@path id: string, @body payload: SetJudgmentStaffNoteRequest): APINoDataResponseOrAvailableErrors;

    // Newest first, across every contest project: this answers "what just broke",
    // which is a question asked before you know where to look.
    @get
    @usePermission(AuthRead)
    @route("/judgment-errors")
    @operationId("listJudgmentErrors")
    listJudgmentErrors(@query limit?: int32): APIGetResponse<JudgmentError[]>;

    // The same listing pushed on a timer. Ticker-driven rather than
    // event-driven, like the judge-queue stream: there is no failure event to
    // subscribe to here, only a query over stored judgments. Each frame carries
    // the identical envelope the polled endpoint returns, so one client parser
    // serves both.
    @get
    @usePermission(AuthRead)
    @route("/judgment-errors/stream")
    @operationId("streamJudgmentErrors")
    streamJudgmentErrors(@query limit?: int32): SSEResponse<"judgmentErrors">;
}
