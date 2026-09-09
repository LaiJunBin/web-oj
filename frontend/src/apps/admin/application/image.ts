import { Application } from '@/apps/common/application/base';
import type { BundleExportMode } from '@/apps/admin/domain/bundle';
import type { ImageServiceInterface } from '@/apps/admin/application/service/image';
import type {
	CreateImagePayload,
	UpdateImagePayload,
	ImagesStorageInterface,
	ImageStorageInterface
} from '@/apps/admin/domain/image';
import { container } from '@/apps/admin/container';

export class ImageApplication extends Application {
	constructor(
		private readonly imageService: ImageServiceInterface,

		private readonly imagesStorage: ImagesStorageInterface,
		private readonly imageStorage: ImageStorageInterface
	) {
		super();
	}

	fetchImages() {
		return this.imageService
			.fetchImages()
			.then((images) => {
				this.imagesStorage.set(images);
			})
			.catch((error) => {
				this.imagesStorage.setError(error);
			});
	}

	fetchImage(tag: string) {
		return this.imageService
			.fetchImageDetail(tag)
			.then((image) => {
				this.imageStorage.set(image);
			})
			.catch((error) => {
				this.imageStorage.setError(error);
			});
	}

	createImage(payload: CreateImagePayload) {
		return this.imageService.createImage(payload);
	}

	updateImage(tag: string, payload: UpdateImagePayload) {
		return this.imageService.updateImage(tag, payload);
	}

	deleteImage(tag: string) {
		return this.imageService.deleteImage(tag);
	}

	rebuildImage(tag: string) {
		return this.imageService.rebuildImage(tag);
	}

	cancelBuildImage(tag: string) {
		return this.imageService.cancelBuildImage(tag);
	}

	exportImage(tag: string, mode: BundleExportMode) {
		return this.imageService.exportImage(tag, mode);
	}

	importImage(file: File) {
		return this.imageService.importImage(file);
	}
}

export const useImageApp = () => container.get(ImageApplication);
