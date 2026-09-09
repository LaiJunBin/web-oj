import { expect, test } from '@playwright/test';
import { ADMIN_STATE, CONTESTANT_STATE, env, readState } from '../../fixtures';
import { pickAdminProject, seedContestantSession } from '../../helpers';

// #12 提問/公告雙向鏈:選手提問 → admin 回覆+標記已解決 → 選手看到回覆;
// admin 發公告 → 選手看到公告。
test('@destructive 提問與公告的雙向流程', async ({ browser }, testInfo) => {
	test.setTimeout(240_000);
	const baseURL = (testInfo.project.use.baseURL as string) ?? env.baseURL;
	const state = readState();

	// 每次 attempt 的標題都不一樣。retry 會在同一場比賽再走一輪,留下上一輪的問答;
	// 標題若固定,row / 文字定位就會同時命中兩輪的資料,觸發 strict-mode violation。
	// 之前是靠 .first() 掩蓋,但那只是挑一個,挑到哪一輪並不受控。
	const tag = `${Date.now()}#${testInfo.retry}`;
	const QUESTION = `E2E 提問標題 ${tag}`;
	const REPLY = `E2E 回覆標題 ${tag}`;
	const ANNOUNCEMENT = `E2E 公告標題 ${tag}`;
	const QUESTION_BODY = `E2E 提問內容 ${tag}`;
	const REPLY_BODY = `E2E 回覆內容 ${tag}`;
	const ANNOUNCEMENT_BODY = `E2E 公告內容 ${tag}`;

	// 選手提問
	const cctx = await browser.newContext({ baseURL, storageState: CONTESTANT_STATE });
	const cpage = await cctx.newPage();
	await seedContestantSession(cpage, state);
	await cpage.goto('/clarifications');
	// .first():側欄連結與頁面標題都叫「提問與回應」
	await expect(cpage.getByText('提問與回應').first()).toBeVisible({ timeout: 15_000 });
	await cpage.getByRole('button', { name: '+' }).click();
	await cpage.fill('#subject-input', QUESTION);
	await cpage.fill('#content-input', QUESTION_BODY);
	await cpage.getByRole('dialog').getByRole('button', { name: '送出' }).click();
	await expect(cpage.getByText('提問成功')).toBeVisible({ timeout: 15_000 });
	await cpage.getByRole('button', { name: 'OK' }).click();

	// admin 回覆
	const actx = await browser.newContext({ baseURL, storageState: ADMIN_STATE });
	const apage = await actx.newPage();
	await apage.goto('/dashboard/clarifications');
	await pickAdminProject(apage, state);
	await expect(apage.getByText(QUESTION)).toBeVisible({ timeout: 15_000 });
	await apage.getByText(QUESTION).click();
	// .first():Accordion 標頭與表單的送出鈕都叫「回覆」
	await expect(apage.getByText('回覆').first()).toBeVisible({ timeout: 15_000 });
	await apage.fill('#subject-input', REPLY);
	await apage.fill('#content-input', REPLY_BODY);
	await apage.locator('form').getByRole('button', { name: '回覆' }).click();
	await expect(apage.getByText('回覆成功')).toBeVisible({ timeout: 15_000 });
	await apage.getByRole('button', { name: 'OK' }).click();

	// admin 標記已解決(列表上點狀態徽章切換)
	await apage.goto('/dashboard/clarifications');
	await pickAdminProject(apage, state);
	const questionRow = apage.getByRole('row').filter({ hasText: QUESTION });
	await expect(questionRow.getByText('未解決')).toBeVisible({ timeout: 15_000 });
	await questionRow.getByText('未解決').click();
	await expect(questionRow.getByText('已解決')).toBeVisible({ timeout: 15_000 });

	// admin 發公告
	await apage.getByRole('button', { name: '新增公告' }).click();
	await apage.getByRole('dialog').locator('#subject-input').fill(ANNOUNCEMENT);
	await apage.getByRole('dialog').locator('#content-input').fill(ANNOUNCEMENT_BODY);
	await apage.getByRole('dialog').getByRole('button', { name: '新增' }).click();
	await expect(apage.getByText('發布成功')).toBeVisible({ timeout: 15_000 });
	await actx.close();

	// 選手看到回覆與公告(先等 bootstrap 完成 —— 標題含專案名 —— 再斷言內容)
	await cpage.goto('/clarifications');
	await expect(cpage.getByText('提問與回應').first()).toBeVisible({ timeout: 15_000 });
	await expect(cpage.getByText(ANNOUNCEMENT)).toBeVisible({ timeout: 15_000 });
	// 回覆是**獨立的一則** clarification,關聯方向是「回覆 → 問題」,所以從問題本身
	// 查不到回覆——它的詳情頁永遠是「先前的內容 (0)」。直接斷言清單上的回覆那一列。
	const reply = cpage.getByText(REPLY);
	await expect(reply).toBeVisible({ timeout: 15_000 });
	await reply.click();
	await expect(cpage.getByText(REPLY_BODY)).toBeVisible({ timeout: 15_000 });
	await cctx.close();
});
