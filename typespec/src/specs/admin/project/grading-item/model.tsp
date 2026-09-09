namespace Admin.ProjectSpec.GradingItemSpec;

using TestcaseSpec;
using Shared;

model GradingItemInfo {
    @key id: int32;
    label: string;
    name: string;
    description: string;
    score: float64;
    order: int32;
    createdAt: utcDateTime;
    updatedAt: utcDateTime;
}

model GradingItem {
    ...GradingItemInfo;
    testcaseLength: int32;
}

model GradingItemDetail {
    ...GradingItemInfo;
    scoringStrategy: string;
    depends: DependTestcaseInfo[];
    outputFormat: string;
    resultParser: string;
    testcases: TestcaseInfo[];
}
