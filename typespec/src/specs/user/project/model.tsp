namespace User.ProjectSpec;

using Shared;

model Project {
    @key id: int32;
    name: string;
    description: string;
}

model ProjectDetail {
    ...Project;
    gradingItems: GradingItem[];
    dynamicFields: ProjectDynamicField[];
    assets: AssetInfo[];
    timeLeft: int64;

    // Whether the current user is a participant of the contest that owns this
    // project (added directly or via a non-deleted group). Submitting requires
    // this to be true even on public contests, so the frontend gates the submit
    // button/modal on it instead of the global "contestant" role.
    isParticipant: boolean;
}

model GradingItem {
    @key id: int32;
    label: string;
    name: string;
    description: string;
    score: float64;
    outputFormat: GradingItemOutputFormat;
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
