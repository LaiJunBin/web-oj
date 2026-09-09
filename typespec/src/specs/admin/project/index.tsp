import "@typespec/http";
import "@typespec/openapi";
import "./model.tsp";

using TypeSpec.OpenAPI;
using TypeSpec.Http;
using Shared;

@route("/projects")
@tag("Projects")
namespace Admin.ProjectSpec;

model CreateProjectRequest {
    name: string;
    description: string;
    projectImageTag: string;
    testImageTag: string;
    dynamicFields: ProjectDynamicField[];
    projectEnv: string;
    testEnv: string;
    resultParser?: string;
}

model UpdateProjectRequest {
    name?: string;
    description?: string;
    projectImageTag?: string;
    testImageTag?: string;
    dynamicFields?: ProjectDynamicField[];
    projectEnv?: string;
    testEnv?: string;
    resultParser?: string;
}

model PutGradingProjectRequest {
    project: bytes;
}

model ImportProjectRequest {
    /** 題目匯出檔（.zip）。內含 project.json，以及選用的 assets/ 與 grading-files.zip */
    bundle: bytes;
}

model ImportProjectResult {
    projectId: int32;

    /** 實際建立的題目名稱。與匯出檔內同名的題目已存在時會加上後綴，此處回傳最終名稱 */
    name: string;

    /** 這個題目參照的映像 tag，依序為專案端與測試端 */
    imageTag: string;

    testImageTag: string;

    /** 上列 tag 當中系統目前沒有的。非空表示題目已建立但還不能評測，需先建立並建置這些映像 */
    missingImageTags: string[];
}

model WriteGradingProjectFileRequest {
    content: string;
}

interface Interface
    extends CqrsResourcesInterface<
            "Project",
            Project,
            ProjectDetail,
            CreateProjectRequest,
            UpdateProjectRequest
        > {
    @put
    @usePermission(AuthAdmin)
    @route("/{id}/grading-project")
    @operationId("putGradingProject")
    putGradingProject(
        @header contentType: "multipart/form-data",
        @path id: int32,
        @body payload: PutGradingProjectRequest,
    ): APINoDataResponseOrAvailableErrors;

    @get
    @usePermission(AuthRead)
    @route("/{id}/grading-project/download")
    @operationId("downloadGradingProject")
    downloadGradingProject(@path id: int32): DownloadFileResponse;

    /** 匯出題目（.zip）：題目定義、評分項與測資、生命週期、動態欄位，以及附件與評分檔案。
     * 不含映像內容，只以 tag 參照 —— 映像另由映像匯出檔處理。 */
    @get
    @usePermission(AuthRead)
    @route("/{id}/export")
    @operationId("exportProject")
    exportProject(@path id: int32): DownloadFileResponse;

    /** 匯入題目匯出檔，一律建立新題目（不覆蓋既有）。
     * 匯出檔內含會在評測主機上執行的指令，因此限管理員。 */
    @post
    @usePermission(AuthAdmin)
    @route("/import")
    @operationId("importProject")
    importProject(
        @header contentType: "multipart/form-data",
        @body payload: ImportProjectRequest,
    ): APICreateResponse<ImportProjectResult>;

    @get
    @usePermission(AuthRead)
    @route("/{id}/grading-project/files")
    @operationId("listGradingProjectFiles")
    listGradingProjectFiles(@path id: int32): APIListResponse<string>;

    @get
    @usePermission(AuthRead)
    @route("/{id}/grading-project/files/{file}")
    @operationId("readGradingProjectFile")
    readGradingProjectFile(
        @path id: int32,
        @path file: string,
    ): APIGetResponse<string>;

    // temporary removed
    // @get
    // @usePermission(AuthAdmin)
    // @route("/{id}/grading-project/files/{file}/download")
    // @operationId("downloadGradingProjectFile")
    // downloadGradingProjectFile(
    //     @path id: int32,
    //     @path file: string,
    // ): DownloadFileResponse;

    // @delete
    // @usePermission(AuthAdmin)
    // @route("/{id}/grading-project/files/{file}")
    // @operationId("deleteGradingProjectFile")
    // deleteGradingProjectFile(
    //     @path id: int32,
    //     @path file: string,
    // ): APINoDataResponseOrAvailableErrors;

    // @put
    // @usePermission(AuthAdmin)
    // @route("/{id}/grading-project/files/{file}")
    // @operationId("writeGradingProjectFile")
    // writeGradingProjectFile(
    //     @path id: int32,
    //     @path file: string,
    //     @body content: WriteGradingProjectFileRequest,
    // ): APINoDataResponseOrAvailableErrors;
}
