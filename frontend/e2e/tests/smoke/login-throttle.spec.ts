import { expect, request, test } from '@playwright/test';
import { env } from '../../fixtures';

// 登入的連續失敗保護會回 429,而畫面上必須說「這不是帳密錯誤」——講錯的代價
// 很具體:被擋下的人照著「請檢查帳號密碼」的提示一直重按,只會讓額度更難回復。
// 這條路徑先前壞過一次(狀態碼在應用層被壓平成單一種錯誤,型別檢查看不出來),
// 所以值得用真的 429 守著,而不是只靠單元測試。
//
// 之所以能穩定重現,靠兩件事:
//  1. 桶的 key 是 (來源 IP, 帳號),所以拿一個「不存在且沒有別人會用」的帳號去
//     榨,不會動到 demo01/admin 的額度,也不必和其他測試搶時序。
//  2. 只有失敗會扣額度,預設補充速率是每分鐘 6 次(每 10 秒一枚)。榨乾之後
//     有大約十秒的窗口讓 UI 送出那一發,而不是一秒。
//
// 隱含假設:榨額度的 API context(Node runner)與按下送出的瀏覽器要被後端視為
// 同一個來源 IP——同一台機器上跑(本機與 CI 的情況)自然成立。從別台機器對
// 遠端環境跑這條會打到不同的桶,測試會紅而不是 skip;它標了 @destructive,
// 本來就不該對著遠端環境跑。
const THROTTLE_USER = `e2e-throttle-${Date.now()}`;
const DRAIN_ATTEMPTS = 60;

test('@destructive 被登入頻率限制擋下時,畫面說的是「嘗試過於頻繁」而不是帳密錯誤', async ({
	page
}) => {
	const ctx = await request.newContext();
	const csrf = async () => (await (await ctx.get(`${env.apiURL}/csrf-token`)).json()).data.token;

	// 先把頁面備好、表單填好,榨乾後才有最短的時間差去送出
	await page.goto('/login');
	await page.fill('#username', THROTTLE_USER);
	await page.fill('input[type="password"]', 'whatever');

	let throttled = false;
	for (let i = 0; i < DRAIN_ATTEMPTS && !throttled; i++) {
		const res = await ctx.post(`${env.apiURL}/auth/login`, {
			headers: { 'csrf-token': await csrf() },
			data: { username: THROTTLE_USER, password: 'wrong' }
		});
		throttled = res.status() === 429;
	}
	await ctx.dispose();

	// 限流可以被部署關掉(burst 或 perMinute 非正)。那種環境下這條測不出東西,
	// 但那是設定選擇而不是回歸——說清楚並跳過,不要紅給人看。
	test.skip(
		!throttled,
		`${DRAIN_ATTEMPTS} 次失敗登入都沒有換到 429,這個環境的登入限流看來是關閉的`
	);

	await page.locator('button[type="submit"]').click();

	const dialog = page.locator('.swal2-popup');
	await expect(dialog).toBeVisible({ timeout: 20_000 });
	await expect(page.locator('.swal2-title')).toHaveText('嘗試過於頻繁');
	// 「不是帳密錯誤」是這句話存在的理由,值得直接釘住
	await expect(page.locator('.swal2-html-container')).toContainText('不是帳號密碼錯誤');
});
