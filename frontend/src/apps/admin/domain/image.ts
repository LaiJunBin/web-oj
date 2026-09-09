import type { ImagesApiSchemas } from '@/api/gen/Images';
import type { Storage } from '@/apps/common/domain/storage';
import { z } from 'zod';
import { container } from '@/apps/admin/container';
import { type BuildInfo } from '@/apps/admin/domain/build';

export type Image = z.infer<typeof ImagesApiSchemas.Admin_ImageSpec_ImageInfo>;

type ImageExtraFields = {
	builds: BuildInfo[];
};
export type ImageDetail = z.infer<typeof ImagesApiSchemas.Admin_ImageSpec_Image> & ImageExtraFields;

export type CreateImagePayload = z.infer<
	typeof ImagesApiSchemas.Admin_ImageSpec_CreateImageRequestMultiPart
>;

export type UpdateImagePayload = z.infer<
	typeof ImagesApiSchemas.Admin_ImageSpec_UpdateImageRequestMultiPart
>;

export interface ImagesStorageInterface extends Storage<Image[]> {}
export const useImagesStorage = () => container.get('ImagesStorage') as ImagesStorageInterface;

export interface ImageStorageInterface extends Storage<ImageDetail> {}
export const useImageStorage = () => container.get('ImageStorage') as ImageStorageInterface;
