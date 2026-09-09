import { sanitizeHtml } from '$lib/sanitize';
import { describe, expect, test } from 'vitest';

// 這些斷言守的是一個安全性質,不是實作細節:選手提交的 stdout 會以 HTML 呈現給
// 其他人看(evaluation 的 outputFormat: html),所以任何能執行的東西都必須被拿掉。
// 放在 client 專案是因為 sanitizeHtml 只有在有 DOM 時才走 DOMPurify;server 端
// 那條 escape 分支另外驗。
describe('sanitizeHtml', () => {
	test('拿掉 <script>', () => {
		const out = sanitizeHtml('<p>安全</p><script>alert(1)</script>');
		expect(out).toContain('安全');
		expect(out.toLowerCase()).not.toContain('<script');
		expect(out).not.toContain('alert(1)');
	});

	test('拿掉事件處理器屬性', () => {
		const out = sanitizeHtml('<img src="x" onerror="fetch(\'/steal\')">');
		expect(out.toLowerCase()).not.toContain('onerror');
		expect(out).not.toContain('/steal');
	});

	test('拿掉 javascript: URL', () => {
		const out = sanitizeHtml('<a href="javascript:alert(1)">點我</a>');
		expect(out.toLowerCase()).not.toContain('javascript:');
	});

	test('保留一般格式標籤', () => {
		const out = sanitizeHtml('<p><b>粗體</b><ul><li>項目</li></ul></p>');
		expect(out).toContain('<b>粗體</b>');
		expect(out).toContain('<li>項目</li>');
	});

	test('保留一般連結與圖片', () => {
		const out = sanitizeHtml('<a href="https://example.com">連結</a><img src="/a.png">');
		expect(out).toContain('href="https://example.com"');
		expect(out).toContain('src="/a.png"');
	});

	// svg/MathML 是繞過過濾器的經典路徑,USE_PROFILES: { html: true } 會一併擋掉
	test('拿掉 svg 內的 script', () => {
		const out = sanitizeHtml('<svg><script>alert(1)</script></svg>');
		expect(out).not.toContain('alert(1)');
	});
});
