import { test as setup, expect } from '@playwright/test';
import fs from 'node:fs';
import {
	AdminApi,
	ADMIN_STATE,
	AUTH_DIR,
	CONTESTANT_STATE,
	env,
	writeState
} from '../fixtures';

// 自建一場「進行中」的 E2E 專用比賽,測試不依賴 seed 的時間窗:
// seed 只需提供帳號與至少一個專案素材,時間永遠以 now 為基準。
setup('provision e2e contest and auth states', async ({ browser }) => {
	// 長駐環境偶有慢窗(VM 負載、judge 容器 churn):給佈建流程寬裕時限
	setup.setTimeout(120_000);
	const api = await AdminApi.login();

	const projects = await api.get<{ id: number; name: string }[]>('/projects');
	expect(projects.length, 'seed 至少要有一個專案供 E2E 比賽掛載').toBeGreaterThan(0);
	const baseProject = projects[0];

	const now = Date.now();
	const iso = (ms: number) => new Date(ms).toISOString();
	const H = 3600_000;
	const contestName = `E2E Contest ${now}`;
	const created = (await api.send('post', '/contests', {
		name: contestName,
		description: 'created by the e2e suite; safe to delete',
		activateTime: iso(now - H),
		deactivateTime: iso(now + 24 * H),
		isPublic: true,
		projects: [
			{
				projectId: baseProject.id,
				name: `E2E ${baseProject.name}`,
				activateTime: iso(now - H),
				deactivateTime: iso(now + 24 * H),
				startTime: iso(now - 5 * 60_000),
				endTime: iso(now + 4 * H),
				isPublic: true,
				judgePriorityStrategy: 'default',
				// 測試會在數秒內連續提交;明確關掉提交冷卻,免得測試結果取決於
				// 目標環境的 judge.submissionCooldownSeconds 設定值。
				submissionCooldownSeconds: 0
			}
		]
	})) as { data: number };
	const contestId = created.data;
	const detail = await api.get<{ projects: { id: number; name: string }[] }>(
		`/contests/${contestId}`
	);
	const contestProject = detail.projects[0];

	// 報名選手
	const users = await api.get<{ id: number; username: string }[]>('/users');
	const contestant = users.find((u) => u.username === env.contestant.username);
	expect(contestant, `選手帳號 ${env.contestant.username} 應存在於 seed`).toBeTruthy();
	await api.send('post', `/contests/${contestId}/participants/users`, {
		userIds: [contestant!.id]
	});
	await api.dispose();

	writeState({
		contestId,
		contestName,
		contestProjectId: contestProject.id,
		projectName: contestProject.name
	});

	// 以真實 UI 登入一次,存下兩種身分的 storage state 供各 spec 重用
	fs.mkdirSync(AUTH_DIR, { recursive: true });
	for (const [account, statePath] of [
		[env.admin, ADMIN_STATE],
		[env.contestant, CONTESTANT_STATE]
	] as const) {
		const ctx = await browser.newContext({ baseURL: env.baseURL });
		const page = await ctx.newPage();
		await page.goto('/login');
		await page.fill('#username', account.username);
		await page.fill('input[type="password"]', account.password);
		await page.click('button[type="submit"]');
		// 登入成功會離開 /login(導向 / → /overview)
		await page.waitForURL((url) => !url.pathname.startsWith('/login'), { timeout: 15_000 });
		await ctx.storageState({ path: statePath });
		await ctx.close();
	}
});
