import { expect, test } from '@playwright/test';
import path from 'node:path';
import { CONTESTANT_STATE, readState } from '../../fixtures';

test.use({ storageState: CONTESTANT_STATE });

const SAMPLE_ZIP = path.resolve(import.meta.dirname, '../../../../backend/tests/data/nginx.zip');

// #3 提交:比賽最核心的動作。不需評分完成 —— 驗到「收件並排入佇列」為止,
// 評分結果由 full 套件負責。
test('@destructive 選手上傳壓縮檔後出現在提交紀錄', async ({ page }) => {
	const state = readState();
	await page.addInitScript(
		([cid, pid]) => {
			sessionStorage.setItem('currentContestId', String(cid));
			sessionStorage.setItem('currentProjectId', String(pid));
		},
		[state.contestId, state.contestProjectId]
	);

	await page.goto('/overview');
	await expect(page.getByText(state.projectName).first()).toBeVisible({ timeout: 15_000 });

	// navbar 的提交按鈕(開賽期間內、已報名才會啟用)
	await page.getByRole('button', { name: '提交', exact: true }).first().click();
	await page.setInputFiles('#zip-file-input', SAMPLE_ZIP);
	await page.getByRole('dialog').getByRole('button', { name: '送出' }).click();
	await expect(page.getByText('已送出')).toBeVisible({ timeout: 20_000 });
	await page.getByRole('button', { name: 'OK' }).click();

	await page.goto('/submissions');
	// 選手端是卡片列表:標題含非零筆數,卡片帶類型與狀態徽章
	await expect(page.getByText(/提交紀錄 \([1-9]\d*\)/)).toBeVisible({ timeout: 15_000 });
	await expect(page.getByText('完整提交').first()).toBeVisible();
	await expect(page.getByText(/Created|Queued|Judging|Judged/).first()).toBeVisible();
});
