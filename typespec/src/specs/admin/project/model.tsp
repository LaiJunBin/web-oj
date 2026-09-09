namespace Admin.ProjectSpec;

using ImageSpec;
using LifecycleSpec;
using GradingItemSpec;
using AssetSpec;
using Shared;

model Project {
    @key id: int32;
    name: string;
    description: string;
    image?: string;
    testImage?: string;
    createdAt: utcDateTime;
    updatedAt: utcDateTime;
}

model ProjectDetail  {
    @key id: int32;
    name: string;
    description: string;
    image?: ImageInfo;
    testImage?: ImageInfo;
    env: string;
    testEnv: string;
    resultParser: string;
    assets: Asset[];
    dynamicFields: ProjectDynamicField[];
    lifecycle: ImageLifecycle;
    testLifecycle: ImageLifecycle;
    createdAt: utcDateTime;
    updatedAt: utcDateTime;
}

model ImageLifecycle {
    beforeAll: Lifecycle[];
    beforeEach: Lifecycle[];
    afterAll: Lifecycle[];
    afterEach: Lifecycle[];
}
