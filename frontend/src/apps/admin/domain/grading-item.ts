import { GradingItemsApiSchemas } from '@/api/gen/GradingItems';
import type { Storage } from '@/apps/common/domain/storage';
import type { z } from 'zod';
import { container } from '@/apps/admin/container';

export type GradingItem = z.infer<
	typeof GradingItemsApiSchemas.Admin_ProjectSpec_GradingItemSpec_GradingItem
>;
export type GradingItemDetail = z.infer<
	typeof GradingItemsApiSchemas.Admin_ProjectSpec_GradingItemSpec_GradingItemDetail
>;

export type CreateGradingItemPayload = z.infer<
	typeof GradingItemsApiSchemas.Admin_ProjectSpec_GradingItemSpec_CreateGradingItemRequest
>;

export const ScoringStrategyEnum = GradingItemsApiSchemas.Shared_GradingItemScoringStrategy.Enum;
export type ScoringStrategy = z.infer<
	typeof GradingItemsApiSchemas.Shared_GradingItemScoringStrategy
>;

export const OutputFormatEnum = GradingItemsApiSchemas.Shared_GradingItemOutputFormat.Enum;
export type OutputFormat = z.infer<typeof GradingItemsApiSchemas.Shared_GradingItemOutputFormat>;

export type UpdateGradingItemPayload = z.infer<
	typeof GradingItemsApiSchemas.Admin_ProjectSpec_GradingItemSpec_UpdateGradingItemRequest
>;

export interface GradingItemsStorageInterface extends Storage<GradingItem[]> {}
export const useGradingItemsStorage = () =>
	container.get('GradingItemsStorage') as GradingItemsStorageInterface;

export interface GradingItemStorageInterface extends Storage<GradingItemDetail> {}
export const useGradingItemStorage = () =>
	container.get('GradingItemStorage') as GradingItemStorageInterface;
