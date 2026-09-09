import { apiClient } from '@/api';
import { BundlesApi } from '@/api/gen/Bundles';
import type { BundleServiceInterface } from '@/apps/admin/application/service/bundle';
import {
	BundleConflictError,
	type BundleExportMode,
	type BundleImportReport,
	type BundlePreview
} from '@/apps/admin/domain/bundle';
import { downloadBlob, endpointPathByAlias } from './download';

export class BundleService implements BundleServiceInterface {
	exportContestBundle(contestId: number, mode: BundleExportMode): Promise<void> {
		const path = endpointPathByAlias(BundlesApi, 'exportBundle').replace(':id', String(contestId));
		return downloadBlob(path, `contest-${contestId}-bundle.zip`, {
			params: { mode },
			// archive 模式下最常見的失敗是「有映像還沒建置」,後端會把 tag 列在
			// details 裡,downloadBlob 會一併帶出來,所以這裡只需要保底的一句。
			failureMessage: '匯出完整競賽失敗'
		});
	}

	previewBundle(file: File): Promise<BundlePreview> {
		return apiClient
			.previewBundle({ bundle: file })
			.then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('讀取匯出檔失敗');
				}
				return data;
			})
			.catch((error) => {
				throw new Error(messageOf(error, '讀取匯出檔失敗'));
			});
	}

	importBundle(file: File, force = false): Promise<BundleImportReport> {
		return apiClient
			.importBundle({ bundle: file, force })
			.then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('匯入失敗');
				}
				return data;
			})
			.catch((error) => {
				// 409 是「會覆蓋既有東西」,不是錯誤:原樣把後端寫好的影響清單往上傳,
				// 讓確認畫面逐條顯示後再決定要不要 force。
				const response = (error as { response?: { status?: number; data?: ConflictBody } })
					?.response;
				if (response?.status === 409) {
					throw new BundleConflictError(
						response.data?.message ?? '匯入會覆蓋既有資料',
						response.data?.details ?? []
					);
				}
				throw new Error(messageOf(error, '匯入失敗'));
			});
	}
}

type ConflictBody = { message?: string; details?: string[] };

const messageOf = (error: unknown, fallback: string) => {
	if (error instanceof Error && !('response' in error)) return error.message || fallback;
	const data = (error as { response?: { data?: ConflictBody } })?.response?.data;
	if (!data?.message) return fallback;
	return data.details?.length ? `${data.message}\n${data.details.join('\n')}` : data.message;
};
