import type { Project, SubmitProjectPayload } from '@apps/contest/domain/project';

export interface ProjectServiceInterface {
	fetchProject(id: number): Promise<Project>;
	downloadAsset(projectId: number, assetId: string, saveName: string): Promise<void>;
	submitProject(projectId: number, payload: SubmitProjectPayload): Promise<void>;
}
