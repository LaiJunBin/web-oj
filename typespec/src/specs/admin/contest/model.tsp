namespace Admin.ContestSpec;

using ProjectSpec;
using ContestProjectSpec;

model Contest {
    @key id: int32;
    name: string;
    description: string;
    totalProjects: int32;
    activateTime: utcDateTime;
    deactivateTime: utcDateTime;
    isPublic: boolean;
    createdAt: utcDateTime;
    updatedAt: utcDateTime;
}

model ContestDetail  {
    @key id: int32;
    name: string;
    description: string;
    isPublic: boolean;
    projects: ContestProject[];
    activateTime: utcDateTime;
    deactivateTime: utcDateTime;
    createdAt: utcDateTime;
    updatedAt: utcDateTime;
}

model ContestParticipants {
    userIds: int32[];
    groupIds: int32[];
}

model ModifyContestUsersRequest {
    userIds: int32[];
}

model ModifyContestGroupsRequest {
    groupIds: int32[];
}
