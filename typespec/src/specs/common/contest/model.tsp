namespace Common.ContestSpec;

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

model ContestDetail {
    ...Contest;
    projects: ContestProject[];
}

model ContestProject {
    @key id: int32;
    name: string;
    status: ProjectStatus;
}

model ContestProjectSchedule {
    name: string;
    description: string;
    startTime: utcDateTime;
    endTime: utcDateTime;
    score: float64;
}

model ContestSchedule {
    projects: ContestProjectSchedule[];
}
