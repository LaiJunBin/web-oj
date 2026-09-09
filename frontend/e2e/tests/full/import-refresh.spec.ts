import { expect, test } from '@playwright/test';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { ADMIN_STATE, AdminApi, readState } from '../../fixtures';

test.use({ storageState: ADMIN_STATE });

// 上線實測:後台「匯入」一場比賽後,導覽列的比賽下拉沒有它,要手動重新整理才出現。
// 匯入成功只刷新了後台自己的清單,沒有刷新導覽列讀的 contest 端 store。
test('@destructive 匯入比賽後,導覽列的比賽下拉立即列出它', async ({ page }) => {
	test.setTimeout(150_000);
	const state = readState();
	const api = await AdminApi.login();
	const zip = path.join(os.tmpdir(), `e2e-contest-${Date.now()}.zip`);
	fs.writeFileSync(zip, await api.download(`/contests/${state.contestId}/export`));
	const before = new Set((await api.get<{ id: number }[]>('/contests')).map((c) => c.id));
	let importedId: number | undefined;

	try {
		await page.goto('/dashboard/contests');
		await expect(page.getByRole('heading', { name: '競賽管理' })).toBeVisible({ timeout: 20_000 });
		const chooser = page.waitForEvent('filechooser');
		await page.getByRole('button', { name: '匯入', exact: true }).click();
		await (await chooser).setFiles(zip);

		const done = page.getByText(/已建立競賽「(.+?)」/);
		await expect(done, '匯入成功的提示').toBeVisible({ timeout: 90_000 });
		const name = (await done.textContent())!.match(/已建立競賽「(.+?)」/)![1];
		importedId = (await api.get<{ id: number }[]>('/contests'))
			.map((c) => c.id)
			.find((id) => !before.has(id));
		await page.locator('.swal2-confirm').click();

		// 不重新整理,直接打開導覽列的比賽下拉;只認導覽列裡的,後台表格也會列出同名的列。
		await page.locator('#dropdown-trigger').click();
		await expect(
			page.getByRole('navigation').getByText(name, { exact: true }),
			'匯入的比賽立刻出現在下拉裡'
		).toBeVisible({ timeout: 15_000 });
	} finally {
		if (importedId !== undefined) {
			await api.send('delete', `/contests/${importedId}`).catch(() => undefined);
		}
		await api.dispose();
		fs.rmSync(zip, { force: true });
	}
});
