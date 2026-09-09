import { apiClient } from '@/api';
import type { AssetServiceInterface } from '@/apps/admin/application/service/asset';
import type { UploadAssetPayload, UpdateAssetPayload } from '@/apps/admin/domain/asset';

export class AssetService implements AssetServiceInterface {
	downloadAsset(id: string, saveName: string): Promise<void> {
		return apiClient
			.downloadAsset({
				params: {
					id
				},
				responseType: 'blob'
			})
			.then((blob) => {
				const url = window.URL.createObjectURL(blob as unknown as Blob);
				const a = document.createElement('a');
				a.download = saveName;
				a.href = url;
				a.click();
			})
			.catch((error) => {
				console.error(error);
				throw new Error('下載檔案失敗');
			});
	}
	uploadAsset(projectId: number, payload: UploadAssetPayload): Promise<void> {
		return apiClient
			.uploadAsset(payload, {
				params: {
					id: projectId
				}
			})
			.then((response) => {
				const { success } = response;
				if (!success) {
					throw new Error('Failed to create asset');
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
	updateAsset(id: string, payload: UpdateAssetPayload): Promise<void> {
		return apiClient
			.updateAsset(payload, {
				params: {
					id
				}
			})
			.then((response) => {
				const { success } = response;
				if (!success) {
					throw new Error('Failed to update asset');
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
	deleteAsset(id: string): Promise<void> {
		return apiClient
			.deleteAsset(undefined, {
				params: {
					id
				}
			})
			.then((response) => {
				const { success } = response;
				if (!success) {
					throw new Error('Failed to delete asset');
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
