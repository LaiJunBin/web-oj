import { expect, test } from '@playwright/test';
import { readState } from '../../fixtures';

// #5 kiosk 場次鎖定:匿名(大螢幕不登入)。
// 「錯誤的 pin 必須大聲失敗、絕不 fallback 到別場」是安全性行為,逐一回歸。
test('@readonly 合法 pin:顯示指定場次的記分板與文字', async ({ page }) => {
	const state = readState();
	await page.goto(
		`/scoreboard?kiosk&contestId=${state.contestId}&projectId=${state.contestProjectId}` +
			`&text=${encodeURIComponent('E2E 煙霧測試')}&footer=${encodeURIComponent('請保持安靜')}`
	);
	await expect(page.getByText('記分板').first()).toBeVisible({ timeout: 20_000 });
	await expect(page.getByText('E2E 煙霧測試')).toBeVisible();
	await expect(page.getByText('請保持安靜')).toBeVisible();
});

test('@readonly 不存在的比賽 id:顯示錯誤而非 fallback', async ({ page }) => {
	const state = readState();
	await page.goto(`/scoreboard?kiosk&contestId=999999&projectId=${state.contestProjectId}`);
	await expect(page.getByText('無法顯示指定的記分板')).toBeVisible({ timeout: 20_000 });
	await expect(page.getByText(/找不到網址指定的競賽/)).toBeVisible();
});

test('@readonly 比賽存在但專案不屬於它:顯示錯誤', async ({ page }) => {
	const state = readState();
	await page.goto(`/scoreboard?kiosk&contestId=${state.contestId}&projectId=999999`);
	await expect(page.getByText(/此競賽中找不到網址指定的專案/)).toBeVisible({ timeout: 20_000 });
});

test('@readonly 參數不完整:顯示錯誤', async ({ page }) => {
	const state = readState();
	await page.goto(`/scoreboard?kiosk&projectId=${state.contestProjectId}`);
	await expect(page.getByText(/網址參數不完整/)).toBeVisible({ timeout: 20_000 });
});
