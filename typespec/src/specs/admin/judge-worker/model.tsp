namespace Admin.JudgeWorkerSpec;

using Shared;

enum JudgeWorkerStatus {
    INIT,
    IDLE,
    STARTING,
    RUNNING,
    STOPPED,
    ERROR,
}

enum JudgeOrder {
    JUDGE;
    JUDGE_PARTIAL;
    REJUDGE;
}

model JudgeWorker {
    @key id: string;
    status: JudgeWorkerStatus;
    judgeOrder: JudgeOrder[];
    createdAt: utcDateTime;
    lastExecutionAt: utcDateTime | null;
}

@doc("Judgments waiting per judge type. Every type the pipeline serves is present, including the ones at zero.")
model JudgeQueueDepths {
    judge: int32;
    judgePartial: int32;
    rejudge: int32;
}

model JudgeQueueStatus {
    queueDepths: JudgeQueueDepths;

    @doc("Judgments executing right now. Same number as the judge_active_judgments metric, read from the same counter.")
    activeJudgments: int32;

    @doc("How long the longest-waiting judgment has been waiting, in seconds, or null when nothing is queued. Reported instead of an average because an average absorbs a stall — one judgment stuck for ten minutes barely moves it, but shows up here immediately.")
    oldestQueuedSeconds: float64 | null;
}

@doc("生效中的系統預設值。競賽專案把對應欄位留空時套用的就是這些數字，管理介面用它把「未設定」呈現成具體的值。這裡是手寫的白名單，只放有逐資源覆寫的設定；不得改成自動展開伺服器設定，那會把金鑰與密碼一起帶出去。")
model SystemDefaults {
    @doc("單場評測總時長上限（分鐘）。競賽專案的 maxJudgeDurationMinutes 留空時套用。")
    maxJudgeDurationMinutes: int32;

    @doc("同一位選手兩次提交之間的最短間隔（秒），0 表示不限制。競賽專案的 submissionCooldownSeconds 留空時套用。")
    submissionCooldownSeconds: int32;

    @doc("評測容器的行程數上限（PID），0 表示不限制。競賽專案的 pidsLimit 留空時套用。")
    pidsLimit: int32;

    @doc("評測容器的記憶體上限（MB），0 表示不限制。競賽專案的 memoryLimitMB 留空時套用。")
    memoryLimitMB: int32;
}

