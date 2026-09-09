import { Judge_workersApiSchemas } from '@/api/gen/Judge_workers';
import type { Storage } from '@/apps/common/domain/storage';
import { z } from 'zod';
import { container } from '@/apps/admin/container';

export type JudgeWorker = z.infer<typeof Judge_workersApiSchemas.Admin_JudgeWorkerSpec_JudgeWorker>;
export type JudgeWorkerOrder = z.infer<
	typeof Judge_workersApiSchemas.Admin_JudgeWorkerSpec_JudgeOrder
>;
export const JudgeWorkerOrders = Judge_workersApiSchemas.Admin_JudgeWorkerSpec_JudgeOrder.Values;
export const JudgeWorkerLogAddedEventDataSchema = z
	.object({
		log: z.string(),
		timestamp: z.string().datetime({ offset: true })
	})
	.passthrough();

export type JudgeWorkerLog = z.infer<typeof JudgeWorkerLogAddedEventDataSchema>;

export const JudgeWorkersEventDataSchema = z
	.object({
		data: z.array(Judge_workersApiSchemas.Admin_JudgeWorkerSpec_JudgeWorker),
		success: z.boolean().default(true)
	})
	.passthrough();

export type JudgeWorkersEventData = z.infer<typeof JudgeWorkersEventDataSchema>;

export type CreateJudgeWorkerPayload = z.infer<
	typeof Judge_workersApiSchemas.Admin_JudgeWorkerSpec_CreateJudgeWorkerRequest
>;

export interface JudgeWorkersStorageInterface extends Storage<JudgeWorker[]> {}
export const useJudgeWorkersStorage = () =>
	container.get('JudgeWorkersStorage') as JudgeWorkersStorageInterface;

export interface JudgePriorityStrategiesStorageInterface extends Storage<string[]> {}
export const useJudgePriorityStrategiesStorage = () =>
	container.get('JudgePriorityStrategiesStorage') as JudgePriorityStrategiesStorageInterface;

/**
 * 伺服器端設定的生效預設值。競賽專案的對應欄位留空時就是套用這些值,所以表單要
 * 拿它把「預設」講成實際數字——尤其是更新時無法把欄位清回 null,管理員得知道
 * 要填哪個數字才能回到跟隨預設。
 */
export type SystemDefaults = z.infer<
	typeof Judge_workersApiSchemas.Admin_JudgeWorkerSpec_SystemDefaults
>;

export interface SystemDefaultsStorageInterface extends Storage<SystemDefaults> {}
export const useSystemDefaultsStorage = () =>
	container.get('SystemDefaultsStorage') as SystemDefaultsStorageInterface;
