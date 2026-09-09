import { startFetches } from './sse-proxy';

/**
 * How the SSE proxy routes (`src/routes/api/stream/*`) reach the backend.
 *
 * **位址**:用 `VITE_BACKEND_API_URL` 而非 `VITE_API_URL`。後者是「瀏覽器看到的
 * 後端位址」,而這些程式碼跑在 SvelteKit server 上——兩者未必相同(例如後端在
 * 另一個容器網路,瀏覽器則走轉發的埠)。兩邊位址一樣時把兩個變數設成同值即可。
 *
 * **認證**:手動把瀏覽器送來的 Cookie header 轉發給後端。這不是暫時的變通,而是
 * 跨來源代理的正規做法——前端與後端是不同的 origin(即使是 compose 的預設安裝
 * 也是 :5173 對 :8080),伺服器端 fetch 不會、也不應該自動把使用者的 cookie 附加
 * 到跨來源請求上,那等同把憑證送給任意主機。`credentials: 'include'` 只在請求與
 * app 同源時才有意義。
 *
 * 反方向(後端續期 access token 後回傳的 Set-Cookie)由 sse-proxy 的
 * `readSetCookies` / `sseResponse` 負責轉回瀏覽器。
 *
 * **為什麼用平台的 `fetch` 而不是 SvelteKit 給的 `event.fetch`**:後者會在跨來源
 * 的出站請求上**補一個 `Origin` 標頭**(值是這個 app 自己的 origin),GET 也不例外。
 * 那個標頭是給瀏覽器的 CORS 用的,這裡是伺服器對伺服器的代理,它只會讓後端拿
 * 我們的 origin 去比對 `allowedOrigins`——推導出來的 origin 只要不在白名單裡
 * (adapter-node 未設 `ORIGIN` 環境變數時就會這樣),後端就回 403,而 403 會被
 * `resolveUpstream` 原樣送回瀏覽器,EventSource 隨即放棄:倒數凍結、記分板不更新、
 * 提問通知不到、提交狀態不動。dev server 推導得出正確的 origin,所以這個問題只在
 * 正式建置後才出現——別把它換回 `event.fetch`。
 */
const backendBaseUrl = import.meta.env.VITE_BACKEND_API_URL;

/** Prefixes a backend-relative path (e.g. `/judge-workers/stream`) with the backend base URL. */
export function backendUrl(path: string): string {
	return `${backendBaseUrl}${path}`;
}

/**
 * Starts parallel fetches to the backend for a set of named, backend-relative
 * paths, applying the base URL and forwarding the caller's cookies. Mirrors
 * {@link startFetches} but keeps the address/auth details in this module, so the
 * stream routes stay free of both.
 *
 * `fetchImpl` exists for tests; production must leave it at the platform fetch
 * (see the note above on why `event.fetch` is the wrong one here).
 */
export function startBackendFetches(
	paths: Record<string, string>,
	request: Request,
	signal: AbortSignal,
	fetchImpl: (url: string, init: RequestInit) => Promise<Response> = globalThis.fetch
): Record<string, Promise<Response>> {
	const requests = Object.fromEntries(
		Object.entries(paths).map(([name, path]) => [name, backendUrl(path)])
	);
	return startFetches(requests, fetchImpl, {
		signal,
		headers: { Cookie: request.headers.get('cookie') ?? '' }
	});
}
