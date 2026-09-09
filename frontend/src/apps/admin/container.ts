import { ProjectApplication } from '@/apps/admin/application/project';
import { ContestService } from '@/features/admin/service/contest';
import {
	ContestParticipantsStorage,
	ContestsStorage,
	ContestStorage
} from '@/features/admin/store/contest.svelte';
import { ProjectsStorage, ProjectStorage } from '@/features/admin/store/project.svelte';
import { ProjectService } from '@/features/admin/service/project';
import { Container } from '@/lib/container';
import { ContestApplication } from '@apps/admin/application/contest';
import { UserApplication } from '@/apps/admin/application/user';
import { UserService } from '@/features/admin/service/user';
import { UsersStorage, UserStorage } from '@/features/admin/store/user.svelte';
import { RoleApplication } from '@/apps/admin/application/role';
import { RoleService } from '@/features/admin/service/role';
import { RolesStorage } from '@/features/admin/store/role.svelte';
import { ContestProjectApplication } from '@/apps/admin/application/contest-project';
import { ContestProjectService } from '@/features/admin/service/contest-project';
import {
	ContestProjectStorage,
	JudgmentAnalyticsStorage,
	UserScoresStorage
} from '@/features/admin/store/contest-project.svelte';
import { ImageApplication } from '@/apps/admin/application/image';
import { ImageService } from '@/features/admin/service/image';
import { ImagesStorage, ImageStorage } from '@/features/admin/store/image';
import { BundleApplication } from '@/apps/admin/application/bundle';
import { BundleService } from '@/features/admin/service/bundle';
import { JudgeWorkerApplication } from '@/apps/admin/application/judge-worker';
import { JudgeWorkerService } from '@/features/admin/service/judge-worker';
import {
	JudgePriorityStrategiesStorage,
	JudgeWorkersStorage,
	SystemDefaultsStorage
} from '@/features/admin/store/judge-worker';
import { LifecycleApplication } from '@/apps/admin/application/lifecycle';
import { LifecycleService } from '@/features/admin/service/lifecycle';
import { GradingItemApplication } from '@/apps/admin/application/grading-item';
import { GradingItemService } from '@/features/admin/service/grading-item';
import { GradingItemsStorage, GradingItemStorage } from '@/features/admin/store/grading-item';
import { TestcaseApplication } from '@/apps/admin/application/testcase';
import {
	ProjectTestcaseGradingItemsStorage,
	TestcaseStorage
} from '@/features/admin/store/testcase';
import { TestcaseService } from '@/features/admin/service/testcase';
import { AssetApplication } from '@/apps/admin/application/asset';
import { AssetService } from '@/features/admin/service/asset';
import { ClarificationApplication } from '@/apps/admin/application/clarification';
import { ClarificationService } from '@/features/admin/service/clarification';
import { ClarificationsStorage, ClarificationStorage } from '@/features/admin/store/clarification';
import { SubmissionApplication } from '@/apps/admin/application/submission';
import { SubmissionService } from '@/features/admin/service/submission';
import { SubmissionsStorage, SubmissionStorage } from '@/features/admin/store/submission';
import { RejudgeApplication } from '@/apps/admin/application/rejudge';
import { RejudgeService } from '@/features/admin/service/rejudge';
import {
	RejudgesStorage,
	RejudgeStorage,
	RejudgeComparisonStorage
} from '@/features/admin/store/rejudge';
import { GroupApplication } from '@/apps/admin/application/group';
import { GroupService } from '@/features/admin/service/group';
import { GroupsStorage, GroupStorage } from '@/features/admin/store/group.svelte';
import { SystemStatusApplication } from '@/apps/admin/application/system-status';
import { SystemStatusService } from '@/features/admin/service/system-status';
import {
	JudgeQueueStorage,
	JudgmentErrorsStorage,
	SubmissionRateStorage
} from '@/features/admin/store/system-status';

export const container = new Container();
export function initAdminContainer() {
	// Role
	container.set(RoleApplication);
	container.set(RoleService);
	container.set(RolesStorage);

	// User
	container.set(UserApplication);
	container.set(UserService);
	container.set(UsersStorage);
	container.set(UserStorage);

	// Contest
	container.set(ContestApplication);
	container.set(ContestService);
	container.set(ContestsStorage);
	container.set(ContestStorage);
	container.set(ContestParticipantsStorage);

	// Project
	container.set(ProjectApplication);
	container.set(ProjectService);
	container.set(ProjectsStorage);
	container.set(ProjectStorage);

	// Asset
	container.set(AssetApplication);
	container.set(AssetService);

	// Grading Item
	container.set(GradingItemApplication);
	container.set(GradingItemService);
	container.set(GradingItemsStorage);
	container.set(GradingItemStorage);

	// Testcase
	container.set(TestcaseApplication);
	container.set(TestcaseService);
	container.set(TestcaseStorage);
	container.set(ProjectTestcaseGradingItemsStorage);

	// Lifecycle
	container.set(LifecycleApplication);
	container.set(LifecycleService);

	// Contest Project
	container.set(ContestProjectApplication);
	container.set(ContestProjectService);
	container.set(ContestProjectStorage);
	container.set(UserScoresStorage);
	container.set(JudgmentAnalyticsStorage);

	// Clarification
	container.set(ClarificationApplication);
	container.set(ClarificationService);
	container.set(ClarificationsStorage);
	container.set(ClarificationStorage);

	// Image
	container.set(ImageApplication);
	container.set(ImageService);
	container.set(ImagesStorage);
	container.set(ImageStorage);

	// Bundle (匯出/匯入)
	container.set(BundleApplication);
	container.set(BundleService);

	// Judge Worker
	container.set(JudgeWorkerApplication);
	container.set(JudgeWorkerService);
	container.set(JudgeWorkersStorage);
	container.set(JudgePriorityStrategiesStorage);
	container.set(SystemDefaultsStorage);

	// Submission
	container.set(SubmissionApplication);
	container.set(SubmissionService);
	container.set(SubmissionsStorage);
	container.set(SubmissionStorage);

	// Rejudge
	container.set(RejudgeApplication);
	container.set(RejudgeService);
	container.set(RejudgesStorage);
	container.set(RejudgeStorage);
	container.set(RejudgeComparisonStorage);

	// Group
	container.set(GroupApplication);
	container.set(GroupService);
	container.set(GroupsStorage);
	container.set(GroupStorage);

	// System Status
	container.set(SystemStatusApplication);
	container.set(SystemStatusService);
	container.set(JudgeQueueStorage);
	container.set(JudgmentErrorsStorage);
	container.set(SubmissionRateStorage);
}
