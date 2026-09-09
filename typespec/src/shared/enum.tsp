namespace Shared;

enum SortDirection {
    ASC,
    DESC,
}

enum GradingItemOutputFormat {
    TEXT: "text",
    HTML: "html",
}

enum GradingItemScoringStrategy {
    INCREASE: "increase",
    DECREASE: "decrease",
}

enum RunsOn {
    PROJECT: "project",
    TEST: "test",
}

enum SubmissionStatus {
    Queued,
    Judging,
    Judged,
    Rejudged,
    Cancelled,
    Error,
}

enum SubmissionType {
    JUDGE,
    JUDGE_PARTIAL,
    REJUDGE,
}

enum SubmissionGradingItemStatus {
    Pending,
    Running,
    Passed,
    Failed,
    Error,
    Timeout,
    Cancelled,
    NotSubmitted,
}

enum ClarificationTarget {
    TEAM,
    ADMIN,
}

alias ScoreboardDataStatus = SubmissionGradingItemStatus;
