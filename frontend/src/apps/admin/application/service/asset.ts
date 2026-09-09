import type { UploadAssetPayload, UpdateAssetPayload } from '@/apps/admin/domain/asset';

export interface AssetServiceInterface {
	downloadAsset(id: string, saveName: string): Promise<void>;
	uploadAsset(projectId: number, payload: UploadAssetPayload): Promise<void>;
	updateAsset(id: string, payload: UpdateAssetPayload): Promise<void>;
	deleteAsset(id: string): Promise<void>;
}
