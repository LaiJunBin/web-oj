import { browser } from '$app/environment';
import DOMPurify from 'dompurify';

/**
 * sanitizeHtml returns HTML safe to inject via {@html} or SweetAlert's `html:`
 * option: <script>, event-handler attributes (onerror, onload, …) and
 * javascript: URLs are stripped while ordinary formatting is kept. Use this
 * anywhere untrusted content — problem descriptions, backend error strings that
 * may echo bundle-derived text, imported file contents — is rendered as HTML.
 *
 * On the server there is no DOM for DOMPurify, so values are escaped instead;
 * every sink that uses this runs in the browser, so the escape branch is only a
 * belt-and-braces guard against ever emitting unsanitized markup during SSR.
 */
export function sanitizeHtml(html: string): string {
	if (!browser) {
		return escapeHtml(html);
	}
	return DOMPurify.sanitize(html, { USE_PROFILES: { html: true } });
}

/**
 * escapeHtml renders a value as literal text inside HTML. Use this instead of
 * {@link sanitizeHtml} when the point is to *show* the characters — a preview of
 * a file about to be imported, a raw error string — because sanitizing keeps
 * allowed markup as markup and drops the rest, so the reader sees something
 * other than what they are inspecting.
 */
export function escapeHtml(value: string): string {
	return value
		.replace(/&/g, '&amp;')
		.replace(/</g, '&lt;')
		.replace(/>/g, '&gt;')
		.replace(/"/g, '&quot;')
		.replace(/'/g, '&#39;');
}
