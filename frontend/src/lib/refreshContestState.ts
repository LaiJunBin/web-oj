import { useContestApp } from '@/apps/contest/application/contest';
import { useProjectApp } from '@/apps/contest/application/project';
import { useProjectStorage } from '@/apps/contest/domain/project';

export function refreshContestState() {
	useContestApp().fetchContests();
}

export function resetContestState() {
	sessionStorage.removeItem('currentContestId');
	sessionStorage.removeItem('currentProjectId');
	useContestApp().fetchContests();
}

// Silently re-fetch the contest-app's currently selected project so participant-derived
// UI (the shared Navbar's submit button, gated on isParticipant) reflects admin-side
// participant changes immediately — even while the admin stays on the dashboard.
export function refreshContestProject() {
	const project = useProjectStorage().get();
	if (project) {
		useProjectApp().fetchProject(project.id);
	}
}
