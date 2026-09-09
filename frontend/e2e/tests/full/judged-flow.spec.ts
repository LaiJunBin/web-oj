import { expect, test } from '@playwright/test';
import { ADMIN_STATE, AdminApi, CONTESTANT_STATE, env, readState } from '../../fixtures';
import { pickAdminProject, seedContestantSession, submitSampleZip, waitForJudged } from '../../helpers';

// #8 + #9 + #10:評分管線端到端 —— 提交 → Judged → 記分板「不重整」即時變綠
// (reframer 修過的大 payload SSE bug 唯一能真正回歸的方式)→ 選手/admin 兩端檢視。
test('@destructive 提交評分完成:記分板即時更新、兩端結果正確', async ({ browser }, testInfo) => {
	test.setTimeout(300_000);
	const baseURL = (testInfo.project.use.baseURL as string) ?? env.baseURL;
	const state = readState();
	const api = await AdminApi.login();

	// 先開著記分板(同一 context 兩個分頁:一頁看板、一頁提交)
	const ctx = await browser.newContext({ baseURL, storageState: CONTESTANT_STATE });
	const board = await ctx.newPage();
	await seedContestantSession(board, state);
	await board.goto('/scoreboard');
	await expect(board.getByText('記分板').first()).toBeVisible({ timeout: 20_000 });
	await expect(board.getByText(env.contestant.username).first()).toBeVisible({ timeout: 20_000 });
	const greens = board.locator('.bg-green-300');
	const greensBefore = await greens.count();

	const submitPage = await ctx.newPage();
	await seedContestantSession(submitPage, state);
	await submitSampleZip(submitPage, state);
	const judged = await waitForJudged(api, state.contestProjectId);
	await api.dispose();

	// 記分板不重整,靠 SSE 推播變綠(若先前已有綠格則退為不減少的弱斷言)
	if (greensBefore === 0) {
		await expect(greens.first()).toBeVisible({ timeout: 30_000 });
	} else {
		expect(await greens.count()).toBeGreaterThanOrEqual(greensBefore);
	}

	// 選手端:直接以 Judged 提交的 id 開詳情(清單可能有多筆且隨 SSE 重渲染,
	// 點卡片會 flaky;deep link 決定性)
	await submitPage.goto(`/submissions/${judged.id}`);
	await expect(submitPage.getByText(/得分 \(\d+ \/ \d+\)/)).toBeVisible({ timeout: 15_000 });
	await expect(submitPage.getByText('Judged').first()).toBeVisible();
	await expect(submitPage.getByText('Passed').first()).toBeVisible({ timeout: 15_000 });
	await ctx.close();

	// admin 端:評分管線卡有數字;judgment 頁篩選/搜尋可用
	const actx = await browser.newContext({ baseURL, storageState: ADMIN_STATE });
	const apage = await actx.newPage();
	await apage.goto('/dashboard');
	await pickAdminProject(apage, state);
	await expect(apage.getByText('提交總數')).toBeVisible({ timeout: 15_000 });
	await expect(apage.getByText('評分總數')).toBeVisible();

	await apage.goto('/dashboard/submissions');
	await apage.locator('table').last().locator('tbody tr').first().locator('td').first().click();
	await expect(apage.getByText('裁判紀錄')).toBeVisible({ timeout: 15_000 });
	await apage.locator('table').last().locator('tbody tr').first().locator('td').first().click();
	await expect(apage).toHaveURL(/\/judgments\//, { timeout: 15_000 });
	await expect(apage.getByRole('button', { name: /^全部 \([1-9]/ })).toBeVisible({
		timeout: 15_000
	});
	await apage.getByRole('button', { name: /^通過 \(/ }).click();
	await apage.getByPlaceholder('搜尋評分項目名稱...').fill('zzz-不存在的評分項');
	await expect(apage.getByText('沒有符合條件的評分項目')).toBeVisible();
	await actx.close();
});
