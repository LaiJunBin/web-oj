import "./model.tsp";
import "@typespec/http";
import "@typespec/openapi";

using TypeSpec.OpenAPI;
using TypeSpec.Http;
using Shared;

@tag("GradingItems")
namespace Admin.ProjectSpec.GradingItemSpec;

model CreateGradingItemRequest {
    name: string;
    label: string;
    description?: string;
    score: float64;
    scoringStrategy: GradingItemScoringStrategy;
    outputFormat: GradingItemOutputFormat;
    resultParser?: string;
    depends: int32[];
}

model UpdateGradingItemRequest {
    name?: string;
    label?: string;
    description?: string;
    score?: float64;
    scoringStrategy?: GradingItemScoringStrategy;
    outputFormat?: GradingItemOutputFormat;
    resultParser?: string;
    depends?: int32[];
}

model MoveGradingItemOrderRequest {
    order: int32;
}

interface Interface {
    @post
    @usePermission(AuthAdmin)
    @route("/projects/{id}/grading-items")
    @operationId("createGradingItem")
    create(
        @path id: int32,
        @body payload: CreateGradingItemRequest,
    ): APINoDataCreateResponse;

    @get
    @usePermission(AuthRead)
    @route("/grading-items/{id}")
    @operationId("getGradingItem")
    get(@path id: int32): APIGetResponse<GradingItemDetail>;

    @get
    @usePermission(AuthRead)
    @route("/projects/{id}/grading-items")
    @operationId("listGradingItem")
    list(@path id: int32): APIListResponse<GradingItem>;

    @patch
    @usePermission(AuthAdmin)
    @route("/grading-items/{id}")
    @operationId("updateGradingItem")
    update(
        @path id: int32,
        @body payload: UpdateGradingItemRequest,
    ): APINoDataUpdateResponse;

    @delete
    @usePermission(AuthAdmin)
    @route("/grading-items/{id}")
    @operationId("deleteGradingItem")
    delete(@path id: int32): APINoDataDeleteResponse;

    @post
    @usePermission(AuthAdmin)
    @route("/grading-items/{id}/move")
    @operationId("moveGradingItemOrder")
    move(
        @path id: int32,
        @body payload: MoveGradingItemOrderRequest,
    ): APINoDataResponseOrAvailableErrors;
}
