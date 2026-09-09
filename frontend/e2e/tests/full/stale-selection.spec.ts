import { expect, test } from '@playwright/test';
import { AdminApi, readState } from '../../fixtures';

const SWITCH_NOTICE = '先前選擇的比賽已不存在或無法存取,已為你切換';

// 上線實測:資料庫重置或比賽被刪之後,瀏覽器記住的 contestId 已不存在,bootstrap
// 拿它去抓 → 404 → 「無法載入比賽資料/請稍後再試」——但再試永遠不會成功,唯一出路
// 是使用者自己清 storage。比賽中管理員刪掉某場比賽,所有記著它的瀏覽器都會這樣。
test('@destructive 記住的比賽被刪後:重新整理自動落到剩下的比賽,並說一聲', async ({ page }) => {
	test.setTimeout(120_000);
	const api = await AdminApi.login();
	const projects = await api.get<{ id: number; name: string }[]>('/projects');
	const now = Date.now();
	const iso = (ms: number) => new Date(ms).toISOString();
	const H = 3600_000;
	const staleName = `E2E stale ${now}`;
	const created = (await api.send('post', '/contests', {
		name: staleName,
		description: 'created by the e2e suite; safe to delete',
		activateTime: iso(now - H),
		deactivateTime: iso(now + 24 * H),
		isPublic: true,
		projects: [
			{
				projectId: projects[0].id,
				name: `E2E stale ${projects[0].name}`,
				activateTime: iso(now - H),
				deactivateTime: iso(now + 24 * H),
				startTime: iso(now - 5 * 60_000),
				endTime: iso(now + 4 * H),
				isPublic: true,
				judgePriorityStrategy: 'default',
				submissionCooldownSeconds: 0
			}
		]
	})) as { data: number };
	const staleId = created.data;
	const navbar = page.getByRole('navigation');

	try {
		// 讓瀏覽器記住這場,並確認它真的顯示出來
		await page.goto('/overview');
		await page.evaluate((id) => sessionStorage.setItem('currentContestId', String(id)), staleId);
		await page.reload();
		await expect(navbar.getByText(staleName), '重新整理後顯示記住的那場').toBeVisible({
			timeout: 20_000
		});

		await api.send('delete', `/contests/${staleId}`);
		await page.reload();

		await expect(page.getByText(SWITCH_NOTICE), '不是靜默切換').toBeVisible({ timeout: 20_000 });
		await expect(page.getByText('無法載入比賽資料'), '不是死胡同').not.toBeVisible();
		await expect(navbar.getByText(staleName)).not.toBeVisible();
		await expect(navbar.getByText('競賽選擇'), '有比賽可切,不會停在未選').not.toBeVisible({
			timeout: 20_000
		});
		const remembered = Number(
			await page.evaluate(() => sessionStorage.getItem('currentContestId'))
		);
		expect(remembered, '記憶已換成新的那場').toBeGreaterThan(0);
		expect(remembered).not.toBe(staleId);
	} finally {
		await api.send('delete', `/contests/${staleId}`).catch(() => undefined);
		await api.dispose();
	}
});

// 對照組:5xx 不是「不存在」。後端把基礎設施故障回成 503,這時保留記憶、顯示錯誤畫面,
// 「請稍後再試」才是真話;亂切到別場比賽反而會讓使用者看到不同的標題而一頭霧水。
test('@readonly 記住的比賽暫時抓不到(503):保留選擇、顯示錯誤畫面,不亂切', async ({ page }) => {
	const state = readState();
	await page.goto('/overview');
	await page.evaluate(
		(id) => sessionStorage.setItem('currentContestId', String(id)),
		state.contestId
	);
	await page.route(`**/api/v1/public/contests/${state.contestId}`, (route) =>
		route.fulfill({ status: 503, contentType: 'application/json', body: '{"success":false}' })
	);
	await page.reload();

	await expect(page.getByText('無法載入比賽資料')).toBeVisible({ timeout: 20_000 });
	await expect(page.getByText(SWITCH_NOTICE)).not.toBeVisible();
	expect(
		await page.evaluate(() => sessionStorage.getItem('currentContestId')),
		'記憶保留,下次重新整理還會再試同一場'
	).toBe(String(state.contestId));
});
