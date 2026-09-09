import { apiClient } from '@/api';
import { ProjectsApi } from '@/api/gen/Projects';
import { downloadBlob, endpointPathByAlias, importFailureMessage } from './download';
import type { ProjectServiceInterface } from '@/apps/admin/application/service/project';
import type {
	CreateProjectPayload,
	Project,
	ProjectDetail,
	PutGradingProjectPayload,
	UpdateProjectPayload
} from '@/apps/admin/domain/project';

export class ProjectService implements ProjectServiceInterface {
	fetchProjects(): Promise<Project[]> {
		return apiClient.listProject().then((response) => {
			const { success, data } = response;
			if (!success) {
				throw new Error('Failed to fetch projects');
			}

			return data;
		});
	}
	fetchProjectDetail(projectId: number): Promise<ProjectDetail> {
		return Promise.all([
			apiClient
				.getProject({
					params: {
						id: projectId
					}
				})
				.then((response) => {
					const { success, data } = response;
					if (!success) {
						throw new Error('Failed to fetch project');
					}

					return data;
				}),
			apiClient
				.listGradingProjectFiles({
					params: {
						id: projectId
					}
				})
				.then((response) => {
					const { success, data } = response;
					if (!success) {
						throw new Error('Failed to fetch grading project files');
					}

					return data;
				})
		]).then(([project, files]) => {
			return {
				...project,
				files
			};
		});
	}
	createProject(payload: CreateProjectPayload): Promise<number> {
		return apiClient
			.createProject(payload)
			.then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('Failed to create project');
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
	updateProject(id: number, payload: UpdateProjectPayload): Promise<void> {
		return apiClient
			.updateProject(payload, {
				params: {
					id
				}
			})
			.then((response) => {
				const { success } = response;
				if (!success) {
					throw new Error('Failed to update project');
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
	deleteProject(id: number): Promise<void> {
		return apiClient
			.deleteProject(undefined, {
				params: {
					id
				}
			})
			.then((response) => {
				const { success } = response;
				if (!success) {
					throw new Error('Failed to delete project');
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
	putGradingProject(id: number, payload: PutGradingProjectPayload): Promise<void> {
		return apiClient
			.putGradingProject(payload, {
				params: {
					id
				}
			})
			.then((response) => {
				const { success } = response;
				if (!success) {
					throw new Error('Failed to put grading project');
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
	downloadGradingProject(id: number): Promise<void> {
		return apiClient
			.downloadGradingProject({
				params: {
					id
				},
				responseType: 'blob'
			})
			.then((blob) => {
				const url = window.URL.createObjectURL(blob as unknown as Blob);
				const a = document.createElement('a');
				a.download = `project.zip`;
				a.href = url;
				a.click();
			})
			.catch((error) => {
				console.error(error);
				throw new Error('下載檔案失敗');
			});
	}
	readGradingProjectFile(id: number, file: string): Promise<string> {
		return apiClient
			.readGradingProjectFile({
				params: {
					id,
					file
				}
			})
			.then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('Failed to read grading project file');
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

	exportProject(id: number): Promise<void> {
		const path = endpointPathByAlias(ProjectsApi, 'exportProject').replace(':id', String(id));
		return downloadBlob(path, `project-${id}.zip`, { failureMessage: '匯出題目失敗' });
	}

	importProject(file: File) {
		return apiClient
			.importProject({ bundle: file })
			.then((response) => {
				if (!response.success) throw new Error('匯入題目失敗');
				return response.data;
			})
			.catch((error) => {
				throw new Error(importFailureMessage(error, '匯入題目失敗'));
			});
	}
}
