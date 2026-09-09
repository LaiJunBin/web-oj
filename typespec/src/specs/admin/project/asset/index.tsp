import "@typespec/http";
import "@typespec/openapi";

using TypeSpec.OpenAPI;
using TypeSpec.Http;
using Shared;

@tag("Assets")
namespace Admin.ProjectSpec.AssetSpec;

model UploadAssetRequest {
    name: string;
    file: bytes;
}

model UpdateAssetRequest {
    name?: string;
    file: bytes;
    isEnabled?: boolean;
}

alias UploadAssetResponse = APICreateResponse<string>;

interface Interface {
    @get
    @usePermission(AuthRead)
    @route("/assets/{id}")
    @operationId("getAsset")
    get(@path id: string): APIGetResponse<Asset>;

    @put
    @usePermission(AuthAdmin)
    @route("/projects/{id}/assets")
    @operationId("uploadAsset")
    upload(
        @header contentType: "multipart/form-data",
        @path id: int32,
        @body payload: UploadAssetRequest,
    ): UploadAssetResponse;

    @patch
    @usePermission(AuthAdmin)
    @route("/assets/{id}")
    @operationId("updateAsset")
    update(
        @header contentType: "multipart/form-data",
        @path id: string,
        @body payload: UpdateAssetRequest,
    ): APINoDataUpdateResponse;

    @delete
    @usePermission(AuthAdmin)
    @route("/assets/{id}")
    @operationId("deleteAsset")
    delete(@path id: string): APINoDataDeleteResponse;

    @get
    @usePermission(AuthRead)
    @route("/assets/{id}/download")
    @operationId("downloadAsset")
    download(@path id: string): DownloadFileResponse;
}
