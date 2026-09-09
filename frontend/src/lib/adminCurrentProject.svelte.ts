import { browser } from '$app/environment';

// The whole admin dashboard (landing, submissions, clarifications, user-scores)
// needs a "current contest project" to key its admin-side fetches on. It used to
// borrow the contestant projectStorage for that, which only ever gets populated
// for users who are participants — so an organizer who isn't a participant had no
// way to choose one. This holds an admin-owned selection instead, picked from the
// admin sidebar and persisted per session so it survives navigation.
const CONTEST_KEY = 'adminContestId';
const CONTEST_NAME_KEY = 'adminContestName';
const PROJECT_KEY = 'adminProjectId';
const PROJECT_NAME_KEY = 'adminProjectName';
const PROJECT_START_KEY = 'adminProjectStartTime';
const PROJECT_END_KEY = 'adminProjectEndTime';

class AdminCurrentProject {
	contestId = $state<number | null>(null);
	contestName = $state<string | null>(null);
	projectId = $state<number | null>(null);
	projectName = $state<string | null>(null);
	// Contest window of the selected project, kept on the selection itself so the
	// phase banner doesn't depend on the shared admin contest-detail store (which
	// the management pages overwrite while browsing other contests).
	projectStartTime = $state<string | null>(null);
	projectEndTime = $state<string | null>(null);

	constructor() {
		// Server has no sessionStorage; the selection is a client concern only, so
		// the SSR value stays null and nothing leaks across requests.
		if (!browser) return;
		const contest = sessionStorage.getItem(CONTEST_KEY);
		const project = sessionStorage.getItem(PROJECT_KEY);
		if (contest) this.contestId = Number(contest);
		if (project) this.projectId = Number(project);
		this.contestName = sessionStorage.getItem(CONTEST_NAME_KEY);
		this.projectName = sessionStorage.getItem(PROJECT_NAME_KEY);
		this.projectStartTime = sessionStorage.getItem(PROJECT_START_KEY);
		this.projectEndTime = sessionStorage.getItem(PROJECT_END_KEY);
	}

	setContest(id: number | null, name: string | null) {
		this.contestId = id;
		this.contestName = name;
		this.#persist(CONTEST_KEY, id === null ? null : String(id));
		this.#persist(CONTEST_NAME_KEY, name);
	}

	setProject(
		id: number | null,
		name: string | null,
		startTime: string | null = null,
		endTime: string | null = null
	) {
		this.projectId = id;
		this.projectName = name;
		this.projectStartTime = startTime;
		this.projectEndTime = endTime;
		this.#persist(PROJECT_KEY, id === null ? null : String(id));
		this.#persist(PROJECT_NAME_KEY, name);
		this.#persist(PROJECT_START_KEY, startTime);
		this.#persist(PROJECT_END_KEY, endTime);
	}

	#persist(key: string, value: string | null) {
		if (!browser) return;
		if (value === null) sessionStorage.removeItem(key);
		else sessionStorage.setItem(key, value);
	}
}

export const adminCurrentProject = new AdminCurrentProject();
