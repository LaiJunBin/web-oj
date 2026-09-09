import type {
	Image,
	ImageDetail,
	ImagesStorageInterface,
	ImageStorageInterface
} from '@/apps/admin/domain/image';
import { BrowserStorage } from '@features/common/store/base.svelte';

export class ImagesStorage extends BrowserStorage<Image[]> implements ImagesStorageInterface {}

export class ImageStorage extends BrowserStorage<ImageDetail> implements ImageStorageInterface {}
