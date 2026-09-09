import type {
	Project,
	ProjectDetail,
	ProjectsStorageInterface,
	ProjectStorageInterface
} from '@/apps/admin/domain/project';
import { BrowserStorage } from '@features/common/store/base.svelte';

export class ProjectsStorage
	extends BrowserStorage<Project[]>
	implements ProjectsStorageInterface {}

export class ProjectStorage
	extends BrowserStorage<ProjectDetail>
	implements ProjectStorageInterface {}
