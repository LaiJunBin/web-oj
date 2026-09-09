import "./model.tsp";
import "@typespec/http";
import "@typespec/openapi";

using TypeSpec.OpenAPI;
using TypeSpec.Http;
using Shared;

@tag("Judge-workers")
namespace Admin.JudgeWorkerSpec;

model CreateJudgeWorkerRequest {
    judgeOrder: JudgeOrder[];
}

interface Interface {
    @get
    @usePermission(AuthRead)
    @route("/judge-workers")
    @operationId("listJudgeWorker")
    listJudgeWorker(): APIListResponse<JudgeWorker>;

    @get
    @usePermission(AuthRead)
    @route("/judge-priority-strategies")
    @operationId("listJudgePriorityStrategies")
    listJudgePriorityStrategies(): APIListResponse<string>;

    @get
    @usePermission(AuthRead)
    @route("/system-defaults")
    @operationId("getSystemDefaults")
    getSystemDefaults(): APIGetResponse<SystemDefaults>;

    @get
    @usePermission(AuthRead)
    @route("/judge-workers/{id}")
    @operationId("getJudgeWorker")
    getJudgeWorker(@path id: string): APIGetResponse<JudgeWorker>;

    @post
    @usePermission(AuthAdmin)
    @route("/judge-workers/{id}/stop")
    @operationId("stopJudgeWorker")
    stopJudgeWorker(@path id: string): APINoDataResponseOrAvailableErrors;

    @post
    @usePermission(AuthAdmin)
    @route("/judge-workers/{id}/start")
    @operationId("startJudgeWorker")
    startJudgeWorker(@path id: string): APINoDataResponseOrAvailableErrors;

    @post
    @usePermission(AuthAdmin)
    @route("/judge-workers")
    @operationId("addJudgeWorker")
    addJudgeWorker(@body payload: CreateJudgeWorkerRequest): APICreateResponse<string>;

    @delete
    @usePermission(AuthAdmin)
    @route("/judge-workers/{id}")
    @operationId("deleteJudgeWorker")
    deleteJudgeWorker(@path id: string): APINoDataResponseOrAvailableErrors;

    @get
    @usePermission(AuthRead)
    @route("/judge-workers/stream")
    @operationId("streamJudgeWorkers")
    streamJudgeWorkers(): SSEResponse<"judgeWorkers">;

    @get
    @usePermission(AuthRead)
    @route("/judge-workers/{id}/stream-logs")
    @operationId("streamJudgeWorkerLogs")
    streamJudgeWorkerLogs(@path id: string): SSEResponse<"string">;

    @get
    @usePermission(AuthRead)
    @route("/judge-queue")
    @operationId("getJudgeQueue")
    getJudgeQueue(): APIGetResponse<JudgeQueueStatus>;

    // Ticker-driven rather than event-driven, unlike the worker stream above:
    // oldestQueuedSeconds keeps growing while the queue sits untouched, so a
    // stream that only emitted on queue activity would freeze the one number
    // this page exists to watch.
    @get
    @usePermission(AuthRead)
    @route("/judge-queue/stream")
    @operationId("streamJudgeQueue")
    streamJudgeQueue(): SSEResponse<"judgeQueue">;
}
