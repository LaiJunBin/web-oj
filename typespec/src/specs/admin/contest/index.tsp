import "@typespec/http";
import "@typespec/openapi";
import "./model.tsp";

using TypeSpec.OpenAPI;
using TypeSpec.Http;
using Shared;

@route("/contests")
@tag("Contests")
namespace Admin.ContestSpec;

model ImportContestRequest {
    /** 競賽匯出檔（.zip），內含 contest.json */
    bundle: bytes;
}

model ImportContestResult {
    contestId: int32;

    /** 實際建立的競賽名稱。同名時會加上匯入時間後綴 */
    name: string;

    /** 已建立的競賽專案數 */
    createdProjects: int32;

    /**
     * 匯出檔指名、但題庫裡找不到的題目名稱。
     *
     * 競賽匯出檔只以名稱參照題目，不內嵌題目定義——這樣匯入競賽不會連帶建立題目，
     * 也就沒有先後順序問題。找不到的 slot 不會建立，列在這裡讓操作者知道
     * 還要先匯入哪些題目匯出檔。
     */
    unresolvedProjects: string[];
}

model CreateRequestProject {
    projectId: int32;
    name: string;
    activateTime: utcDateTime;
    deactivateTime: utcDateTime;
    startTime: utcDateTime;
    endTime: utcDateTime;
    isPublic: boolean;
    judgePriorityStrategy: string;

    /** 單場評測總時長上限（分鐘）；未設定時使用系統全域預設 */
    maxJudgeDurationMinutes?: int32;

    /** 同一位選手兩次提交之間的最短間隔（秒）；未設定時使用系統全域預設，0 表示不限制 */
    submissionCooldownSeconds?: int32;

    /** 評測容器的行程數上限（PID）；未設定時使用系統全域預設，0 表示不限制 */
    pidsLimit?: int32;

    /** 評測容器的記憶體上限（MB）；未設定時使用系統全域預設，0 表示不限制 */
    memoryLimitMB?: int32;

    /** 前台計分板使用精簡樣式（單一橫條、逐項漸層）取代逐項表格 */
    scoreboardCompact?: boolean;

    /** 前台計分板將選手姓名隱碼（僅遮罩 name，不影響 username 與後台顯示） */
    scoreboardMaskNames?: boolean;
}

model CreateContestRequest {
    name: string;
    description: string;
    activateTime: utcDateTime;
    deactivateTime: utcDateTime;
    projects: CreateRequestProject[];
    isPublic: boolean;
}

model UpdateRequestProject {
    id?: int32;
    projectId?: int32;
    name?: string;
    activateTime?: utcDateTime;
    deactivateTime?: utcDateTime;
    startTime?: utcDateTime;
    endTime?: utcDateTime;
    isPublic?: boolean;
    judgePriorityStrategy?: string;
    maxJudgeDurationMinutes?: int32;
    submissionCooldownSeconds?: int32;
    pidsLimit?: int32;
    memoryLimitMB?: int32;
    scoreboardCompact?: boolean;
    scoreboardMaskNames?: boolean;
}

model UpdateContestRequest {
    name?: string;
    description?: string;
    activateTime?: utcDateTime;
    deactivateTime?: utcDateTime;
    projects?: UpdateRequestProject[];
    isPublic?: boolean;
}

interface Interface
    extends CqrsResourcesInterface<
            "Contest",
            Contest,
            ContestDetail,
            CreateContestRequest,
            UpdateContestRequest
        > {

    /** 匯出競賽（.zip）：競賽本身的欄位，以及每個競賽專案的時間、資源上限與
     * 計分板設定。題目定義不含在內，只以名稱參照。 */
    @get
    @usePermission(AuthRead)
    @route("/{id}/export")
    @operationId("exportContest")
    exportContest(@path id: int32): DownloadFileResponse;

    /** 匯入競賽匯出檔。一律建立新競賽，不覆蓋既有。
     * 報名者不含在匯出檔內（個資，且換場地必然換人），匯入後需自行加入。 */
    @post
    @usePermission(AuthAdmin)
    @route("/import")
    @operationId("importContest")
    importContest(
        @header contentType: "multipart/form-data",
        @body payload: ImportContestRequest,
    ): APICreateResponse<ImportContestResult>;
}


