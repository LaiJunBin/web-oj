import { z } from 'zod';
import { UserProjectApiSchemas } from '@/api/gen/UserProject';
import { PublicProjectApiSchemas } from '@/api/gen/PublicProject';
import { container } from '@/apps/contest/container';
import type { Storage } from '@apps/common/domain/storage';
import type { Submission } from './submission';
import type { UserApiSchemas } from '@/api/gen/User';

export interface ProjectScore {
	current: number;
	total: number;
}

type ProjectExtraFields = {
	submissions: Submission[];
	score: ProjectScore;
};

export type PublicContestProjectDetail = z.infer<
	typeof PublicProjectApiSchemas.Common_ProjectSpec_Project
> &
	ProjectExtraFields;

export type UserContestProjectDetail = z.infer<
	typeof UserProjectApiSchemas.User_ProjectSpec_ProjectDetail
> &
	ProjectExtraFields;

export type Project = PublicContestProjectDetail | UserContestProjectDetail;
export function isUserProject(project: Project): project is UserContestProjectDetail {
	return project && 'dynamicFields' in project;
}
export function isPublicProject(project: Project): project is PublicContestProjectDetail {
	return project && !('dynamicFields' in project);
}

export type SubmitProjectPayload = z.infer<
	typeof UserApiSchemas.User_UserSpec_SubmitProjectRequestMultiPart
>;

export interface ProjectStorageInterface extends Storage<Project> {}
export const useProjectStorage = () => container.get('ProjectStorage') as ProjectStorageInterface;
