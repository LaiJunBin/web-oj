# Web Judge — Frontend

SvelteKit 2 + Svelte 5(runes)+ Tailwind CSS 4 + flowbite-svelte。選手端、後台管理與 kiosk 記分板都在這一個 app;SSE 由 SvelteKit server route 同源代理(`src/routes/api/stream/*`)。整體架構與快速上手見[根目錄 README](../README.md)。

## 開發

需求:**Node.js ≥22.13**(`.nvmrc` 提供 22)與 **pnpm 11**——版本由 `package.json` 的
`packageManager` 欄位釘住,`corepack enable` 後執行 `pnpm` 會自動使用正確版本
(或 `npm i -g pnpm@11`)。不想在主機裝 Node?[`.devcontainer/`](.devcontainer/) 提供
VS Code Dev Container,「Reopen in Container」即可在隔離環境開發(主機只需 Docker)。

`docker-compose.yml` 為開發測試基底(**host network**:容器的 localhost 即主機的,
`.env` 預設值可直連本機後端——SSE 代理與 SSR 兩條伺服器端 fetch 皆然);它不是
部署工具,正式部署見下方「正式部署(adapter-node)」。

```bash
git submodule update --init   # clone 時沒帶 --recurse-submodules 才需要;gen:api-client 依賴 submodule 內容
cp .env.example .env          # 指向後端 API
pnpm install
pnpm gen:api-client           # 由 src/api/openapi.yaml 產生 typed client(src/api/gen/ 不入庫)
pnpm dev                      # http://localhost:5173
```

`gen:api-client` **免安裝**:openapi-zod-client 的 CLI 已 bundle 成獨立單檔入庫,直接以
node 執行(改到 submodule 原始碼時見 `openapi-zod-client/scripts/bundle-cli.mjs`)。
但它需要 submodule 的**內容**在——漏了 `--recurse-submodules` 的 clone 會以
`Cannot find module …/bundle/index.cjs` 失敗,在 repo 根目錄補跑
`git submodule update --init` 即可(仍不需安裝任何 submodule 相依)。

`gen:mocks` **需要先裝 submodule 的相依**——msw-auto-mock 的 `dist/cli.js` 是 tsup 產物,
把 `cac`、`lodash`、`ts-pattern` 等留成外部相依,沒裝會以 `Cannot find module 'cac'`
失敗:

```bash
cd msw-auto-mock && pnpm install && cd ..
pnpm gen:mocks           # 產出 src/mocks/handlers.js(入庫)與三個 gitignored 的進入點
```

msw 的 mock 只是本機開發用(`hooks.server.ts` 的 `server.listen()` 預設註解掉),
改了 API 規格才需要重跑。

API client 由 [openapi-zod-client](./openapi-zod-client)(submodule)產生;規格變更流程見 [backend/README.md](../backend/README.md#api-規格變更流程)。

## 正式部署(adapter-node)

```bash
pnpm build
ORIGIN=http://<這台機器對外的位址>:5173 PORT=5173 node build/index.js
```

**`ORIGIN` 必填。** 它是 SvelteKit 對 adapter-node 的要求,而在這個專案裡漏掉它的症狀很不直觀:SvelteKit 會在伺服器端發出的請求上蓋一個 `Origin` 標頭,值取自它推導的 app origin;推導錯了就不在後端 `server.allowedOrigins` 裡,後端整批回 **403**,於是**倒數凍結、記分板不更新、提問通知不到、後台網址直接輸入會被踢回登入頁**——而畫面上不會有任何一句話提到 CORS。

同一個位址也要加進後端的 `server.allowedOrigins`(瀏覽器自己的 CORS 也需要)。開發模式(`pnpm dev`)推導得出正確的 origin,所以這件事只有在建置後才會咬人。

**後端與前端必須同 hostname(或後端是前端的子網域)。** SSR 時取使用者資料的是 universal
load,用的是 SvelteKit 的 fetch,而它只在「目標 host 等於本站 host 或其子網域」時才轉發
瀏覽器的 cookie——**埠不影響**,所以 `localhost:5173` 配 `localhost:8080` 沒問題,
`judge.example.com` 配 `api.example.com` 就會壞:SSR 那一趟拿不到 cookie,伺服器把已登入的
使用者渲染成登出狀態,後台網址直接輸入會被導回登入頁。跨 host 的佈署請改成
`api.judge.example.com` 這種子網域形式。(SSE 代理不受這條限制,它自行轉發 Cookie 標頭。)

## 結構速覽

| 路徑 | 內容 |
|---|---|
| `src/apps/` | application + domain 層(自製 name-based DI container) |
| `src/features/` | service(打 API)與 store(`$state` 儲存)實作 |
| `src/routes/(root)/` | 選手端頁面;`dashboard/` 為後台;`manual/` 選手手冊 |
| `src/routes/api/stream/` | SSE 同源代理(多工後端串流、轉發續期 cookie) |
| `src/lib/server/sse-proxy.ts` | 代理核心(訊息重組 reframer、心跳、Set-Cookie 轉發) |
| `src/lib/server/backend.ts` | 代理「往後端那一端」的位址與認證(用哪個 URL、怎麼帶 cookie、為什麼不用 `event.fetch`) |
| `e2e/` | Playwright 端對端測試(自建比賽、參數化目標環境) |
| `static/manual/`、`static/reference/` | 兩份操作手冊的實拍截圖 |

## 品質檢查

```bash
pnpm run check                 # svelte-check;基準為 0 errors
pnpm test:unit                 # vitest
pnpm audit                     # 基準為 0 vulnerabilities(overrides 見 pnpm-workspace.yaml)
pnpm test:e2e --project=smoke  # 對運行中環境的冒煙 E2E(零設定即指向 localhost)
pnpm test:e2e --project=full   # 含評分管線端到端
```

E2E 參數(`E2E_BASE_URL`、`E2E_API_URL`、`E2E_ADMIN_USER/PASS`、`E2E_CONTESTANT_USER/PASS`)見 `e2e/fixtures.ts`;測試會自建專用比賽並於結束後清除,不依賴 seed 的時間窗。
