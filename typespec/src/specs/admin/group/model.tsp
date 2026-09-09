namespace Admin.GroupSpec;

model Group {
    @key id: int32;
    name: string;
    description: string;
}

model GroupDetail {
    @key id: int32;
    name: string;
    description: string;
    userIds: int32[];
}

model CreateGroupRequest {
    name: string;
    description?: string;
}

model UpdateGroupRequest {
    name?: string;
    description?: string;
}

model ModifyGroupUsersRequest {
    userIds: int32[];
}
