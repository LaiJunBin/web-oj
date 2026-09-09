import { beforeEach, describe, expect, test, vi } from 'vitest';
import type { ContestDetail } from '@/apps/contest/domain/contest';

// BrowserStorage 在非瀏覽器環境下 set/clear 直接 return;這個測試要的就是瀏覽器行為。
vi.mock('$app/environment', () => ({ browser: true }));

import { ContestStorage } from './contest.svelte';

const contest = (id: number, projectIds: number[]) =>
	({ id, projects: projectIds.map((pid) => ({ id: pid })) }) as unknown as ContestDetail;

beforeEach(() => sessionStorage.clear());

// 記住的 project 什麼時候該被靜默丟掉,什麼時候該留給根 layout 決定(它丟掉時會提示)。
describe('ContestStorage.set 與記住的 project', () => {
	test('換到另一場比賽:舊的 project 記憶安靜丟掉', () => {
		sessionStorage.setItem('currentContestId', '1');
		sessionStorage.setItem('currentProjectId', '10');

		new ContestStorage().set(contest(2, [20]));

		expect(sessionStorage.getItem('currentContestId')).toBe('2');
		expect(sessionStorage.getItem('currentProjectId')).toBeNull();
	});

	test('同一場比賽重抓、記住的題目已不在:留著,交給根 layout 對清單並提示', () => {
		sessionStorage.setItem('currentContestId', '1');
		sessionStorage.setItem('currentProjectId', '10');

		new ContestStorage().set(contest(1, [11]));

		expect(sessionStorage.getItem('currentProjectId')).toBe('10');
	});

	test('第一次載入(沒有記住的比賽):什麼都不動', () => {
		sessionStorage.setItem('currentProjectId', '10');

		new ContestStorage().set(contest(1, [10]));

		expect(sessionStorage.getItem('currentContestId')).toBe('1');
		expect(sessionStorage.getItem('currentProjectId')).toBe('10');
	});
});
