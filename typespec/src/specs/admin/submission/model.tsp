namespace Admin.SubmissionSpec;

using Shared;
using JudgmentSpec;

model User {
    id: int32;
    name: string;
    username: string;
}

model Contest {
    id: int32;
    name: string;
}

model ProjectInfo {
    id: int32;
    name: string;
}

model Submission {
    @key id: string;
    user: User;
    contest: Contest;
    project: ProjectInfo;
    isPartial: boolean;
    ignored: boolean;
    status: SubmissionStatus;
    type: SubmissionType;
    score: float64;
    totalScore: float64;
    createdAt: utcDateTime;
}


@doc("Submissions received during one minute-long bucket.")
model SubmissionRatePoint {
    @doc("Start of the bucket, truncated to the minute.")
    minute: utcDateTime;

    count: int32;
}

model TestcaseInfo {
    id: int32;
    name: string;
    score: float64;
}

model GradingItemInfo {
    id: int32;
    label: string;
    name: string;
    score: float64;
    testcases: TestcaseInfo[];
}


model SubmissionDetail {
    id:           string;
	user:         User;
	contest:      Contest;
	project:      ProjectInfo;
	gradingItems: GradingItemInfo[];
	isPartial:    boolean;
    ignored:      boolean;
    judgments: Judgment[];
    createdAt: utcDateTime;
}

