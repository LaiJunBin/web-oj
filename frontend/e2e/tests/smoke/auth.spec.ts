import { expect, test } from '@playwright/test';
import { env } from '../../fixtures';

// #1 登入:一切的入口
test('@readonly 選手登入成功並離開登入頁', async ({ page }) => {
	await page.goto('/login');
	await page.fill('#username', env.contestant.username);
	await page.fill('input[type="password"]', env.contestant.password);
	await page.click('button[type="submit"]');
	await page.waitForURL((url) => !url.pathname.startsWith('/login'), { timeout: 15_000 });
	await expect(page.getByRole('button', { name: '登出' })).toBeVisible();
});

test('@readonly 錯誤密碼顯示登入失敗', async ({ page }) => {
	await page.goto('/login');
	await page.fill('#username', env.contestant.username);
	await page.fill('input[type="password"]', 'definitely-wrong');
	await page.click('button[type="submit"]');
	await expect(page.getByText('登入失敗')).toBeVisible({ timeout: 10_000 });
	await expect(page).toHaveURL(/\/login/);
});
