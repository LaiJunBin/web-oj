import '@testing-library/jest-dom/vitest';
import { render } from '@testing-library/svelte';
import { describe, expect, test, vi } from 'vitest';
import ShadowHtml from './ShadowHtml.svelte';

// 這個元件用來呈現選手提交的 stdout(評分項 outputFormat 為 html 時),也就是
// 「一位選手的輸出會被其他人的瀏覽器渲染」。它曾經把 html 裡的 <script> 抽出來
// 重新當成會執行的腳本注入,所以這裡守的是:樣式照舊呈現,但沒有任何東西會跑。
const shadowOf = (container: HTMLElement) =>
	container.ownerDocument.querySelector('[id^="shadow-"]')?.shadowRoot ?? null;

describe('ShadowHtml', () => {
	test('保留一般標記', async () => {
		const { container } = render(ShadowHtml, { html: '<p><b>粗體</b><i>斜體</i></p>' });
		const root = shadowOf(container);

		expect(root?.querySelector('b')?.textContent).toBe('粗體');
		expect(root?.querySelector('i')?.textContent).toBe('斜體');
	});

	test('不注入任何 script 元素', async () => {
		const { container } = render(ShadowHtml, {
			html: '<p>輸出</p><script>window.__SHADOW_XSS__ = true;</script>'
		});
		const root = shadowOf(container);

		expect(root?.querySelectorAll('script').length).toBe(0);
		expect(root?.innerHTML).not.toContain('__SHADOW_XSS__');
		expect(root?.querySelector('p')?.textContent).toBe('輸出');
	});

	test('拿掉事件處理器屬性', async () => {
		const onError = vi.fn();
		const { container } = render(ShadowHtml, {
			html: '<img src="x" onerror="window.__SHADOW_ONERROR__ = true">'
		});
		const root = shadowOf(container);

		expect(root?.querySelector('img')?.hasAttribute('onerror')).toBe(false);
		expect(onError).not.toHaveBeenCalled();
	});
});
