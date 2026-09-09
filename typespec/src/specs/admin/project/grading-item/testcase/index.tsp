import "./model.tsp";
import "@typespec/http";
import "@typespec/openapi";

using TypeSpec.OpenAPI;
using TypeSpec.Http;
using Shared;

@tag("Testcases")
namespace Admin.ProjectSpec.GradingItemSpec.TestcaseSpec;

model CreateTestcaseRequest {
    name: string;
    command: string;
    score: float64;
    isSample: boolean;
    depends: int32[];
}

model UpdateTestcaseRequest {
    name?: string;
    command?: string;
    score?: float64;
    isSample?: boolean;
    depends?: int32[];
}

interface Interface {
    @get
    @usePermission(AuthRead)
    @route("/projects/{id}/testcases")
    @operationId("listProjectTestcases")
    list(@path id: int32): APIListResponse<ProjectTestcaseGradingItem>;

    @post
    @usePermission(AuthAdmin)
    @route("/grading-items/{id}/testcases")
    @operationId("createTestcase")
    create(
        @path id: int32,
        @body payload: CreateTestcaseRequest,
    ): APINoDataCreateResponse;

    @get
    @usePermission(AuthRead)
    @route("/testcases/{id}")
    @operationId("getTestcase")
    get(@path id: int32): APIGetResponse<Testcase>;

    @patch
    @usePermission(AuthAdmin)
    @route("/testcases/{id}")
    @operationId("updateTestcase")
    update(
        @path id: int32,
        @body payload: UpdateTestcaseRequest,
    ): APINoDataUpdateResponse;

    @delete
    @usePermission(AuthAdmin)
    @route("/testcases/{id}")
    @operationId("deleteTestcase")
    delete(@path id: int32): APINoDataDeleteResponse;
}
