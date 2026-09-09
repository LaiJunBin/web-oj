import type {
	BundleExportMode,
	BundleImportReport,
	BundlePreview
} from '@/apps/admin/domain/bundle';

export interface BundleServiceInterface {
	/** 下載整場競賽的完整匯出檔(競賽 + 題目 + 映像)。 */
	exportContestBundle(contestId: number, mode: BundleExportMode): Promise<void>;
	/** 看這個匯出檔會做什麼,不寫入任何東西。 */
	previewBundle(file: File): Promise<BundlePreview>;
	/** 實際匯入;force 用於覆蓋既有資料(先讓使用者在 409 的說明上確認過)。 */
	importBundle(file: File, force?: boolean): Promise<BundleImportReport>;
}
