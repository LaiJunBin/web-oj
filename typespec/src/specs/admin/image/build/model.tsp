namespace Admin.BuildSpec;

enum BuildStatus {
    PENDING,
    BUILDING,
    ERROR,
    SUCCESS,
    CANCELLED,
}

model BuildInfo {
    batch: int32;
    status: BuildStatus;
}

model Log {
    level: string;
    message: string;
    time: utcDateTime;
}

model Build {
    ...BuildInfo;
    logs: Log[];
}
