/// <reference types="vite/client" />

interface ImportMetaEnv {
	/**
	 * 瀏覽器直呼後端 REST API 的位址。
	 *
	 * ⚠ 它也會在 **SSR** 時被 universal load 用到(見 `routes/(root)/+layout.ts`
	 * 取 profile),而那時的 fetch 是 SvelteKit 的:它只在「目標 host 等於本站 host
	 * 或其子網域」時才轉發瀏覽器的 cookie(**埠不影響**)。因此後端的 hostname 必須
	 * 與前端相同或為其子網域,否則 SSR 那一趟是未認證的,伺服器會把已登入的使用者
	 * 渲染成登出狀態。同 host 不同埠(localhost:5173 / localhost:8080)沒問題;
	 * judge.example.com 配 api.example.com 就會壞——那種佈署要改用
	 * api.judge.example.com,或把 profile 取用搬到 +layout.server.ts 自行轉發 cookie。
	 */
	VITE_API_URL: string;
	/**
	 * SvelteKit server(SSE 代理路由)呼叫後端的位址。與 VITE_API_URL 相同即可;
	 * 只有在前端伺服器與瀏覽器看到的後端位址不同時才需要分開設(例如後端在另一個
	 * 容器網路,而瀏覽器走轉發的埠)。用法見 src/lib/server/backend.ts。
	 */
	VITE_BACKEND_API_URL: string;
}
