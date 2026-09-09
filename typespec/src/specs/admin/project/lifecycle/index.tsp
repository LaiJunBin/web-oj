import "./model.tsp";
import "@typespec/http";
import "@typespec/openapi";

using TypeSpec.OpenAPI;
using TypeSpec.Http;
using Shared;

@tag("Lifecycles")
namespace Admin.ProjectSpec.LifecycleSpec;

model CreateLifecycleRequest {
    type: LifecycleType;
    name: string;
    command: string;
    description?: string;
    runsOn: RunsOn;
}

model UpdateLifecycleRequest {
    type?: LifecycleType;
    name?: string;
    command?: string;
    description?: string;
    runsOn?: RunsOn;
}

interface Interface {
    @post
    @usePermission(AuthAdmin)
    @route("/projects/{id}/lifecycles")
    @operationId("createLifecycle")
    create(
        @path id: int32,
        @body payload: CreateLifecycleRequest,
    ): APINoDataCreateResponse;

    @get
    @usePermission(AuthRead)
    @route("/lifecycles/{id}")
    @operationId("getLifecycle")
    get(@path id: int32): APIGetResponse<LifecycleDetail>;

    @patch
    @usePermission(AuthAdmin)
    @route("/lifecycles/{id}")
    @operationId("updateLifecycle")
    update(
        @path id: int32,
        @body payload: UpdateLifecycleRequest,
    ): APINoDataUpdateResponse;

    @delete
    @usePermission(AuthAdmin)
    @route("/lifecycles/{id}")
    @operationId("deleteLifecycle")
    delete(@path id: int32): APINoDataDeleteResponse;
}
