import { Application } from '@/apps/common/application/base';
import type { TestcaseServiceInterface } from '@/apps/admin/application/service/testcase';
import type {
	CreateTestcasePayload,
	ProjectTestcaseGradingItemsStorageInterface,
	TestcaseStorageInterface,
	UpdateTestcasePayload
} from '@/apps/admin/domain/testcase';
import { container } from '@/apps/admin/container';

export class TestcaseApplication extends Application {
	constructor(
		private readonly testcaseService: TestcaseServiceInterface,

		private readonly projectTestcaseGradingItemsStorage: ProjectTestcaseGradingItemsStorageInterface,
		private readonly testcaseStorage: TestcaseStorageInterface
	) {
		super();
	}

	fetchProjectTestcases(projectId: number) {
		return this.testcaseService
			.fetchProjectTestcases(projectId)
			.then((testcases) => {
				this.projectTestcaseGradingItemsStorage.set(testcases);
			})
			.catch((error) => {
				this.projectTestcaseGradingItemsStorage.setError(error);
			});
	}

	fetchTestcase(id: number) {
		return this.testcaseService
			.fetchTestcase(id)
			.then((testcase) => {
				this.testcaseStorage.set(testcase);
			})
			.catch((error) => {
				this.testcaseStorage.setError(error);
			});
	}

	createTestcase(gradingItemId: number, payload: CreateTestcasePayload) {
		return this.testcaseService.createTestcase(gradingItemId, payload);
	}

	updateTestcase(id: number, payload: UpdateTestcasePayload) {
		return this.testcaseService.updateTestcase(id, payload);
	}

	deleteTestcase(id: number) {
		return this.testcaseService.deleteTestcase(id);
	}
}

export const useTestcaseApp = () => container.get(TestcaseApplication);
