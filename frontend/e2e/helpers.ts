import { expect, type Browser, type Locator, type Page } from '@playwright/test';
import path from 'node:path';
import { AdminApi, CONTESTANT_STATE, env, readState, type E2EState } from './fixtures';

export const SAMPLE_ZIP = path.resolve(import.meta.dirname, '../../backend/tests/data/nginx.zip');

/** 用 bootstrap 既有的 sessionStorage 機制決定性選定 E2E 場次(選手端) */
export async function seedContestantSession(page: Page, state: E2EState) {
	await page.addInitScript(
		([cid, pid]) => {
			sessionStorage.setItem('currentContestId', String(cid));
			sessionStorage.setItem('currentProjectId', String(pid));
		},
		[state.contestId, state.contestProjectId]
	);
}

/**
 * 等一個選項出現,逾時前 reload 一次再試。
 *
 * 側欄的比賽清單是 dashboard layout 在 **onMount 抓一次**的
 * (`dashboard/+layout.svelte`),之後不再重取,所以只有兩種卡法,而它們要的解法
 * 相反:
 *
 * - **慢**(慢 runner 上 hydration + 那一支 API 還沒回來):要的是耐心。這裡量過
 *   —— 把該 API 人為延遲 25 秒,三次 10 秒的 reload 重試全部失敗,因為每次
 *   reload 都把還在飛的 fetch 取消、從頭再來,愈重試愈慢。
 * - **卡住**(那一次 fetch 失敗或永遠不回):等再久也沒用,只有 reload 會重抓。
 *
 * 所以先給一段足夠長的等待,真的過了才 reload 一次當最後手段。
 */
const PICKER_SLICE_MS = 45_000;

async function selectOptionWhenLoaded(page: Page, select: Locator, value: number, what: string) {
	const option = select.locator(`option[value="${value}"]`);
	try {
		await expect(option).toBeAttached({ timeout: PICKER_SLICE_MS });
	} catch {
		// 等到這裡,「慢」已經不成立了;剩下的可能是那一次 fetch 根本沒成功。
		await page.reload();
		await expect(
			option,
			`側欄 picker 等不到${what} ${value}:資料已確認在 API 上,所以這是前端載入問題,` +
				`不是測試與 setup 之間的時序競態`
		).toBeAttached({ timeout: PICKER_SLICE_MS });
	}
	await select.selectOption(String(value));
}

/**
 * 側欄 picker:等選項出現後明確選取 E2E 場次(admin 端,不依賴 auto-select)。
 *
 * 先確認這場比賽真的在 admin 的清單 API 回應裡。它幾乎一定在(setup 建立完才會
 * 跑測試),但這一步讓失敗訊息能分辨兩件完全不同的事:資料還沒好,或是資料好了
 * 而畫面沒跟上。
 */
export async function pickAdminProject(page: Page, state: E2EState) {
	const res = await page.request.get(`${env.apiURL}/contests`);
	expect(res.ok(), `admin 清單 API 回應 ${res.status()}`).toBe(true);
	const contests = ((await res.json()).data ?? []) as { id: number }[];
	expect(
		contests.map((c) => c.id),
		`比賽 ${state.contestId} 應該已經在 admin 清單 API 裡`
	).toContain(state.contestId);

	await selectOptionWhenLoaded(page, page.locator('aside select').first(), state.contestId, '比賽');
	await selectOptionWhenLoaded(
		page,
		page.locator('aside select').nth(1),
		state.contestProjectId,
		'專案'
	);
}

/** 走真實 UI 上傳範例壓縮檔(navbar 提交 modal) */
export async function submitSampleZip(page: Page, state: E2EState) {
	await page.goto('/overview');
	await expect(page.getByText(state.projectName).first()).toBeVisible({ timeout: 15_000 });
	await page.getByRole('button', { name: '提交', exact: true }).first().click();
	await page.setInputFiles('#zip-file-input', SAMPLE_ZIP);
	await page.getByRole('dialog').getByRole('button', { name: '送出' }).click();
	await expect(page.getByText('已送出')).toBeVisible({ timeout: 20_000 });
	await page.getByRole('button', { name: 'OK' }).click();
}

export type AdminSubmission = { id: string; status: string };

/** 輪詢 admin API 直到本場次出現 Judged 的提交(評分管線端到端完成) */
export async function waitForJudged(
	api: AdminApi,
	contestProjectId: number,
	timeoutMs = 180_000
): Promise<AdminSubmission> {
	const deadline = Date.now() + timeoutMs;
	for (;;) {
		const subs = await api.get<AdminSubmission[]>(
			`/contest-projects/${contestProjectId}/submissions`
		);
		const judged = subs.find((s) => s.status === 'Judged');
		if (judged) return judged;
		const errored = subs.find((s) => s.status === 'Error');
		if (errored) throw new Error(`submission ${errored.id} errored during judging`);
		if (Date.now() > deadline) {
			throw new Error(`timed out waiting for a Judged submission (${JSON.stringify(subs)})`);
		}
		await new Promise((r) => setTimeout(r, 3000));
	}
}

/**
 * 確保本場次已有一筆 Judged 提交(沒有就以選手身分實際提交並等評分完成),
 * 讓 full 套件各 spec 不必依賴彼此的執行順序。
 */
export async function ensureJudgedSubmission(browser: Browser, baseURL: string) {
	const state = readState();
	const api = await AdminApi.login();
	try {
		const subs = await api.get<AdminSubmission[]>(
			`/contest-projects/${state.contestProjectId}/submissions`
		);
		if (subs.some((s) => s.status === 'Judged')) return;
		const ctx = await browser.newContext({ baseURL, storageState: CONTESTANT_STATE });
		const page = await ctx.newPage();
		await seedContestantSession(page, state);
		await submitSampleZip(page, state);
		await ctx.close();
		await waitForJudged(api, state.contestProjectId);
	} finally {
		await api.dispose();
	}
}
