import type {
	CreateTestcasePayload,
	ProjectTestcaseGradingItem,
	Testcase,
	UpdateTestcasePayload
} from '@/apps/admin/domain/testcase';

export interface TestcaseServiceInterface {
	fetchProjectTestcases(projectId: number): Promise<ProjectTestcaseGradingItem[]>;
	fetchTestcase(id: number): Promise<Testcase>;
	createTestcase(gradingItemId: number, payload: CreateTestcasePayload): Promise<number>;
	updateTestcase(id: number, payload: UpdateTestcasePayload): Promise<void>;
	deleteTestcase(id: number): Promise<void>;
}
