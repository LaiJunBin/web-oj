import type { BundlesApiSchemas } from '@/api/gen/Bundles';
import type { z } from 'zod';

/** 匯出時映像以哪種形態隨行:source 要在匯入端建置,archive 可離線使用。 */
export type BundleExportMode = 'source' | 'archive';

export type BundlePreview = z.infer<typeof BundlesApiSchemas.Admin_BundleSpec_BundlePreview>;
export type BundleImageItem = z.infer<typeof BundlesApiSchemas.Admin_BundleSpec_BundleImageItem>;
export type BundleProjectItem = z.infer<typeof BundlesApiSchemas.Admin_BundleSpec_BundleProjectItem>;
export type BundleImportReport = z.infer<
	typeof BundlesApiSchemas.Admin_BundleSpec_BundleImportReport
>;

/**
 * 匯入被既有資料擋下時,後端回 409 並在 details 裡逐條說明影響範圍。帶著這個
 * 型別把它傳到 UI,讓確認畫面能原文顯示,而不是壓成一句「匯入失敗」。
 */
export class BundleConflictError extends Error {
	constructor(
		message: string,
		readonly details: string[]
	) {
		super(message);
		this.name = 'BundleConflictError';
	}
}
