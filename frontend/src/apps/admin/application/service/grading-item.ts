import type {
	CreateGradingItemPayload,
	GradingItem,
	GradingItemDetail,
	UpdateGradingItemPayload
} from '@/apps/admin/domain/grading-item';

export interface GradingItemServiceInterface {
	fetchGradingItems(projectId: number): Promise<GradingItem[]>;
	fetchGradingItemDetail(id: number): Promise<GradingItemDetail>;
	createGradingItem(projectId: number, payload: CreateGradingItemPayload): Promise<number>;
	updateGradingItem(id: number, payload: UpdateGradingItemPayload): Promise<void>;
	deleteGradingItem(id: number): Promise<void>;
}
