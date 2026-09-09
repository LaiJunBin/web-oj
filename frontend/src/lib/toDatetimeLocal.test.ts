import { toDatetimeLocal } from '@/lib';
import { describe, expect, test } from 'vitest';

describe('toDatetimeLocal', () => {
	// 這是真正要守住的性質:表單載入時用 toDatetimeLocal() 換算,送出時用
	// new Date(值).toISOString() 換回去(見 features/admin/service/contest.ts)。
	// 兩者必須互為反函數,否則管理員每按一次儲存,競賽時間就位移一次。
	// 斷言方式刻意不綁任何固定時區——不管測試跑在哪個時區都必須成立。
	const roundTrip = (iso: string) => new Date(toDatetimeLocal(iso)).toISOString();

	test.each([
		'2026-08-08T23:54:00.000Z',
		'2026-01-01T00:00:00.000Z',
		'2026-12-31T16:00:00.000Z',
		'2026-06-30T12:34:00.000Z'
	])('%s 來回換算後不變', (iso) => {
		expect(roundTrip(iso)).toBe(iso);
	});

	test('輸出符合 datetime-local 的格式(到分鐘、不帶時區)', () => {
		expect(toDatetimeLocal('2026-08-08T23:54:26.000Z')).toMatch(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}$/);
	});

	test('顯示的是本地牆鐘時間,而不是 UTC 字面值', () => {
		const iso = '2026-08-08T23:54:00.000Z';
		const expected = new Date(iso).toLocaleString('sv-SE').replace(' ', 'T').slice(0, 16);
		expect(toDatetimeLocal(iso)).toBe(expected);
	});

	test('省略參數時代表現在', () => {
		// 輸出只到分鐘,所以它必然是「現在」往前截斷不到一分鐘的時刻。前後各取一次
		// 時間,跨過分鐘邊界時也不會誤判——早期版本用 ±50 秒的容差,但截斷誤差最大
		// 是 60 秒,於是每分鐘有十幾秒的區間會無故失敗。
		const before = Date.now();
		const value = new Date(toDatetimeLocal()).getTime();
		const after = Date.now();

		expect(value).toBeLessThanOrEqual(after);
		expect(value).toBeGreaterThan(before - 60_000);
	});
});
