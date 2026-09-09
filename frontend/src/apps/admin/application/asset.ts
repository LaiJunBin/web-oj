import { Application } from '@/apps/common/application/base';
import type { AssetServiceInterface } from '@/apps/admin/application/service/asset';
import type { UploadAssetPayload, UpdateAssetPayload } from '@/apps/admin/domain/asset';
import { container } from '@/apps/admin/container';

export class AssetApplication extends Application {
	constructor(private readonly assetService: AssetServiceInterface) {
		super();
	}

	downloadAsset(id: string, saveName: string) {
		return this.assetService.downloadAsset(id, saveName);
	}

	uploadAsset(projectId: number, payload: UploadAssetPayload) {
		return this.assetService.uploadAsset(projectId, payload);
	}

	updateAsset(id: string, payload: UpdateAssetPayload) {
		return this.assetService.updateAsset(id, payload);
	}

	deleteAsset(id: string) {
		return this.assetService.deleteAsset(id);
	}
}

export const useAssetApp = () => container.get(AssetApplication);
