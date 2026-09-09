import type { Project, ProjectStorageInterface } from '@/apps/contest/domain/project';
import { BrowserStorage } from '@features/common/store/base.svelte';

export class ProjectStorage extends BrowserStorage<Project> implements ProjectStorageInterface {
	set(data: Project): void {
		super.set(data);
		sessionStorage.setItem('currentProjectId', data.id.toString());
	}

	clear(): void {
		super.clear();
		sessionStorage.removeItem('currentProjectId');
	}
}
