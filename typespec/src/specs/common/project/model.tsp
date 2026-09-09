namespace Common.ProjectSpec;

using Shared;

model Contest {
    @key id: int32;
    name: string;
}

enum ProjectStatus {
    NotStarted,
    InProgress,
    Ended
}

model Project {
    @key id: int32;
    name: string;
    description: string;
    status: ProjectStatus;
    gradingItems: GradingItem[];
    assets: AssetInfo[];
    timeLeft: int64;
}

model GradingItem {
    @key id: int32;
    label: string;
    name: string;
    description: string;
    score: float64;
    publicTestcases: Testcase[];
}

model Testcase {
    @key id: int32;
    name: string;
}

model ProjectScoreboardData {
    name: string;
    username: string;
    statuses: ScoreboardDataStatus[];
}

model ProjectScoreboardHead {
    id: int32;
    name: string;
}

model ProjectScoreboard {
    heads: ProjectScoreboardHead[];
    rows: ProjectScoreboardData[];

    /** true 時前端以精簡樣式（單一橫條、逐項漸層）呈現 */
    compact: boolean;
}
