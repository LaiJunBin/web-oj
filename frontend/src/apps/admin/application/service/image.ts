import type { BundleExportMode } from '@/apps/admin/domain/bundle';
import type {
	CreateImagePayload,
	Image,
	ImageDetail,
	UpdateImagePayload
} from '@/apps/admin/domain/image';

export interface ImageServiceInterface {
	fetchImages(): Promise<Image[]>;
	fetchImageDetail(tag: string): Promise<ImageDetail>;
	createImage(payload: CreateImagePayload): Promise<string>;
	updateImage(tag: string, payload: UpdateImagePayload): Promise<void>;
	deleteImage(tag: string): Promise<void>;

	rebuildImage(tag: string): Promise<void>;
	cancelBuildImage(tag: string): Promise<void>;
	exportImage(tag: string, mode: BundleExportMode): Promise<void>;
	importImage(file: File): Promise<{ tag: string; mode: string; buildQueued: boolean }>;
}
