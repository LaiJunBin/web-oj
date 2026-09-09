import { ScoreboardApplication } from '@/apps/contest/application/scoreboard';
import type { ScoreboardServiceInterface } from '@/apps/contest/application/service/scoreboard';
import type { Scoreboard, ScoreboardStorageInterface } from '@/apps/contest/domain/scoreboard';
import { describe, expect, test, vi } from 'vitest';

class FakeScoreboardStorage implements ScoreboardStorageInterface {
	private value: Scoreboard | null = null;
	error: Error | null = null;
	set(data: Scoreboard) {
		this.value = data;
		this.error = null;
	}
	get() {
		return this.value;
	}
	clear() {
		this.value = null;
	}
	// 和真正的 BrowserStorage 一樣:設錯誤時把資料清掉。這正是 bug 的形狀,
	// 假物件必須忠實複製它,測試才守得住東西。
	setError(error: Error) {
		this.value = null;
		this.error = error;
	}
}

const board = { heads: [], rows: [], compact: false } as unknown as Scoreboard;

function appWith(service: Pick<ScoreboardServiceInterface, 'fetchScoreboard'>) {
	const storage = new FakeScoreboardStorage();
	return { app: new ScoreboardApplication(service, storage), storage };
}

// 上線實測抓到的:kiosk 斷線後,每 30 秒的 REST 回補在後端還沒回來時失敗,
// fetchScoreboard 的 catch 走 setError,把正在顯示的最後一次成績清成
// 「Loading...」。手冊承諾的是「畫面停在最後一次更新」——救畫面的 fallback
// 不能在失敗時把畫面清掉。
describe('ScoreboardApplication.resyncScoreboard', () => {
	test('失敗時保留畫面上的資料,連 error 都不設', async () => {
		const { app, storage } = appWith({ fetchScoreboard: () => Promise.reject(new Error('502')) });
		storage.set(board);
		vi.spyOn(console, 'warn').mockImplementation(() => {});

		await app.resyncScoreboard(1);

		expect(storage.get(), '最後一次的成績要留在畫面上').toBe(board);
		expect(storage.error, '斷線已由告示表達,store 不必再記一筆').toBeNull();
	});

	test('成功時照常覆寫', async () => {
		const fresh = { ...board, rows: [{}] } as unknown as Scoreboard;
		const { app, storage } = appWith({ fetchScoreboard: () => Promise.resolve(fresh) });
		storage.set(board);

		await app.resyncScoreboard(1);

		expect(storage.get()).toBe(fresh);
	});

	test('對照:fetchScoreboard 失敗仍會清空——初次載入時那是對的,回補時不是', async () => {
		const { app, storage } = appWith({ fetchScoreboard: () => Promise.reject(new Error('502')) });
		storage.set(board);

		await app.fetchScoreboard(1);

		expect(storage.get()).toBeNull();
	});
});
