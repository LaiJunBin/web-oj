import { expect, test } from '@playwright/test';
import { ADMIN_STATE, AdminApi, env, readState } from '../../fixtures';
import { ensureJudgedSubmission, pickAdminProject } from '../../helpers';

type RejudgeItem = {
	id: string;
	status: string;
	completedJudgments: number;
	effectedJudgments: number;
};

// #11 重評鏈:單筆重評 → 進度完成 → 核准 → 前後比對頁
test('@destructive 重評:發起、完成、核准與前後比對', async ({ browser }, testInfo) => {
	test.setTimeout(300_000);
	const baseURL = (testInfo.project.use.baseURL as string) ?? env.baseURL;
	await ensureJudgedSubmission(browser, baseURL);
	const state = readState();
	const api = await AdminApi.login();

	const ctx = await browser.newContext({ baseURL, storageState: ADMIN_STATE });
	const page = await ctx.newPage();
	await page.goto('/dashboard/submissions');
	await pickAdminProject(page, state);
	await page.locator('table').last().locator('tbody tr').first().locator('td').first().click();
	await expect(page.getByText('裁判紀錄')).toBeVisible({ timeout: 15_000 });

	await page.getByRole('button', { name: '請求重新評分' }).click();
	await page.locator('.swal2-input').fill('E2E rejudge');
	await page.getByRole('button', { name: '確定' }).click();
	await expect(page.getByText('操作成功')).toBeVisible({ timeout: 30_000 });
	await page.getByRole('button', { name: 'OK' }).click();

	// 等重評跑完(Finished 待審核)
	const deadline = Date.now() + 180_000;
	let rejudge: RejudgeItem | undefined;
	for (;;) {
		const list = await api.get<RejudgeItem[]>(
			`/contest-projects/${state.contestProjectId}/rejudges`
		);
		rejudge = list.find((r) => r.status === 'Finished');
		if (rejudge) break;
		if (Date.now() > deadline) {
			throw new Error(`rejudge did not finish: ${JSON.stringify(list)}`);
		}
		await new Promise((r) => setTimeout(r, 3000));
	}
	await api.dispose();

	// 重評紀錄(picker 範圍)顯示進度,點入詳情核准
	await page.goto('/dashboard/rejudgings');
	await expect(page.getByText('E2E rejudge').first()).toBeVisible({ timeout: 15_000 });
	await expect(
		page.getByText(`${rejudge.completedJudgments} / ${rejudge.effectedJudgments}`).first()
	).toBeVisible();
	// 整列點擊(row onclick 導頁),等 URL 切到詳情;
	// 「重新評分原因」在列表是兩個欄頭、在詳情是 heading,用 role 區分
	await page.getByRole('row').filter({ hasText: 'E2E rejudge' }).first().click();
	await expect(page).toHaveURL(/\/rejudgings\/[^/]+$/, { timeout: 15_000 });
	await expect(page.getByRole('heading', { name: '重新評分原因' })).toBeVisible({
		timeout: 15_000
	});

	await page.getByRole('button', { name: '做出決策' }).click();
	await page.getByRole('button', { name: '通過' }).click();
	await expect(page.getByText('操作成功')).toBeVisible({ timeout: 30_000 });
	await page.getByRole('button', { name: 'OK' }).click();
	await expect(page.getByText('Approved').first()).toBeVisible({ timeout: 15_000 });

	// 前後比對頁
	await expect(page.getByText('此次重新評分受影響的提交')).toBeVisible();
	await page.locator('table').last().locator('tbody tr').first().locator('td').first().click();
	await expect(page.getByText('評分前後比對')).toBeVisible({ timeout: 15_000 });
	await expect(page.getByText('總分(前 → 後)')).toBeVisible();
	await ctx.close();
});
