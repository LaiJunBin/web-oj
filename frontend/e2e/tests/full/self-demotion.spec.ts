import { expect, test } from '@playwright/test';
import { AdminApi, env, readState } from '../../fixtures';
import { seedContestantSession } from '../../helpers';

type Role = { id: number; name: string };
type WithMarker = Window & { __e2eNoReload?: boolean };

// 上線實測:admin 在使用者管理把自己的角色切成 contestant,導覽列不反應,要硬重整
// 才跟上 —— roles 只在根 layout 的 load 載入一次,client-side 導覽一路重用它。
// 這條測試用一個臨時帳號走真實 UI,釘住三件事:不重整(window 上的記號活著)、
// 後台入口立刻消失、提交鈕變成可按(contest 狀態以新身分重抓,拿到 isParticipant)。
test('@destructive admin 自降為 contestant:不重整,導覽列立即反應', async ({
	browser
}, testInfo) => {
	test.setTimeout(150_000);
	const baseURL = (testInfo.project.use.baseURL as string) ?? env.baseURL;
	const state = readState();
	const api = await AdminApi.login();
	const roles = await api.get<Role[]>('/roles');
	const roleId = (name: string) => {
		const role = roles.find((r) => r.name === name);
		if (!role) throw new Error(`seed 缺少角色 ${name}`);
		return role.id;
	};
	const username = `e2e-selfdemote-${Date.now()}`;
	const password = 'e2e-1234';
	const created = (await api.send('post', '/users', {
		username,
		name: 'E2E 自降權',
		password,
		roles: [roleId('admin')]
	})) as { data: number };
	const userId = created.data;
	// 先報名 E2E 比賽:降為選手之後提交鈕能不能按,取決於 contest 狀態有沒有以新身分重抓。
	await api.send('post', `/contests/${state.contestId}/participants/users`, { userIds: [userId] });

	const ctx = await browser.newContext({ baseURL });
	try {
		const page = await ctx.newPage();
		await seedContestantSession(page, state);
		await page.goto('/login');
		await page.fill('#username', username);
		await page.fill('input[type="password"]', password);
		await page.click('button[type="submit"]');
		await page.waitForURL((url) => !url.pathname.startsWith('/login'), { timeout: 15_000 });

		const adminLink = page.getByRole('link', { name: '後台管理' });
		const submit = page.getByRole('button', { name: '提交' });
		await expect(adminLink, '一開始是 admin').toBeVisible({ timeout: 15_000 });
		await expect(submit, '純 admin 不是參賽者,提交鈕是灰的').toBeDisabled({ timeout: 20_000 });

		await page.goto(`/dashboard/users/${userId}/edit`);
		// 只認可見的那個:flowbite 的 MultiSelect 是 div[role=listbox],表單用的 hidden
		// <select multiple> 也是 listbox,兩個一起抓會踩到 strict mode。
		const roleBox = page.locator('#role-group div[role="listbox"]');
		// 表單狀態看 hidden <select> 的選中項,比讀標籤文字可靠。
		const selected = page.locator('#role-select option:checked');
		await expect(selected, '一開始只有 admin').toHaveText(['admin'], { timeout: 15_000 });
		await page.evaluate(() => {
			(window as WithMarker).__e2eNoReload = true;
		});

		// 走 flowbite MultiSelect 的真實 UI:× 掉 admin 標籤、從下拉清單點 contestant。
		// 標籤 div 的文字是「admin」加上關閉鈕的 sr-only「Close」,所以不能 exact。
		await roleBox.getByText('admin').getByRole('button', { name: 'Remove badge' }).click();
		const dropdown = roleBox.locator('div[role="presentation"]').first();
		const option = dropdown.getByText('contestant', { exact: true });
		if (!(await option.isVisible())) await roleBox.click();
		await option.click();
		await expect(selected, '表單值已換成 contestant').toHaveText(['contestant']);
		// 下拉清單選完仍開著,而且蓋在「編輯」鈕上 —— 不關掉,click 會等到測試逾時。
		await page.keyboard.press('Escape');
		await expect(dropdown).toBeHidden();
		await page.locator('#submit-form').getByRole('button', { name: '編輯' }).click();

		await expect(adminLink, '後台入口立刻消失').not.toBeVisible({ timeout: 15_000 });
		await page.waitForURL((url) => !url.pathname.startsWith('/dashboard'), { timeout: 15_000 });
		expect(
			await page.evaluate(() => (window as WithMarker).__e2eNoReload),
			'是 invalidate 讓畫面跟上的,不是重整'
		).toBe(true);
		await expect(submit, '以選手身分重抓 contest 狀態後,提交鈕可按').toBeEnabled({
			timeout: 30_000
		});
	} finally {
		// 先刪臨時帳號:測試逾時後瀏覽器已被收掉,ctx.close() 會丟錯,不能讓它擋住清理。
		await api.send('delete', `/users/${userId}`).catch(() => undefined);
		await api.dispose();
		await ctx.close().catch(() => undefined);
	}
});
