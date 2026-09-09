import type { ProjectsApiSchemas } from '@/api/gen/Projects';
import type { Storage } from '@/apps/common/domain/storage';
import type { z } from 'zod';
import { container } from '@/apps/admin/container';

export type Project = z.infer<typeof ProjectsApiSchemas.Admin_ProjectSpec_Project>;

type ProjectExtraFields = {
	files: string[];
};
export type ProjectDetail = z.infer<typeof ProjectsApiSchemas.Admin_ProjectSpec_ProjectDetail> &
	ProjectExtraFields;

export type CreateProjectPayload = z.infer<
	typeof ProjectsApiSchemas.Admin_ProjectSpec_CreateProjectRequest
>;

export type UpdateProjectPayload = z.infer<
	typeof ProjectsApiSchemas.Admin_ProjectSpec_UpdateProjectRequest
>;

export type PutGradingProjectPayload = {
	project: File;
};

export interface ProjectsStorageInterface extends Storage<Project[]> {}
export const useProjectsStorage = () =>
	container.get('ProjectsStorage') as ProjectsStorageInterface;

export interface ProjectStorageInterface extends Storage<ProjectDetail> {}
export const useProjectStorage = () => container.get('ProjectStorage') as ProjectStorageInterface;
