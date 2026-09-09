import "@typespec/http";
import "@typespec/openapi";
import "./model.tsp";

using TypeSpec.OpenAPI;
using TypeSpec.Http;
using Shared;

@route("/bundles")
@tag("Bundles")
namespace Admin.BundleSpec;

model PreviewBundleRequest {
    /** 完整匯出檔（.zip）：manifest.json、contest.json、projects/、images/ */
    bundle: bytes;
}

model ImportBundleRequest {
    bundle: bytes;

    /** 同 tag 但內容不同的映像，強制覆蓋。未指定則該項會被列為 blocker */
    force?: boolean;
}

interface Interface {
    /** 匯出整場競賽（.zip）：一場競賽、它的所有題目、以及這些題目參照的映像。
     *
     * mode 決定映像以哪種形態隨行：source 是建置原始檔（KB 級，匯入端要建置、
     * 要能連外拉 base image）；archive 是 docker save 的映像檔（可能數百 MB，
     * 匯入端不需網路也不需建置，適合離線考場）。archive 需要所有映像都已建置
     * 完成，否則回 400 並列出還沒建置的 tag。 */
    @get
    @usePermission(AuthRead)
    @route("/export/contests/{id}")
    @operationId("exportBundle")
    exportBundle(
        @path id: int32,
        @query mode: "source" | "archive",
    ): DownloadFileResponse;

    /** 檢視匯出檔會做什麼，不寫入任何東西。
     *
     * 匯入分兩步的原因：完整匯出檔會建立映像、題目與競賽，其中有些會覆蓋既有的
     * 東西。先讓操作者看到「哪些是新建、哪些已存在、哪些會被覆蓋、影響誰」，
     * 再決定要不要執行——這比任何自動判斷規則都可靠，對沒有開發背景的人也
     * 比較好懂：它把決定變成選擇題。 */
    @post
    @usePermission(AuthAdmin)
    @route("/preview")
    @operationId("previewBundle")
    preview(
        @header contentType: "multipart/form-data",
        @body payload: PreviewBundleRequest,
    ): APICreateResponse<BundlePreview>;

    /** 執行匯入。順序固定為映像 → 題目 → 競賽，因為後者參照前者。
     *
     * 匯出檔內含會在評測主機上建置並執行的 Dockerfile 與指令，因此限管理員。 */
    @post
    @usePermission(AuthAdmin)
    @route("/import")
    @operationId("importBundle")
    importBundle(
        @header contentType: "multipart/form-data",
        @body payload: ImportBundleRequest,
    ): APICreateResponse<BundleImportReport> | APIConflictResponse;
}
