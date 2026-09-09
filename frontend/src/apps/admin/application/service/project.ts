import type {
	CreateProjectPayload,
	Project,
	ProjectDetail,
	PutGradingProjectPayload,
	UpdateProjectPayload
} from '@/apps/admin/domain/project';

export interface ProjectServiceInterface {
	fetchProjects(): Promise<Project[]>;
	fetchProjectDetail(id: number): Promise<ProjectDetail>;
	createProject(payload: CreateProjectPayload): Promise<number>;
	updateProject(id: number, payload: UpdateProjectPayload): Promise<void>;
	deleteProject(id: number): Promise<void>;

	putGradingProject(id: number, payload: PutGradingProjectPayload): Promise<void>;
	downloadGradingProject(id: number): Promise<void>;
	readGradingProjectFile(id: number, file: string): Promise<string>;
	exportProject(id: number): Promise<void>;
	importProject(file: File): Promise<{ projectId: number; name: string; missingImageTags: string[] }>;
}
