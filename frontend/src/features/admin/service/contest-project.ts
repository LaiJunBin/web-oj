import { apiClient } from '@/api';
import { ContestProjectApi } from '@/api/gen/ContestProject';
import type { ContestProjectServiceInterface } from '@/apps/admin/application/service/contest-project';
import type { ContestProject, JudgmentAnalytics, RejudgeContestProjectPayload, UserScore } from '@/apps/admin/domain/contest-project';

const resolveEndpointPathByAlias = (alias: string) => {
	const endpoint = ContestProjectApi.find((item) => item.alias === alias);
	if (!endpoint) {
		throw new Error("Endpoint alias not found: " + alias);
	}
	return endpoint.path;
};

const resolveDownloadFilename = (contentDisposition?: string, fallback = "score-detail.xlsx") => {
	if (!contentDisposition) return fallback;

	const utf8Match = contentDisposition.match(/filename\*\s*=\s*UTF-8''([^;]+)/i);
	if (utf8Match?.[1]) {
		try {
			return decodeURIComponent(utf8Match[1].trim());
		} catch {
			// ignore malformed URI and continue with regular filename parsing
		}
	}

	const asciiMatch = contentDisposition.match(/filename\s*=\s*"([^"]+)"|filename\s*=\s*([^;]+)/i);
	if (asciiMatch) {
		const filename = (asciiMatch[1] ?? asciiMatch[2] ?? "").trim();
		if (filename) return filename;
	}

	return fallback;
};

const buildScoreDetailFallbackFilename = (contestProjectName: string) => {
	const normalizedName = contestProjectName.trim();
	const now = new Date();
	const pad = (value: number) => String(value).padStart(2, "0");
	const date = String(now.getFullYear()) + pad(now.getMonth() + 1) + pad(now.getDate());
	const time = pad(now.getHours()) + pad(now.getMinutes()) + pad(now.getSeconds());
	return normalizedName + "-" + date + "-" + time + "-score-detail.xlsx";
};

export class ContestProjectService implements ContestProjectServiceInterface {
	fetchContestProject(id: number): Promise<ContestProject> {
		return Promise.all([
			apiClient
				.getContestProject({
					params: {
						id
					}
				})
				.then((response) => {
					const { success, data } = response;
					if (!success) {
						throw new Error('Failed to fetch contest project');
					}

					return data;
				}),
			apiClient
				.getSubmissionsByContestProject({
					params: {
						id
					}
				})
				.then((response) => {
					const { success, data } = response;
					if (!success) {
						throw new Error('Failed to fetch user submissions');
					}

					return data;
				})
		]).then(([contestProject, submissions]) => {
			return {
				...contestProject,
				submissions
			};
		});
	}

	fetchUserScores(contestProjectID: number): Promise<UserScore[]> {
		return apiClient
			.getUserScores({
				params: {
					id: contestProjectID
				}
			})
			.then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('Failed to fetch user scores');
				}

				return data;
			});
	}

	fetchJudgmentAnalytics(contestProjectID: number): Promise<JudgmentAnalytics> {
		return apiClient
			.getContestProjectJudgmentAnalytics({
				params: {
					id: contestProjectID
				}
			})
			.then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('Failed to fetch judgment analytics');
				}

				return data;
			});
	}

	exportContestProjectScoreDetail(contestProjectID: number, contestProjectName: string): Promise<void> {
		const endpointPathTemplate = resolveEndpointPathByAlias("exportContestProjectScoreDetail");
		const endpointPath = endpointPathTemplate.replace(":id", String(contestProjectID));

		return apiClient.axios
			.get(endpointPath, {
				responseType: "blob",
				withCredentials: true
			})
			.then((response) => {
				const filename = resolveDownloadFilename(
					response.headers["content-disposition"],
					buildScoreDetailFallbackFilename(contestProjectName)
				);
				const blobURL = URL.createObjectURL(response.data);
				const anchor = document.createElement("a");
				anchor.href = blobURL;
				anchor.download = filename;
				anchor.click();
				setTimeout(() => URL.revokeObjectURL(blobURL), 0);
			})
			.catch((error) => {
				console.error(error);
				if ("response" in error) {
					const { data } = error.response;
					if (data?.message) {
						throw new Error(data.message);
					}
				}
				throw new Error("匯出分數明細失敗");
			});
	}

	rejudgeContestProject(id: number, payload: RejudgeContestProjectPayload): Promise<void> {
		return apiClient
			.rejudgeContestProject(payload, {
				params: {
					id
				}
			})
			.then((response) => {
				const { success } = response;
				if (!success) {
					throw new Error('Failed to rejudge contest project');
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
}
