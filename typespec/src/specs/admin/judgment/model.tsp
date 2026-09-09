namespace Admin.JudgmentSpec;

using Shared;


model Project {
    id: int32;
    imageTag: string;
    testImageTag: string;
    env: string;
    testEnv: string;
    mountPath: string;
    resultParser: string;
}

model Result {
    exitCode: int32;
    execTime: int64;
    stdout: string;
    stderr: string;
}

model Lifecycle {
    id: int32;
    executionEnvironment: string;
    command: string;
    result: Result;
}

model JudgmentLifeycle {
    beforeAll: Lifecycle[];
    afterAll: Lifecycle[];
}

model TestcaseLifecycle {
    beforeEach: Lifecycle[];
    afterEach: Lifecycle[];
}

model Testcase {
    id: int32;
    name: string;

    @doc("Score actually earned by this testcase, decomposed from the grading item's scoring strategy. NOTE: this field previously carried the configured max score, which now lives in `totalScore`.")
    score: float64;

    @doc("Configured max score for this testcase.")
    totalScore: float64;

    @doc("Raw score awarded by the result parser. Under the `increase` strategy it is added when the testcase passes; under `decrease` it is subtracted from the grading item's max when the testcase does not pass. Prefer `score` for display; this is the underlying parser output.")
    parsedScore: float64;

    status: string;
    command: string;
    lifecycle: TestcaseLifecycle;
    result: Result;
}

model GradingItem {
    id: int32;
    label: string;
    name: string;

    @doc("Score actually earned by this grading item.")
    score: float64;

    @doc("Configured max score for this grading item.")
    totalScore: float64;

    status: string;

    @doc("How testcase scores roll up into this item's score: `increase` sums the parsed score of passing testcases, `decrease` subtracts the parsed score of non-passing ones from the max.")
    scoringStrategy: string;

    resultParser: string;
    testcases: Testcase[];
}

model Judgment {
    id: string;
    judgeType: string;
    project: Project;
    gradingItems: GradingItem[];
    judgmentLifecycle: JudgmentLifeycle;
    status: string;
    score: float64;
    rejudgeId: string | null;
    totalScore: float64;
    judgingStartedAt: utcDateTime | null;
    judgingFinishedAt: utcDateTime | null;
    createdAt: utcDateTime;

    /** 評測以 Error 收場時的原因（例如環境建立重試耗盡）；僅後台 API 暴露，選手端不提供 */
    errorMessage: string | null;

    /** 裁判寫給選手的附註（選手可見，與系統 reason 並存）；null 表示尚未填寫。後台以此回顯供編輯 */
    staffNote: string | null;
}

enum RejudgeStatus {
    Pending;
    Running;
    Finished;
    Approved;
    Rejected;
    Error;
    Canceled;
}

model Rejudge {
    id: string;

    @doc("The contest project this rejudge was raised against. Both entry points, a single submission and a whole contest project, resolve to exactly one contest project, so this is always set.")
    contestProjectId: int32;

    reason: string;
    effectedJudgments: int32;

    @doc("How many of this rejudge's judgments have reached a terminal state. Never exceeds effectedJudgments, which is the total; the two together are the progress bar.")
    completedJudgments: int32;

    status: RejudgeStatus;
    createdAt: utcDateTime;
}

model JudgmentError {
    judgmentId: string;
    submissionId: string;
    contestProjectId: int32;

    @doc("Why the judgment ended in Error. Null for an errored judgment recorded before the reason was captured.")
    errorMessage: string | null;

    createdAt: utcDateTime;
}

model JudgmentOutcomeChange  {
    status: string;
    score: float64;
    totalScore: float64;
}

model SubmissionComparison {
    id: string;

    @doc("The judgment this submission had adopted before the rejudge ran, or null when it had no adopted judgment yet.")
    before: JudgmentOutcomeChange | null;

    after: JudgmentOutcomeChange;
}

model RejudgeDetail {
    id: string;
    reason: string;
    submissions: SubmissionComparison[];
    status: RejudgeStatus;
    createdAt: utcDateTime;
}

model TestcaseOutcome {
    id: int32;
    name: string;
    status: string;

    @doc("Score actually earned by this testcase, decomposed from the grading item's scoring strategy.")
    score: float64;

    @doc("Configured max score for this testcase.")
    totalScore: float64;

    @doc("Raw score awarded by the result parser. Prefer `score` for display; this is the underlying parser output.")
    parsedScore: float64;
}

model GradingItemOutcome {
    id: int32;
    label: string;
    name: string;
    status: string;

    @doc("Score actually earned by this grading item.")
    score: float64;

    @doc("Configured max score for this grading item.")
    totalScore: float64;

    @doc("How testcase scores roll up into this item's score: `increase` sums the parsed score of passing testcases, `decrease` subtracts the parsed score of non-passing ones from the max.")
    scoringStrategy: string;

    testcases: TestcaseOutcome[];
}

model JudgmentOutcomeDetail {
    ...JudgmentOutcomeChange;

    judgmentId: string;

    @doc("Grading items as recorded by this judgment. Pair before/after by `id`; an id present on only one side means the item was added or removed between the two judge runs.")
    gradingItems: GradingItemOutcome[];
}

model SubmissionComparisonDetail {
    id: string;

    @doc("The judgment this submission had adopted before the rejudge ran, or null when it had no adopted judgment yet.")
    before: JudgmentOutcomeDetail | null;

    after: JudgmentOutcomeDetail;
}
