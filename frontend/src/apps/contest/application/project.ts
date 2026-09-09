import { container } from '@/apps/contest/container';
import type { ContestStorageInterface } from '@apps/contest/domain/contest';
import {
	type ProjectStorageInterface,
	type SubmitProjectPayload
} from '@apps/contest/domain/project';
import { Application } from '@apps/common/application/base';
import type { ProjectServiceInterface } from './service/project';

export class ProjectApplication extends Application {
	constructor(
		private readonly projectService: ProjectServiceInterface,

		private readonly projectStorage: ProjectStorageInterface,
		private readonly contestStorage: ContestStorageInterface
	) {
		super();
	}

	fetchProject(id: number) {
		return this.projectService
			.fetchProject(id)
			.then((project) => {
				this.projectStorage.set(project);
			})
			.catch((error) => {
				this.projectStorage.setError(error);
			});
	}

	downloadAsset(projectId: number, assetId: string, saveName: string) {
		return this.projectService.downloadAsset(projectId, assetId, saveName);
	}

	submitProject(projectId: number, payload: SubmitProjectPayload) {
		return this.projectService.submitProject(projectId, payload);
	}

	getCurrentContestAvailableProjects() {
		return (
			this.contestStorage
				.get()
				?.projects.filter((project) => project.id !== this.projectStorage.get()?.id) || []
		);
	}

	// getCurrentProjectScore(): ProjectScore {
	// 	const project = this.projectStorage.get();
	// 	if (!project) return { currentScore: 0, totalScore: 0 };

	// 	const totalScore = project?.gradingItems.reduce((acc, item) => acc + item.score, 0) ?? 0;
	// 	const currentScore = isUserProjectDetail(project)
	// 		? Math.max(...project?.submissions.map((submission) => submission.score), 0)
	// 		: 0;
	// 	return {
	// 		currentScore,
	// 		totalScore
	// 	};
	// }
}

export const useProjectApp = () => container.get(ProjectApplication);
