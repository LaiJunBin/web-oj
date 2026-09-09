import { beforeEach, describe, expect, test, vi } from 'vitest';

// vi.mock 會被提升到 import 之前,工廠函式只能引用 vi.hoisted 出來的東西。
const mocks = vi.hoisted(() => ({
	invalidate: vi.fn(() => Promise.resolve()),
	goto: vi.fn(() => Promise.resolve()),
	refreshContestState: vi.fn(),
	logout: vi.fn(() => Promise.resolve()),
	fire: vi.fn(),
	userStorage: { set: vi.fn(), clear: vi.fn() },
	currentUsername: 'admin'
}));
vi.mock('$app/navigation', () => ({
	invalidate: mocks.invalidate,
	goto: mocks.goto
}));
vi.mock('$lib/refreshContestState', () => ({
	refreshContestState: mocks.refreshContestState
}));
vi.mock('@/apps/contest/application/user', () => ({
	useUserProfile: () => ({ username: mocks.currentUsername }),
	useUserApp: () => ({ logout: mocks.logout })
}));
vi.mock('sweetalert2', () => ({ default: { fire: mocks.fire } }));
// 根 layout 的 load 會把 profile 寫進 container 裡的 UserStorage;測試沒有 bootstrap
// 整個 container,所以只換掉這一個 hook,其餘(User 類別)照原樣。
vi.mock('@/apps/contest/domain/user', async (importOriginal) => ({
	...(await importOriginal<typeof import('@/apps/contest/domain/user')>()),
	useUserStorage: () => mocks.userStorage
}));

import {
	PROFILE_DEPENDENCY,
	isSelf,
	leaveAfterSelfDelete,
	refreshProfileIfSelf
} from '$lib/profileRefresh';
import { load } from '../routes/(root)/+layout';

beforeEach(() => {
	vi.clearAllMocks();
	mocks.currentUsername = 'admin';
});

// 上線實測:admin 在使用者管理把自己的角色切成 contestant,導覽列要硬重整才跟上。
// roles 只在根 layout 的 load 載入一次,client-side 導覽一路重用,所以「改到自己」
// 之後必須請 SvelteKit 重跑它;改別人則什麼都不必做。
describe('refreshProfileIfSelf', () => {
	test('改到自己:invalidate profile 依賴,並以新身分重抓 contest 狀態', async () => {
		await expect(refreshProfileIfSelf('admin')).resolves.toBe(true);

		expect(mocks.invalidate).toHaveBeenCalledWith(PROFILE_DEPENDENCY);
		// contest 狀態是「抓取當下的角色」決定的:只有 contestant 會拿到自己的比賽與
		// isParticipant(提交鈕的開關),所以 profile 之後還得重抓一次。
		expect(mocks.refreshContestState).toHaveBeenCalledTimes(1);
	});

	test('改到別人:不 invalidate、不重抓', async () => {
		await expect(refreshProfileIfSelf('demo01')).resolves.toBe(false);

		expect(mocks.invalidate).not.toHaveBeenCalled();
		expect(mocks.refreshContestState).not.toHaveBeenCalled();
	});

	test('重抓發生在 invalidate 完成之後(守門的 redirect 先走完)', async () => {
		const order: string[] = [];
		mocks.invalidate.mockImplementation(() => {
			order.push('invalidate');
			return Promise.resolve();
		});
		mocks.refreshContestState.mockImplementation(() => order.push('refresh'));

		await refreshProfileIfSelf('admin');

		expect(order).toEqual(['invalidate', 'refresh']);
	});
});

describe('isSelf', () => {
	test('以 username 比對(profile 沒有 id,而 username 不可能被改)', () => {
		mocks.currentUsername = 'someone';
		expect(isSelf('someone')).toBe(true);
		expect(isSelf('someone-else')).toBe(false);
	});
});

// 後端允許刪除自己;刪完後這個 session 的每個請求都是 401,但 JWT cookie 還在瀏覽器裡,
// 而登入頁會把帶著 cookie 的人一律彈回首頁 —— 不走一次 logout(它就算回 401 也會
// 送出清 cookie 的 Set-Cookie),使用者要等 cookie 過期才登得進來。
describe('leaveAfterSelfDelete', () => {
	test('先 logout 清 cookie,再回登入頁;logout 的 401 不能擋住流程', async () => {
		mocks.logout.mockRejectedValueOnce(new Error('Failed to logout'));

		await leaveAfterSelfDelete();

		expect(mocks.logout).toHaveBeenCalledTimes(1);
		expect(mocks.goto).toHaveBeenCalledWith('/login');
		expect(mocks.fire).toHaveBeenCalledWith(
			expect.objectContaining({ title: '已刪除你自己的帳號' })
		);
	});
});

// 另一半:根 layout 的 load 真的註冊了這個依賴。少了它,invalidate 什麼都不會重跑,
// 而上面的測試照樣全綠 —— 這條把兩端釘在一起。
describe('root layout load', () => {
	test('註冊 PROFILE_DEPENDENCY,讓 invalidate 找得到它', async () => {
		const depends = vi.fn();
		const fetch = vi.fn(() =>
			Promise.resolve({
				json: () =>
					Promise.resolve({
						success: true,
						data: { name: 'Admin', username: 'admin', roles: ['admin'] }
					})
			})
		);

		const result = await load({ fetch, depends } as unknown as Parameters<typeof load>[0]);

		expect(depends).toHaveBeenCalledWith(PROFILE_DEPENDENCY);
		expect(result?.user.username).toBe('admin');
		expect(mocks.userStorage.set).toHaveBeenCalledWith(result?.user);
	});
});
