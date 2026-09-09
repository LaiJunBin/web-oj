namespace Admin.ContestProjectSpec;

model ContestProject {
    @key id: int32;
    projectId: int32;
    name: string;
    description: string;
    image?: string;
    testImage?: string;
    startTime: utcDateTime;
    endTime: utcDateTime;
    activateTime: utcDateTime;
    deactivateTime: utcDateTime;
    isPublic: boolean;
    judgePriorityStrategy: string;
    maxJudgeDurationMinutes: int32 | null;
    submissionCooldownSeconds: int32 | null;
    pidsLimit: int32 | null;
    memoryLimitMB: int32 | null;
    scoreboardCompact: boolean;
    scoreboardMaskNames: boolean;
}

model User {
    id: int32;
    name: string;
    username: string;
}

model UserScore {
    user: User;
    score: float64;
    submissionId: string;
}

model JudgmentStatusCount {
    status: string;
    count: int32;
}

model JudgmentAnalytics {
    totalSubmissions: int32;
    totalJudgments: int32;
    statusCounts: JudgmentStatusCount[];
    avgJudgingDurationMs: float64 | null;
    avgQueueWaitMs: float64 | null;
}
