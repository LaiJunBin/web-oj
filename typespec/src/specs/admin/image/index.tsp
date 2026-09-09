import "./model.tsp";
import "@typespec/http";
import "@typespec/openapi";

using TypeSpec.Http;
using TypeSpec.OpenAPI;
using Shared;

@route("/images")
@tag("Images")
namespace Admin.ImageSpec;

model ImportImageRequest {
    /** 映像匯出檔（.zip）。內含 image.json，以及 context.zip 或 image.tar 其中之一 */
    bundle: bytes;

    /** 同 tag 但內容不同時，覆蓋既有映像。未指定則以 409 拒絕 */
    force?: boolean;
}

model ImportImageResult {
    tag: string;

    /** source：通常會排入建置(以 buildQueued 為準——內容與現有映像完全相同時直接沿用、不重建);archive：映像已直接載入,可立即使用 */
    mode: string;

    buildQueued: boolean;
}

model CreateImageRequest {
    tag: string;
    mountPath: string;
    description?: string;
    noCache?: boolean;
    project: bytes;
}

model UpdateImageRequest {
    // tag?: string;
    mountPath?: string;
    description?: string;
    noCache?: boolean;
    project?: bytes;
}

alias CancelBuildResponse = APINoDataResponseOrAvailableErrors;
alias RebuildImageResponse = APINoDataResponseOrAvailableErrors;
alias StreamLogsBody = "[\"Step 1/11 : FROM nginx\",\"\\n\"]";
alias StreamLogsResponse = SSEResponse<StreamLogsBody>;

interface Interface {
    @get
    @usePermission(AuthRead)
    @operationId("listImage")
    list(): APIListResponse<ImageInfo>;

    @post
    @usePermission(AuthAdmin)
    @operationId("createImage")
    create(
        @header contentType: "multipart/form-data",
        @body payload: CreateImageRequest,
    ): APICreateResponse<string>;

    @get
    @usePermission(AuthRead)
    @operationId("getImage")
    get(@path tag: string): APIGetResponse<Image>;

    @patch
    @usePermission(AuthAdmin)
    @operationId("updateImage")
    update(
        @header contentType: "multipart/form-data",
        @path tag: string,
        @body payload: UpdateImageRequest,
    ): APINoDataUpdateResponse;

    @delete
    @usePermission(AuthAdmin)
    @operationId("deleteImage")
    delete(@path tag: string): APINoDataDeleteResponse;

    @delete
    @usePermission(AuthAdmin)
    @route("/{tag}/build")
    @operationId("cancelBuild")
    cancelBuild(@path tag: string): CancelBuildResponse;

    @post
    @usePermission(AuthAdmin)
    @route("/{tag}/rebuild")
    @operationId("rebuild")
    rebuild(@path tag: string): RebuildImageResponse;

    /** 匯出映像（.zip）。
     *
     * mode=source 帶的是建置用的原始檔（context.zip），體積小、可讀可改，
     * 但匯入端要能連外拉 base image 才建得起來。
     * mode=archive 帶的是 docker save 的映像檔（image.tar），可能數百 MB，
     * 但匯入端不需網路也不需建置，適合離線考場。
     * archive 需要映像已建置完成，否則回 400。 */
    @get
    @usePermission(AuthRead)
    @route("/{tag}/export")
    @operationId("exportImage")
    exportImage(
        @path tag: string,
        @query mode: "source" | "archive",
    ): DownloadFileResponse;

    /** 匯入映像匯出檔。檔案內含會在評測主機上建置並執行的 Dockerfile，因此限管理員。 */
    @post
    @usePermission(AuthAdmin)
    @route("/import")
    @operationId("importImage")
    importImage(
        @header contentType: "multipart/form-data",
        @body payload: ImportImageRequest,
    ): APICreateResponse<ImportImageResult> | APIConflictResponse;
}
