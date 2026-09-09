namespace User.UserSpec;

using Shared;

model SubmissionTestcase {
    id: int32;
    execTime: int64;
    exitCode: int32;
    stdout: string;
    stderr: string;
    status: string;
}

model SubmissionGradingItem {
    id: int32;
    label: string;
    score: float64;
    status: SubmissionGradingItemStatus;
    testcases: SubmissionTestcase[];
}

enum SubmissionStatus {
    Queued,
    Judging,
    Judged,
    Cancelled,
    Error,
}

model Submission {
    id: string;
    status: SubmissionStatus;
    isPartial: boolean;
    score: float64;
    totalScore: float64;
    createdAt: utcDateTime;

    @doc("Why the judgment ended the way it did, when the system recorded one: an Error's contestant-facing explanation, or the reason given for a cancellation. Absent when there is nothing to explain. Never carries commands, paths or internal details — those stay on the admin API.")
    reason?: string;

    @doc("A note a judge wrote to the contestant after looking into this judgment. Shown alongside — never instead of — `reason`, on the list as well as the detail. Absent when no note was written. Never carries commands, paths or internal details.")
    staffNote?: string;
}

model SubmissionDetail {
    ...Submission;
    gradingItems: SubmissionGradingItem[];
}

model Clarification {
    @key id: int32;
    subject: string;
    content: string;
    isAdminBroadcast: boolean;
    isAdminResponse: boolean;
    createdAt: utcDateTime;
}

model ClarificationDetail {
    ...Clarification;
    replyClarification: ClarificationDetail | null;
}
