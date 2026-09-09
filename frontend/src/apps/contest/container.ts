import { Container } from '$lib/container';
import { AuthService } from '@features/contest/service/auth';
import { UserStorage } from '@features/contest/store/user.svelte';
import { UserApplication, UserProfile } from '@/apps/contest/application/user';
import { ContestApplication } from '@apps/contest/application/contest';
import { ContestService } from '@features/contest/service/contest';
import {
	ContestScheduleStorage,
	ContestsStorage,
	ContestStorage
} from '@features/contest/store/contest.svelte';
import { ProjectApplication } from '@apps/contest/application/project';
import { ProjectStorage } from '@features/contest/store/project.svelte';
import { ProjectService } from '@features/contest/service/project';
import { ClarificationApplication } from '@apps/contest/application/clarification';
import { ClarificationService } from '@features/contest/service/clarification';
import {
	ClarificationsStorage,
	ClarificationStorage
} from '@features/contest/store/clarification.svelte';
import { SubmissionApplication } from '@apps/contest/application/submission';
import { SubmissionService } from '@features/contest/service/submission';
import { SubmissionStorage } from '@features/contest/store/submission.svelte';
import { SettingStorage } from '@features/contest/store/setting.svelte';
import { ScoreboardApplication } from '@apps/contest/application/scoreboard';
import { ScoreboardService } from '@features/contest/service/scoreboard';
import { ScoreboardStorage } from '@features/contest/store/scoreboard.svelte';
import { TimeStorage } from '@features/contest/store/time.svelte';

export const container = new Container();

export function initContestContainer() {
	// Common
	container.set(SettingStorage);
	container.set(TimeStorage);

	// User
	container.set(UserApplication);
	container.set(AuthService);
	container.set(UserStorage);
	container.set(UserProfile);

	// Contest
	container.set(ContestApplication);
	container.set(ContestService);
	container.set(ContestStorage);
	container.set(ContestsStorage);
	container.set(ContestScheduleStorage);

	// Project
	container.set(ProjectApplication);
	container.set(ProjectService);
	container.set(ProjectStorage);

	// Clarification
	container.set(ClarificationApplication);
	container.set(ClarificationService);
	container.set(ClarificationsStorage);
	container.set(ClarificationStorage);

	// Submission
	container.set(SubmissionApplication);
	container.set(SubmissionService);
	container.set(SubmissionStorage);

	// Scoreboard
	container.set(ScoreboardApplication);
	container.set(ScoreboardService);
	container.set(ScoreboardStorage);
}
