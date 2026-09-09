import { expect, test } from '@playwright/test';
import { ADMIN_STATE, readState } from '../../fixtures';
import { pickAdminProject } from '../../helpers';

test.use({ storageState: ADMIN_STATE });

// #4 admin picker 鏈:選定場次 → 階段列倒數 → 競賽過程頁以該場次為對象
test('@readonly 後台 picker 選定場次並顯示進行中階段列', async ({ page }) => {
	await page.goto('/dashboard/submissions');
	const state = readState();
	await pickAdminProject(page, state);
	await expect(page.getByText(/競賽進行中/)).toBeVisible({ timeout: 15_000 });
	await expect(page.getByText(`${state.projectName} 的完整提交`)).toBeVisible({
		timeout: 15_000
	});
});

test('@readonly 系統狀態頁:監控卡片齊備且列出 E2E 比賽', async ({ page }) => {
	await page.goto('/dashboard');
	const state = readState();
	// 評分工作者由後端啟動時自建(goroutine 常駐),數量必為正;
	// 顯示 0/0 代表清單 fetch 失敗 —— 比賽當天的監控台不能默默空白。
	await expect(page.getByRole('heading', { name: /評分工作者 [1-9]\d* \/ [1-9]\d* 運作中/ })).toBeVisible({
		timeout: 15_000
	});
	await expect(page.getByRole('heading', { name: /評分佇列/ })).toBeVisible();
	await expect(page.getByRole('heading', { name: /開放中的比賽/ })).toBeVisible();
	// 開放中的比賽卡片以連結列出比賽名(用 role 避開側欄 picker 的 option 撞名)
	await expect(page.getByRole('link', { name: state.contestName })).toBeVisible({
		timeout: 15_000
	});
});
