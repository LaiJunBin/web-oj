import { apiClient } from '@/api';
import type { ProjectServiceInterface } from '@/apps/contest/application/service/project';
import type { UserProfile } from '@/apps/contest/application/user';
import type { Project, SubmitProjectPayload } from '@/apps/contest/domain/project';

export class ProjectService implements ProjectServiceInterface {
	constructor(private readonly userProfile: UserProfile) {}

	fetchProject(id: number): Promise<Project> {
		if (this.userProfile.isContestant) {
			return Promise.all([
				apiClient
					.getUserProject({
						params: {
							id
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
					.listUserProjectSubmissions({
						params: {
							id
						}
					})
					.then((response) => {
						const { success, data } = response;
						if (!success) {
							throw new Error('Failed to fetch project submissions');
						}

						return data;
					})
			]).then(([project, submissions]) => {
				const totalScore = project.gradingItems.reduce((acc, item) => acc + item.score, 0) ?? 0;
				const currentScore = Math.max(
					...submissions.map((submission) => (submission.isPartial ? 0 : submission.score)),
					0
				);
				return {
					...project,
					score: {
						current: currentScore,
						total: totalScore
					},
					submissions
				};
			});
		}

		return apiClient
			.getPublicProject({
				params: {
					id
				}
			})
			.then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('Failed to fetch project');
				}
				const totalScore = data.gradingItems.reduce((acc, item) => acc + item.score, 0) ?? 0;

				return {
					...data,
					score: {
						current: 0,
						total: totalScore
					},
					submissions: []
				};
			});
	}

	downloadAsset(projectId: number, assetId: string, saveName: string): Promise<void> {
		if (this.userProfile.isContestant) {
			return apiClient
				.downloadUserProjectAsset({
					params: {
						id: projectId,
						assetId: assetId
					},
					responseType: 'blob'
				})
				.then((blob) => {
					const url = window.URL.createObjectURL(blob as unknown as Blob);
					const a = document.createElement('a');
					a.download = saveName;
					a.href = url;
					a.click();
				})
				.catch((error) => {
					console.error(error);
					throw new Error('下載檔案失敗');
				});
		}
		return apiClient
			.downloadPublicProjectAsset({
				params: {
					id: projectId,
					assetId: assetId
				},
				responseType: 'blob'
			})
			.then((blob) => {
				const url = window.URL.createObjectURL(blob as unknown as Blob);
				const a = document.createElement('a');
				a.download = saveName;
				a.href = url;
				a.click();
			})
			.catch((error) => {
				console.error(error);
				throw new Error('請先登入');
			});
	}

	submitProject(projectId: number, payload: SubmitProjectPayload): Promise<void> {
		return apiClient
			.submitUserProject(payload, {
				params: {
					id: projectId
				}
			})
			.then((response) => {
				const { success } = response;
				if (!success) {
					throw new Error('Failed to submit project');
				}
			});
	}
}
