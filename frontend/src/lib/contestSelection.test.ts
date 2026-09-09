import { beforeEach, describe, expect, test } from 'vitest';
import { afterLoad, choose, forget, isGone, rememberedId } from '$lib/contestSelection';

beforeEach(() => sessionStorage.clear());

// 上線實測:資料庫重置或比賽被刪之後,sessionStorage 記住的 contestId 已不存在,
// bootstrap 拿它去抓 → 404 → 錯誤畫面「請稍後再試」——可是再試永遠不會成功。
// 修法分兩層:先對清單(不在就丟),真的抓了才看狀態碼(只有 404 是終局)。
describe('choose', () => {
	test('記住的 id 還在清單上:照舊載入', () => {
		expect(choose(2, [1, 2, 3])).toEqual({ id: 2, remembered: true, forgotten: false });
	});

	test('記住的 id 不在清單上:丟掉記憶、改選第一個,不去抓那個註定失敗的', () => {
		expect(choose(9, [1, 2, 3])).toEqual({ id: 1, remembered: false, forgotten: true });
	});

	test('沒有記憶:第一個', () => {
		expect(choose(null, [4, 5])).toEqual({ id: 4, remembered: false, forgotten: false });
	});

	test('清單是空的:沒東西可選,但過期的記憶照樣算失效', () => {
		expect(choose(9, [])).toEqual({ id: undefined, remembered: false, forgotten: true });
		expect(choose(null, [])).toEqual({ id: undefined, remembered: false, forgotten: false });
	});
});

// 後台「競賽過程」的預設比賽選擇也重用 choose:管理員多半從某場比賽的情境切進後台,
// 預設應跟隨 navbar 當前比賽(rememberedId('currentContestId')),而不是固定第一場。
// 優先序:後台自己記住的選擇(由呼叫端的 not-null 守衛處理)> navbar 記住且在後台清單中
// > 第一場。單向讀取,後台切換不回寫選手側記憶。
describe('admin 競賽過程 picker 預設(重用 choose)', () => {
	const ongoing = [11, 22, 33];

	test('navbar 記住的比賽在後台進行中清單裡 → 選中它', () => {
		expect(choose(22, ongoing)).toEqual({ id: 22, remembered: true, forgotten: false });
	});

	test('navbar 記住的比賽不在清單(非公開端差異或已刪)→ 落到第一場', () => {
		expect(choose(99, ongoing)).toEqual({ id: 11, remembered: false, forgotten: true });
	});

	test('沒有 navbar 記憶 → 第一場', () => {
		expect(choose(null, ongoing)).toEqual({ id: 11, remembered: false, forgotten: false });
	});

	test('後台清單為空 → 沒有可選,不硬塞', () => {
		expect(choose(22, []).id).toBeUndefined();
	});
});

describe('afterLoad', () => {
	const http = (status: number) => ({ response: { status } });

	test('記住的選擇 404:忘掉,重選', () => {
		expect(afterLoad(http(404), true)).toBe('forget');
	});

	test('記住的選擇 503:保留,錯誤畫面的「請稍後再試」才是真話', () => {
		expect(afterLoad(http(503), true)).toBe('keep');
	});

	test('網路錯誤(根本沒有 response):保留', () => {
		expect(afterLoad(new Error('Network Error'), true)).toBe('keep');
	});

	test('載入成功(沒有 error):保留', () => {
		expect(afterLoad(null, true)).toBe('keep');
	});

	test('預設選擇 404:沒有記憶可忘,維持錯誤畫面而不是無限重選', () => {
		expect(afterLoad(http(404), false)).toBe('keep');
	});

	test('isGone 只認 404', () => {
		expect(isGone(http(404))).toBe(true);
		expect(isGone(http(403))).toBe(false);
		expect(isGone(http(500))).toBe(false);
		expect(isGone(undefined)).toBe(false);
	});
});

describe('rememberedId / forget', () => {
	test('讀 sessionStorage;壞值一律視為沒有記憶', () => {
		expect(rememberedId('currentContestId')).toBeNull();
		for (const bad of ['abc', '0', '-1', '']) {
			sessionStorage.setItem('currentContestId', bad);
			expect(rememberedId('currentContestId'), `"${bad}"`).toBeNull();
		}
		sessionStorage.setItem('currentContestId', '7');
		expect(rememberedId('currentContestId')).toBe(7);
	});

	test('忘掉比賽也一併忘掉題目;忘掉題目不動比賽', () => {
		sessionStorage.setItem('currentContestId', '7');
		sessionStorage.setItem('currentProjectId', '70');
		forget('currentProjectId');
		expect(sessionStorage.getItem('currentContestId')).toBe('7');
		expect(sessionStorage.getItem('currentProjectId')).toBeNull();

		sessionStorage.setItem('currentProjectId', '70');
		forget('currentContestId');
		expect(sessionStorage.getItem('currentContestId')).toBeNull();
		expect(sessionStorage.getItem('currentProjectId')).toBeNull();
	});
});
