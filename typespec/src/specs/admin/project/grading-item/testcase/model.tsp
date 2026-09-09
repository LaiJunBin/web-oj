namespace Admin.ProjectSpec.GradingItemSpec.TestcaseSpec;

model TestcaseInfo {
    @key id: int32;
    name: string;
    score: float64;
    isSample: boolean;
    command: string;
    createdAt: utcDateTime;
    updatedAt: utcDateTime;
}

model TestcaseGradingItemInfo {
    id: int32;
    label: string;
    name: string;
}

model DependTestcaseInfo {
    @key id: int32;
    name: string;
    gradingItem: TestcaseGradingItemInfo;
}

model Testcase {
    ...TestcaseInfo;
    depends: DependTestcaseInfo[];
}

model ProjectTestcase {
    id: int32;
    name: string;

}
model ProjectTestcaseGradingItem {
    id: int32;
    label: string;
    name: string;
    testcases:ProjectTestcase[];
}
