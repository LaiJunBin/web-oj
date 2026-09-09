import { describe, expect, test } from 'vitest';
import { nextSummaryScore } from '$lib/summaryScore';

// 上線 bug:部分提交 Judged 的瞬間,SSE handler 把側欄彙總分數推成它的得分;切頁後
// fetchProject 重算才歸回。規則是:部分提交永不採計彙總,只有完整提交能更新它 ——
// 與 fetchProject 的 isPartial ? 0 : score 一致。
describe('nextSummaryScore', () => {
	test('部分提交:不採計,維持原值(即使它的得分更高)', () => {
		expect(nextSummaryScore(0, 18, true)).toBe(0);
		expect(nextSummaryScore(12, 25, true)).toBe(12);
	});

	test('完整提交且更高:推高到新分數', () => {
		expect(nextSummaryScore(0, 18, false)).toBe(18);
		expect(nextSummaryScore(10, 25, false)).toBe(25);
	});

	test('完整提交但不更高:維持原值(取 max,不會被較低分蓋掉)', () => {
		expect(nextSummaryScore(20, 15, false)).toBe(20);
		expect(nextSummaryScore(18, 18, false)).toBe(18);
	});
});
