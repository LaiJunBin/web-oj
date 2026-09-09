import type { AssetsApiSchemas } from '@/api/gen/Assets';
import type { z } from 'zod';

export type Asset = z.infer<typeof AssetsApiSchemas.Shared_Asset>;

export type UploadAssetPayload = z.infer<
	typeof AssetsApiSchemas.Admin_ProjectSpec_AssetSpec_UploadAssetRequestMultiPart
>;

export type UpdateAssetPayload = z.infer<
	typeof AssetsApiSchemas.Admin_ProjectSpec_AssetSpec_UpdateAssetRequestMultiPart
>;
