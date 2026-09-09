import { apiClient } from '@/api';
import type { TestcaseServiceInterface } from '@/apps/admin/application/service/testcase';
import type {
	CreateTestcasePayload,
	ProjectTestcaseGradingItem,
	Testcase,
	UpdateTestcasePayload
} from '@/apps/admin/domain/testcase';

export class TestcaseService implements TestcaseServiceInterface {
	fetchProjectTestcases(projectId: number): Promise<ProjectTestcaseGradingItem[]> {
		return apiClient
			.listProjectTestcases({
				params: {
					id: projectId
				}
			})
			.then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('Failed to fetch testcases');
				}

				return data;
			});
	}
	fetchTestcase(id: number): Promise<Testcase> {
		return apiClient
			.getTestcase({
				params: {
					id
				}
			})
			.then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('Failed to fetch testcase');
				}

				return data;
			});
	}
	createTestcase(gradingItemId: number, payload: CreateTestcasePayload): Promise<number> {
		return apiClient
			.createTestcase(payload, {
				params: {
					id: gradingItemId
				}
			})
			.then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('Failed to create testcase');
				}
				return data;
			})
			.catch((error) => {
				if ('response' in error) {
					const { data } = error.response;
					throw new Error(data.message);
				}

				throw error;
			});
	}
	updateTestcase(id: number, payload: UpdateTestcasePayload): Promise<void> {
		return apiClient
			.updateTestcase(payload, {
				params: {
					id
				}
			})
			.then((response) => {
				const { success } = response;
				if (!success) {
					throw new Error('Failed to update testcase');
				}
			})
			.catch((error) => {
				if ('response' in error) {
					const { data } = error.response;
					throw new Error(data.message);
				}

				throw error;
			});
	}
	deleteTestcase(id: number): Promise<void> {
		return apiClient
			.deleteTestcase(undefined, {
				params: {
					id
				}
			})
			.then((response) => {
				const { success } = response;
				if (!success) {
					throw new Error('Failed to delete testcase');
				}
			})
			.catch((error) => {
				if ('response' in error) {
					const { data } = error.response;
					throw new Error(data.message);
				}

				throw error;
			});
	}
}
