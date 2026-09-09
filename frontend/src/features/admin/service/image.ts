import { apiClient } from '@/api';
import { ImagesApi } from '@/api/gen/Images';
import type { BundleExportMode } from '@/apps/admin/domain/bundle';
import { downloadBlob, endpointPathByAlias, importFailureMessage } from './download';
import type { ImageServiceInterface } from '@/apps/admin/application/service/image';
import type {
	CreateImagePayload,
	Image,
	ImageDetail,
	UpdateImagePayload
} from '@/apps/admin/domain/image';

export class ImageService implements ImageServiceInterface {
	fetchImages(): Promise<Image[]> {
		return apiClient.listImage().then((response) => {
			const { success, data } = response;
			if (!success) {
				throw new Error('Failed to fetch images');
			}

			return data;
		});
	}
	fetchImageDetail(tag: string): Promise<ImageDetail> {
		return Promise.all([
			apiClient
				.getImage({
					params: {
						tag
					}
				})
				.then((response) => {
					const { success, data } = response;
					if (!success) {
						throw new Error('Failed to fetch image');
					}

					return data;
				}),
			apiClient
				.listBuild({
					params: {
						tag
					}
				})
				.then((response) => {
					const { success, data } = response;
					if (!success) {
						throw new Error('Failed to fetch image builds');
					}

					return data;
				})
		]).then(([image, builds]) => {
			return {
				...image,
				builds
			};
		});
	}

	createImage(payload: CreateImagePayload): Promise<string> {
		return apiClient
			.createImage(payload)
			.then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('Failed to create image');
				}

				return data;
			})
			.catch((error) => {
				if ('response' in error) {
					const { data } = error.response;
					throw new Error(data.message);
				}

				throw error;
			});
	}

	updateImage(tag: string, payload: UpdateImagePayload): Promise<void> {
		return apiClient
			.updateImage(payload, {
				params: {
					tag
				}
			})
			.then((response) => {
				const { success } = response;
				if (!success) {
					throw new Error('Failed to update image');
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
	deleteImage(tag: string): Promise<void> {
		return apiClient
			.deleteImage(undefined, {
				params: {
					tag
				}
			})
			.then((response) => {
				const { success } = response;
				if (!success) {
					throw new Error('Failed to delete image');
				}
			});
	}

	rebuildImage(tag: string): Promise<void> {
		return apiClient
			.rebuild(undefined, {
				params: {
					tag
				}
			})
			.then((response) => {
				const { success } = response;
				if (!success) {
					throw new Error('Failed to rebuild image');
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

	cancelBuildImage(tag: string): Promise<void> {
		return apiClient
			.cancelBuild(undefined, {
				params: {
					tag
				}
			})
			.then((response) => {
				const { success } = response;
				if (!success) {
					throw new Error('Failed to cancel build image');
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

	exportImage(tag: string, mode: BundleExportMode): Promise<void> {
		const path = endpointPathByAlias(ImagesApi, 'exportImage').replace(':tag', encodeURIComponent(tag));
		return downloadBlob(path, `image-${tag.replace(/[^\w.-]/g, '_')}.zip`, {
			params: { mode },
			failureMessage: '匯出映像失敗'
		});
	}

	importImage(file: File) {
		return apiClient
			.importImage({ bundle: file })
			.then((response) => {
				if (!response.success) throw new Error('匯入映像失敗');
				return response.data;
			})
			.catch((error) => {
				throw new Error(importFailureMessage(error, '匯入映像失敗'));
			});
	}
}
