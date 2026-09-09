import { optionalNumber } from '@/lib';
import { ContestsApiSchemas } from '@/api/gen/Contests';
import { describe, expect, test } from 'vitest';

describe('optionalNumber', () => {
	test('清空欄位(空字串)得到 undefined,代表不送這個 key', () => {
		expect(optionalNumber('')).toBeUndefined();
	});

	test('null / undefined 同樣視為未填', () => {
		expect(optionalNumber(null)).toBeUndefined();
		expect(optionalNumber(undefined)).toBeUndefined();
	});

	// 這一題是重點:0 是「明確不限制」,與「未填 = 沿用系統預設」不同語意,
	// 用 truthy 判斷會把它一起吃掉。
	test('0 必須保留成 0,不可變成 undefined', () => {
		expect(optionalNumber('0')).toBe(0);
		expect(optionalNumber(0)).toBe(0);
	});

	test('字串數字轉成 number', () => {
		expect(optionalNumber('30')).toBe(30);
	});
});

// 後端的契約:未設定的欄位必須「不出現在 request body」。送 null 會被當成一個
// 明確的值,使用者會看到 200 卻發現設定沒生效。這裡連同實際的 zod schema 一起
// 驗,確保 schema parse 之後序列化出來仍然沒有這個 key。
describe('競賽更新 payload 的序列化', () => {
	const project = {
		id: 1,
		projectId: 1,
		name: '測試專案',
		activateTime: '2026-01-01T00:00:00Z',
		deactivateTime: '2026-01-02T00:00:00Z',
		startTime: '2026-01-01T01:00:00Z',
		endTime: '2026-01-01T02:00:00Z',
		isPublic: true,
		judgePriorityStrategy: 'default'
	};

	const serialize = (submissionCooldownSeconds: number | undefined) =>
		JSON.parse(
			JSON.stringify(
				ContestsApiSchemas.Admin_ContestSpec_UpdateRequestProject.parse({
					...project,
					submissionCooldownSeconds
				})
			)
		);

	test('未填時 key 不會出現在 body(而不是 null)', () => {
		const body = serialize(optionalNumber(''));
		expect('submissionCooldownSeconds' in body).toBe(false);
	});

	test('填 0 時 key 會出現且值為 0', () => {
		const body = serialize(optionalNumber('0'));
		expect(body.submissionCooldownSeconds).toBe(0);
	});
});
