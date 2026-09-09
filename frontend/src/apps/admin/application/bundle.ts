import { Application } from '@/apps/common/application/base';
import type { BundleServiceInterface } from '@/apps/admin/application/service/bundle';
import type { BundleExportMode } from '@/apps/admin/domain/bundle';
import { container } from '@/apps/admin/container';

/**
 * 完整匯出檔的匯出/匯入。這裡刻意沒有 storage:預覽與匯入報告都只屬於當下那一次
 * 上傳,存進共用狀態只會讓下一次開啟頁面時看到上一批的殘留。
 */
export class BundleApplication extends Application {
	constructor(private readonly bundleService: BundleServiceInterface) {
		super();
	}

	exportContestBundle(contestId: number, mode: BundleExportMode) {
		return this.bundleService.exportContestBundle(contestId, mode);
	}

	previewBundle(file: File) {
		return this.bundleService.previewBundle(file);
	}

	importBundle(file: File, force = false) {
		return this.bundleService.importBundle(file, force);
	}
}

export const useBundleApp = () => container.get(BundleApplication);
