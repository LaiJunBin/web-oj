namespace Admin.BundleSpec;

/** 一個項目在匯入時會發生什麼事。 */
enum BundleItemStatus {
    /** 系統沒有，會新建 */
    new: "new",

    /** 系統已有且內容相同，直接沿用，不會有任何改動 */
    identical: "identical",

    /** 系統已有但內容不同。未強制覆蓋時這會擋下整個匯入 */
    conflict: "conflict",
}

model BundleImageItem {
    tag: string;
    mode: string;
    status: BundleItemStatus;

    /** 覆蓋這個映像會連帶影響的題目名稱。空陣列代表沒有題目在用它 */
    usedByProjects: string[];
}

model BundleProjectItem {
    name: string;
    status: BundleItemStatus;

    /** 這個題目參照、但匯出檔與系統都沒有的映像 tag */
    missingImageTags: string[];
}

model BundleContestItem {
    name: string;

    /** 競賽的題目 slot 數 */
    projectSlots: int32;

    /** 對不到題庫的 slot（匯出檔內的題目也算數，匯入時會先建立） */
    unresolvedProjects: string[];
}

/** 預覽結果：這個匯出檔會做什麼，以及有什麼擋著。 */
model BundlePreview {
    formatVersion: int32;
    images: BundleImageItem[];
    projects: BundleProjectItem[];

    /** 匯出檔若不含競賽則為 null */
    contest: BundleContestItem | null;

    /**
     * 阻擋匯入的原因。非空時 importBundle 以 409 拒絕;其中「同 tag 但內容
     * 不同」類的衝突可用 force 覆蓋,缺映像則不行(400 保留給格式不合法的
     * 匯出檔,那種請求到不了這份清單)。
     *
     * 分開列出而不是塞進各項目裡，是因為前端要用它決定「執行」按鈕能不能按，
     * 而那個判斷不該是走訪三個陣列再自己推論出來的。
     */
    blockers: string[];
}

/** 匯入結果：實際建立了什麼，以及還要做什麼。 */
model BundleImportReport {
    createdImages: string[];
    reusedImages: string[];

    /** 已排入建置的映像 tag。source 模式需要建置，archive 模式不需要 */
    queuedBuilds: string[];

    createdProjectIds: int32[];
    contestId: int32 | null;

    /** 競賽中對不到題目的 slot，這些沒有建立 */
    unresolvedProjects: string[];

    /**
     * 僅承載匯入成功時的非致命提醒（目前恆為空）。中途失敗不會出現在這裡：
     * 那會以 500 回報，details 列出失敗原因與已建立的項目，由操作者自行
     * 接續完成或移除——系統不做假裝的自動回滾。
     */
    warnings: string[];
}
