import { expect, test } from '@playwright/test';
import { CONTESTANT_STATE, readState } from '../../fixtures';

test.use({ storageState: CONTESTANT_STATE });

// #2 選手 bootstrap 鏈:選定比賽 → 專案概述渲染 → SSE 倒數真的在跳
// (倒數凍結是正式環境發生過的事故,這條驗證 bootstrap + SSE proxy 整鏈)
test('@readonly 選手端:專案概述渲染且競賽倒數持續跳動', async ({ page }) => {
	const state = readState();
	// 用 bootstrap 既有的 sessionStorage 機制決定性地選定 E2E 場次,
	// 不依賴「恰好只有一場」的 auto-select
	await page.addInitScript(
		([cid, pid]) => {
			sessionStorage.setItem('currentContestId', String(cid));
			sessionStorage.setItem('currentProjectId', String(pid));
		},
		[state.contestId, state.contestProjectId]
	);

	await page.goto('/overview');
	await expect(page.getByText(state.projectName).first()).toBeVisible({ timeout: 15_000 });

	const clock = page.getByText(/\d+h\d+m\d+s/).first();
	await expect(clock).toBeVisible({ timeout: 15_000 });
	const before = await clock.innerText();
	await page.waitForTimeout(2500);
	expect(await clock.innerText(), '倒數應持續更新(SSE time 事件有到達)').not.toBe(before);
});
