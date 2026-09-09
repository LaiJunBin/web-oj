import { describe, expect, test } from 'vitest';
import { notifyStudentTemplates } from '$lib/notifyStudentTemplates';

// 「通知選手」的三個模板:用選手看得到的「提交時間」指涉那筆提交(不是選手看不到的
// Mongo id),送出前可改。
describe('notifyStudentTemplates', () => {
	const submittedAt = '2026-09-09 14:32:10';
	const templates = notifyStudentTemplates(submittedAt);

	test('三式,key 不重複', () => {
		expect(templates.map((t) => t.key)).toEqual(['checked', 'rejudged', 'resubmit']);
		expect(new Set(templates.map((t) => t.key)).size).toBe(3);
	});

	test('每個模板的標題與內容都帶入提交時間,且不再出現原始 id 前綴 #', () => {
		for (const t of templates) {
			expect(t.subject, t.key).toContain(submittedAt);
			expect(t.content, t.key).toContain(submittedAt);
			expect(t.content, t.key).not.toContain('#');
			expect(t.label.length, t.key).toBeGreaterThan(0);
		}
	});

	test('三式的措辭對得上檢查/重評/需重交', () => {
		expect(templates[0].content).toContain('檢查已有結果');
		expect(templates[1].content).toContain('已重新評測');
		expect(templates[2].content).toContain('需修正後重新提交');
	});
});
