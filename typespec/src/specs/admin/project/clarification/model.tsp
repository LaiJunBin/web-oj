namespace Admin.ProjectSpec.ClarificationSpec;

using Shared;

model UserInfo {
    id: int32;
    name: string;
    username: string;
}

model ClarificationInfo {
    @key id: int32;
    subject: string;
    content: string;
    senderUser: UserInfo;
    receiverUser: UserInfo | null;
    isAdmin: boolean;
    sendToAll: boolean;
    isAnswered: boolean;
    createdAt: utcDateTime;
    updatedAt: utcDateTime;
}

model Clarification {
    ...ClarificationInfo;
    replyClarification: Clarification | null;
}
