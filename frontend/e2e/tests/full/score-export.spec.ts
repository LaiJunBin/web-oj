import { expect, test } from '@playwright/test';
import fs from 'node:fs';
import { ADMIN_STATE, env, readState } from '../../fixtures';
import { ensureJudgedSubmission, pickAdminProject } from '../../helpers';

// #13 分數統計與匯出:選手成績列於統計頁,匯出檔可下載且非空。
test('@destructive 分數統計列出成績並可匯出明細', async ({ browser }, testInfo) => {
	test.setTimeout(240_000);
	const baseURL = (testInfo.project.use.baseURL as string) ?? env.baseURL;
	await ensureJudgedSubmission(browser, baseURL);
	const state = readState();

	const ctx = await browser.newContext({ baseURL, storageState: ADMIN_STATE });
	const page = await ctx.newPage();
	await page.goto('/dashboard/user-scores');
	await pickAdminProject(page, state);
	await expect(
		page.getByRole('cell', { name: new RegExp(env.contestant.username) })
	).toBeVisible({ timeout: 15_000 });

	const downloadPromise = page.waitForEvent('download');
	await page.getByRole('button', { name: '匯出分數明細' }).click();
	const download = await downloadPromise;
	expect(download.suggestedFilename()).toBeTruthy();
	const filePath = await download.path();
	expect(fs.statSync(filePath).size, '匯出檔不應為空').toBeGreaterThan(0);
	await ctx.close();
});
