import type { TestcasesApiSchemas } from '@/api/gen/Testcases';
import type { Storage } from '@/apps/common/domain/storage';
import type { z } from 'zod';
import { container } from '@/apps/admin/container';

export type Testcase = z.infer<
	typeof TestcasesApiSchemas.Admin_ProjectSpec_GradingItemSpec_TestcaseSpec_Testcase
>;

export type ProjectTestcaseGradingItem = z.infer<
	typeof TestcasesApiSchemas.Admin_ProjectSpec_GradingItemSpec_TestcaseSpec_ProjectTestcaseGradingItem
>;

export type CreateTestcasePayload = z.infer<
	typeof TestcasesApiSchemas.Admin_ProjectSpec_GradingItemSpec_TestcaseSpec_CreateTestcaseRequest
>;

export type UpdateTestcasePayload = z.infer<
	typeof TestcasesApiSchemas.Admin_ProjectSpec_GradingItemSpec_TestcaseSpec_UpdateTestcaseRequest
>;

export interface TestcasesStorageInterface extends Storage<Testcase[]> {}
export const useTestcasesStorage = () =>
	container.get('TestcasesStorage') as TestcasesStorageInterface;

export interface TestcaseStorageInterface extends Storage<Testcase> {}
export const useTestcaseStorage = () =>
	container.get('TestcaseStorage') as TestcaseStorageInterface;

export interface ProjectTestcaseGradingItemsStorageInterface
	extends Storage<ProjectTestcaseGradingItem[]> {}
export const useProjectTestcaseGradingItemsStorage = () =>
	container.get(
		'ProjectTestcaseGradingItemsStorage'
	) as ProjectTestcaseGradingItemsStorageInterface;
