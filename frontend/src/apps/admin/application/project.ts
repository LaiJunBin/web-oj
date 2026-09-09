import { Application } from '@/apps/common/application/base';
import type { ProjectServiceInterface } from '@/apps/admin/application/service/project';
import type {
	CreateProjectPayload,
	ProjectsStorageInterface,
	ProjectStorageInterface,
	PutGradingProjectPayload,
	UpdateProjectPayload
} from '@/apps/admin/domain/project';
import { container } from '@/apps/admin/container';

export class ProjectApplication extends Application {
	constructor(
		private readonly projectService: ProjectServiceInterface,

		private readonly projectsStorage: ProjectsStorageInterface,
		private readonly projectStorage: ProjectStorageInterface
	) {
		super();
	}

	fetchProjects() {
		return this.projectService
			.fetchProjects()
			.then((projects) => {
				this.projectsStorage.set(projects);
			})
			.catch((error) => {
				this.projectsStorage.setError(error);
			});
	}

	fetchProject(id: number) {
		return this.projectService
			.fetchProjectDetail(id)
			.then((project) => {
				this.projectStorage.set(project);
			})
			.catch((error) => {
				this.projectStorage.setError(error);
			});
	}

	createProject(payload: CreateProjectPayload) {
		return this.projectService.createProject(payload);
	}

	updateProject(id: number, payload: UpdateProjectPayload) {
		return this.projectService.updateProject(id, payload);
	}

	deleteProject(id: number) {
		return this.projectService.deleteProject(id);
	}

	putGradingProject(id: number, payload: PutGradingProjectPayload) {
		return this.projectService.putGradingProject(id, payload);
	}

	downloadGradingProject(id: number) {
		return this.projectService.downloadGradingProject(id);
	}

	readGradingProjectFile(id: number, file: string) {
		return this.projectService.readGradingProjectFile(id, file);
	}

	exportProject(id: number) {
		return this.projectService.exportProject(id);
	}

	importProject(file: File) {
		return this.projectService.importProject(file);
	}
}

export const useProjectApp = () => container.get(ProjectApplication);
