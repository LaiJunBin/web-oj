namespace Admin.ProjectSpec.LifecycleSpec;

using Shared;

enum LifecycleType {
    BEFORE_ALL: "before_all",
    BEFORE_EACH: "before_each",
    AFTER_EACH: "after_each",
    AFTER_ALL: "after_all",
}

model Lifecycle {
    @key id: int32;
    name: string;
    description: string;
    command: string;
}

model LifecycleDetail {
    ...Lifecycle;
    type: LifecycleType;
    runsOn: RunsOn;
    createdAt: utcDateTime;
    updatedAt: utcDateTime;
}
