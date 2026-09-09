import { expect, test, type Locator } from '@playwright/test';
import { env, readState } from '../../fixtures';

// kiosk 大螢幕在串流退化時會掛出告示。它曾經是 fixed bottom-4、後來 bottom-24,
// 兩次都撞上底部公告 —— 因為公告文字來自 ?footer= 自由輸入、會自然換行,任何
// 固定偏移量都只是把界線往後推。這條測試用一段真的會換到多行的公告釘住「不重疊」。
//
// 版面只能在真瀏覽器裡驗:jsdom 的 clientHeight 一律是 0。
// 場地大螢幕的實際尺寸。釘死它,測試才不會因為預設 viewport 改變而悄悄失去意義 ——
// 公告帶子的高度取決於它在多寬的畫面上換幾行。
const KIOSK_VIEWPORT = { width: 1920, height: 1080 };

// 一段真的會換到四行以上的公告。長度不是隨手取的:舊的 bottom-24 是 96px,
// 而這段文字在上面的 viewport 下會把帶子撐到 96px 以上 —— 少於這個長度,
// 這條測試會在有 bug 的版本上也通過。
const LONG_FOOTER = (
	'請保持安靜,禁止交談與離座;如需協助請舉手,監考人員會前往處理。' +
	'本場次禁止攜帶任何電子裝置,違者以違規論處,並取消該題成績。'
).repeat(3);

async function box(locator: Locator) {
	const b = await locator.boundingBox();
	if (!b) throw new Error('element has no box');
	return b;
}

test.use({ viewport: KIOSK_VIEWPORT });

test('@readonly kiosk:串流退化的告示不會蓋住多行的底部公告', async ({ page }) => {
	const state = readState();

	// 讓每次連線都回報一個可重試的上游失敗,三次之後告示就會出現。
	await page.route('**/api/stream/public-contests/**', (route) =>
		route.fulfill({
			status: 200,
			headers: { 'content-type': 'text/event-stream', 'cache-control': 'no-cache' },
			// 心跳不是裝飾:Node 在寫出任何 body 之前不會把 headers 推到 socket,
			// 少了它瀏覽器根本不會認為連線已開啟。
			body:
				': keepalive\n\n' + 'event: upstream-error\ndata: {"stream":"scoreboard","status":500}\n\n'
		})
	);

	await page.goto(
		`/scoreboard?kiosk&contestId=${state.contestId}&projectId=${state.contestProjectId}` +
			`&text=${encodeURIComponent('E2E 煙霧測試')}&footer=${encodeURIComponent(LONG_FOOTER)}`
	);

	// 先讓記分板真的載入,之後才把 REST 回補也弄壞。上線實測的 bug 正是這裡:
	// 斷線後每次重連與每 30 秒的回補都打 REST,後端不在時失敗,而失敗曾經會把
	// 正在顯示的成績清成「Loading...」——手冊承諾的是「畫面停在最後一次更新」。
	// 用「已看到選手列」當開關而不是計數,不管初次載入抓了幾次都不會誤擋。
	const contestantRow = page.getByText(env.contestant.username).first();
	await expect(contestantRow, '記分板先正常載入').toBeVisible({ timeout: 30_000 });
	let boardLoaded = true;
	await page.route('**/api/v1/public/contest-projects/*/scoreboard', (route) =>
		boardLoaded
			? route.fulfill({ status: 503, contentType: 'application/json', body: '{"success":false}' })
			: route.continue()
	);

	const notice = page.getByRole('status').filter({ hasText: '即時連線異常' });
	await expect(notice, '退避 1s + 2s + 4s 之後才會亮').toBeVisible({ timeout: 30_000 });

	// 到這裡至少已經歷三次重連,每次都觸發一次失敗的回補。
	await expect(contestantRow, '回補失敗不能把最後一次的成績清掉').toBeVisible();
	await expect(page.getByText('Loading...'), '更不能退化成 Loading').not.toBeVisible();

	const footer = page.getByText(LONG_FOOTER);
	await expect(footer).toBeVisible();

	const n = await box(notice);
	const f = await box(footer);
	const overlaps =
		n.x < f.x + f.width && n.x + n.width > f.x && n.y < f.y + f.height && n.y + n.height > f.y;
	expect(overlaps, `告示 y=${n.y}..${n.y + n.height} 與公告 y=${f.y}..${f.y + f.height} 重疊`).toBe(
		false
	);
});
