namespace Admin.ImageSpec;

model ImageInfo {
    @key tag: string;
    description: string;
    isReady: boolean;
    createdAt: utcDateTime;
    updatedAt: utcDateTime;
}

model Image {
    ...ImageInfo;
    mountPath: string;
}
