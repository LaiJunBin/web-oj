import "./responses.tsp";
import "@typespec/http";
import "@typespec/openapi";

using TypeSpec.Http;
using TypeSpec.OpenAPI;
namespace Shared;

interface ResourcesInterface<
    OperationId extends valueof string,
    TList,
    TGet,
    TCreate,
    TUpdate,
    CreateRequest,
    UpdateRequest = null
> {
    @post
    @usePermission(AuthAdmin)
    @operationId("create${OperationId}")
    create(@body payload: CreateRequest): APICreateResponse<TCreate>;

    @get
    @usePermission(AuthRead)
    @operationId("get${OperationId}")
    get(@path id: int32): APIGetResponse<TGet>;

    @get
    @usePermission(AuthRead)
    @operationId("list${OperationId}")
    list(): APIListResponse<TList>;

    @patch
    @usePermission(AuthAdmin)
    @operationId("update${OperationId}")
    update(@path id: int32, @body payload: UpdateRequest): APIUpdateResponse<TUpdate>;

    @delete
    @usePermission(AuthAdmin)
    @operationId("delete${OperationId}")
    delete(@path id: int32): APIDeleteResponse;
}

interface FormResourcesInterface<
    OperationId extends valueof string,
    TList,
    TGet,
    TCreate,
    TUpdate,
    CreateRequest,
    UpdateRequest = null
> {
    @post
    @usePermission(AuthAdmin)
    @operationId("create${OperationId}")
    create(
        @header contentType: "multipart/form-data",
        @body payload: CreateRequest,
    ): APICreateResponse<TCreate>;

    @get
    @usePermission(AuthRead)
    @operationId("get${OperationId}")
    get(@path id: int32): APIGetResponse<TGet>;

    @get
    @usePermission(AuthRead)
    @operationId("list${OperationId}")
    list(): APIListResponse<TList>;

    @patch
    @usePermission(AuthAdmin)
    @operationId("update${OperationId}")
    update(
        @header contentType: "multipart/form-data",
        @path id: int32,
        @body payload: UpdateRequest,
    ): APIUpdateResponse<TUpdate>;

    @delete
    @usePermission(AuthAdmin)
    @operationId("delete${OperationId}")
    delete(@path id: int32): APIDeleteResponse;
}

interface CqrsResourcesInterface<
    OperationId extends valueof string,
    TList,
    TGet,
    CreateRequest,
    UpdateRequest = null
> {
    @post
    @usePermission(AuthAdmin)
    @operationId("create${OperationId}")
    create(@body payload: CreateRequest): APINoDataCreateResponse;

    @get
    @usePermission(AuthRead)
    @operationId("get${OperationId}")
    get(@path id: int32): APIGetResponse<TGet>;

    @get
    @usePermission(AuthRead)
    @operationId("list${OperationId}")
    list(): APIListResponse<TList>;

    @patch
    @usePermission(AuthAdmin)
    @operationId("update${OperationId}")
    update(@path id: int32, @body payload: UpdateRequest): APINoDataUpdateResponse;

    @delete
    @usePermission(AuthAdmin)
    @operationId("delete${OperationId}")
    delete(@path id: int32): APINoDataDeleteResponse;
}

interface FormCqrsResourcesInterface<
    OperationId extends valueof string,
    TList,
    TGet,
    CreateRequest,
    UpdateRequest = null
> {
    @post
    @usePermission(AuthAdmin)
    @operationId("create${OperationId}")
    create(
        @header contentType: "multipart/form-data",
        @body payload: CreateRequest,
    ): APINoDataCreateResponse;

    @get
    @usePermission(AuthRead)
    @operationId("get${OperationId}")
    get(@path id: int32): APIGetResponse<TGet>;

    @get
    @usePermission(AuthRead)
    @operationId("list${OperationId}")
    list(): APIListResponse<TList>;

    @patch
    @usePermission(AuthAdmin)
    @operationId("update${OperationId}")
    update(
        @header contentType: "multipart/form-data",
        @path id: int32,
        @body payload: UpdateRequest,
    ): APINoDataUpdateResponse;

    @delete
    @usePermission(AuthAdmin)
    @operationId("delete${OperationId}")
    delete(@path id: int32): APINoDataDeleteResponse;
}

