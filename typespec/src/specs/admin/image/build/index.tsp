import "./model.tsp";
import "@typespec/http";
import "@typespec/openapi";

using TypeSpec.Http;
using TypeSpec.OpenAPI;
using Shared;

@route("/images/{tag}/builds")
@tag("Builds")
namespace Admin.BuildSpec;

alias StreamLogsBody = "[\"Step 1/11 : FROM nginx\",\"\\n\"]";
alias StreamLogsResponse = SSEResponse<StreamLogsBody>;

interface Interface {
    @get
    @usePermission(AuthRead)
    @route("/")
    @operationId("listBuild")
    list(@path tag: string): APIListResponse<BuildInfo>;

    @get
    @usePermission(AuthRead)
    @route("/{batch}")
    @operationId("getBuild")
    get(@path tag: string, @path batch: int32): APIGetResponse<Build>;

    @get
    @usePermission(AuthRead)
    @route("/{batch}/stream-logs")
    @operationId("streamLogs")
    streamLogs(@path tag: string, @path batch: int32): StreamLogsResponse;
}
