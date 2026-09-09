import { apiClient } from '@/api';
import { ContestsApi } from '@/api/gen/Contests';
import { downloadBlob, endpointPathByAlias, importFailureMessage } from './download';
import type { ContestServiceInterface } from '@/apps/admin/application/service/contest';
import type {
	Contest,
	ContestDetail,
	ContestParticipants,
	CreateContestPayload,
	UpdateContestPayload
} from '@/apps/admin/domain/contest';

export class ContestService implements ContestServiceInterface {
	fetchContests(): Promise<Contest[]> {
		return apiClient.listContest().then((response) => {
			const { success, data } = response;
			if (!success) {
				throw new Error('Failed to fetch contests');
			}

			return data;
		});
	}
	fetchContestDetail(contestId: number): Promise<ContestDetail> {
		return apiClient
			.getContest({
				params: {
					id: contestId
				}
			})
			.then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('Failed to fetch contest');
				}

				return data;
			});
	}
	createContest(payload: CreateContestPayload): Promise<number> {
		const newPayload: CreateContestPayload = {
			...payload,
			activateTime: `${new Date(payload.activateTime).toISOString().slice(0, 16)}:00Z`,
			deactivateTime: `${new Date(payload.deactivateTime).toISOString().slice(0, 16)}:00Z`,
			projects: payload.projects.map((project) => ({
				...project,
				startTime: `${new Date(project.startTime).toISOString().slice(0, 16)}:00Z`,
				endTime: `${new Date(project.endTime).toISOString().slice(0, 16)}:00Z`,
				activateTime: `${new Date(project.activateTime).toISOString().slice(0, 16)}:00Z`,
				deactivateTime: `${new Date(project.deactivateTime).toISOString().slice(0, 16)}:00Z`
			}))
		};

		return apiClient
			.createContest(newPayload)
			.then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('Failed to create contest');
				}

				return data;
			})
			.catch((error) => {
				if ('response' in error) {
					const { data } = error.response;
					throw new Error(data.message);
				}

				throw error;
			});
	}
	updateContest(id: number, payload: UpdateContestPayload): Promise<void> {
		const newPayload: UpdateContestPayload = { ...payload };
		if (newPayload.activateTime) {
			newPayload.activateTime = `${new Date(newPayload.activateTime).toISOString().slice(0, 16)}:00Z`;
		}
		if (newPayload.deactivateTime) {
			newPayload.deactivateTime = `${new Date(newPayload.deactivateTime).toISOString().slice(0, 16)}:00Z`;
		}
		if (newPayload.projects) {
			for (let i = 0; i < newPayload.projects.length; i++) {
				if (newPayload.projects[i].startTime) {
					newPayload.projects[i].startTime =
						`${new Date(newPayload.projects[i].startTime!).toISOString().slice(0, 16)}:00Z`;
				}
				if (newPayload.projects[i].endTime) {
					newPayload.projects[i].endTime =
						`${new Date(newPayload.projects[i].endTime!).toISOString().slice(0, 16)}:00Z`;
				}
				if (newPayload.projects[i].activateTime) {
					newPayload.projects[i].activateTime =
						`${new Date(newPayload.projects[i].activateTime!).toISOString().slice(0, 16)}:00Z`;
				}
				if (newPayload.projects[i].deactivateTime) {
					newPayload.projects[i].deactivateTime =
						`${new Date(newPayload.projects[i].deactivateTime!).toISOString().slice(0, 16)}:00Z`;
				}
			}
		}

		return apiClient
			.updateContest(newPayload, {
				params: {
					id
				}
			})
			.then((response) => {
				const { success } = response;
				if (!success) {
					throw new Error('Failed to update contest');
				}
			})
			.catch((error) => {
				if ('response' in error) {
					const { data } = error.response;
					throw new Error(data.message);
				}

				throw error;
			});
	}
	deleteContest(id: number): Promise<void> {
		return apiClient
			.deleteContest(undefined, {
				params: {
					id
				}
			})
			.then((response) => {
				const { success } = response;
				if (!success) {
					throw new Error('Failed to delete contest');
				}
			})
			.catch((error) => {
				if ('response' in error) {
					const { data } = error.response;
					throw new Error(data.message);
				}

				throw error;
			});
	}

	fetchContestParticipants(id: number): Promise<ContestParticipants> {
		return apiClient
			.listContestParticipants({
				params: {
					id
				}
			})
			.then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('Failed to fetch contest participants');
				}

				return data;
			});
	}

	addUsersToContest(id: number, userIds: number[]): Promise<void> {
		return apiClient
			.addUsersToContest(
				{ userIds },
				{
					params: {
						id
					}
				}
			)
			.then((response) => {
				const { success } = response;
				if (!success) {
					throw new Error('Failed to add users to contest');
				}
			});
	}

	removeUsersFromContest(id: number, userIds: number[]): Promise<void> {
		return apiClient
			.removeUsersFromContest(
				{ userIds },
				{
					params: {
						id
					}
				}
			)
			.then((response) => {
				const { success } = response;
				if (!success) {
					throw new Error('Failed to remove users from contest');
				}
			});
	}

	addGroupsToContest(id: number, groupIds: number[]): Promise<void> {
		return apiClient
			.addGroupsToContest(
				{ groupIds },
				{
					params: {
						id
					}
				}
			)
			.then((response) => {
				const { success } = response;
				if (!success) {
					throw new Error('Failed to add groups to contest');
				}
			});
	}

	removeGroupsFromContest(id: number, groupIds: number[]): Promise<void> {
		return apiClient
			.removeGroupsFromContest(
				{ groupIds },
				{
					params: {
						id
					}
				}
			)
			.then((response) => {
				const { success } = response;
				if (!success) {
					throw new Error('Failed to remove groups from contest');
				}
			});
	}

	exportContest(id: number): Promise<void> {
		const path = endpointPathByAlias(ContestsApi, 'exportContest').replace(':id', String(id));
		return downloadBlob(path, `contest-${id}.zip`, { failureMessage: '匯出競賽失敗' });
	}

	importContest(file: File) {
		return apiClient
			.importContest({ bundle: file })
			.then((response) => {
				if (!response.success) throw new Error('匯入競賽失敗');
				return response.data;
			})
			.catch((error) => {
				throw new Error(importFailureMessage(error, '匯入競賽失敗'));
			});
	}
}
