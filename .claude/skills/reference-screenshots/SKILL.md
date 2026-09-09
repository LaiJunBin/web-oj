---
name: reference-screenshots
description: Regenerate the in-app manual screenshots (frontend/static/reference and static/manual) by staging live contest data and driving the running app with Playwright. Use when UI changes make the manuals' screenshots stale, or after the seeder changes.
---

# Regenerating manual screenshots

The admin manual (`/dashboard/reference`, images in `frontend/static/reference/`)
and contestant manual (`/manual`, images in `frontend/static/manual/`) use real
screenshots of the running system. This skill captures the whole workflow:
verify the environment, stage data, drive the browser, restore state.

## 1. Environment check

- Frontend dev server: Vite at `http://localhost:5173` (`pnpm dev` in
  `frontend/` if not running).
- The browser calls the API at whatever `VITE_API_URL` says (repo default
  `http://localhost:8080/api/v1`; check `frontend/.env`). That address must be
  reachable **from the machine running the browser** — in setups where the
  backend runs on another host/VM behind a forwarded port, the forward may not
  exist where this skill runs. Probe it first; if dead, start a local TCP
  proxy to wherever the backend actually listens (replace `<backend-host>` and
  the ports to match your `VITE_API_URL` / backend):

```bash
curl -s -o /dev/null -w "%{http_code}\n" --max-time 2 "$VITE_API_URL/contests"
# 000 → start a proxy (leave it running in the background):
node -e "
const net=require('node:net');
net.createServer(c=>{const u=net.connect(8080,'<backend-host>');
c.pipe(u).on('error',()=>{});u.pipe(c).on('error',()=>{});
c.on('error',()=>u.destroy());u.on('error',()=>c.destroy());
}).listen(8080,()=>console.log('proxy up'));" &
```

- Credentials come from the seeder — **verify against the current seed before
  staging**. As of the demo-data rework: admin `admin/1234`, contestants
  `demo01`–`demo05` (all `1234`), one contest 「示範競賽:靜態網頁架設」 (id 1)
  with a single project 「第一題:用 Nginx 部署靜態網站」. The seeder's text lives
  in `backend/database/seeder/seeders/demo_data.json`.

## 2. Stage data (and ALWAYS restore it)

The demo contest now seeds a 7-day window that is already in progress, so
staging is usually unnecessary — check the window first and only extend it if
you are screenshotting a stale database. When you do need to change it, use the
admin API and **save the original first**:

```bash
J=/tmp/refstage; mkdir -p $J
# csrf + login (cookie jar)
curl -s -c $J/cookies.txt $VITE_API_URL/csrf-token > $J/csrf.json
TOKEN=$(sed -E 's/.*"token":"([^"]+)".*/\1/' $J/csrf.json)
curl -s -b $J/cookies.txt -c $J/cookies.txt -H "Content-Type: application/json" \
  -H "csrf-token: $TOKEN" -X POST $VITE_API_URL/auth/login \
  -d '{"username":"admin","password":"1234"}' -o /dev/null
# SAVE ORIGINAL TIMES (restore payload comes from this)
curl -s -b $J/cookies.txt $VITE_API_URL/contests/1 > $J/contest1-original.json
# extend: activate now-1h → now+48h, start now-30m, end now+4h (nice countdown)
# build payload with node, then PATCH (method is PATCH, not PUT):
TOKEN=$(curl -s -b $J/cookies.txt -c $J/cookies.txt $VITE_API_URL/csrf-token | sed -E 's/.*"token":"([^"]+)".*/\1/')
curl -s -b $J/cookies.txt -H "Content-Type: application/json" -H "csrf-token: $TOKEN" \
  -X PATCH $VITE_API_URL/contests/1 -d @$J/extend.json
```

`extend.json` shape (update PATCH is per-field partial; project entries are
matched by `id` and only the provided fields change):

```json
{ "activateTime": "...", "deactivateTime": "...",
  "projects": [ { "id": 1, "activateTime": "...", "startTime": "...",
                  "endTime": "...", "deactivateTime": "..." } ] }
```

If the submissions/judgment pages need data and none exists: submit
`backend/tests/data/nginx.zip` as a contestant through the navbar 提交 modal
(file input `#zip-file-input`), then poll
`GET /api/v1/contest-projects/1/submissions` until status `Judged` (the judge
pipeline usually finishes in well under a minute).

**After shooting, PATCH the original times back** (build the restore payload
from `contest1-original.json`: top-level activate/deactivate + per-project
id/activate/start/end/deactivate). A leftover test submission is acceptable —
tell the user so they can ignore/delete it from the dashboard.

## 3. Drive the browser

`@playwright/test` + chromium are installed in `frontend/`. **Scripts must live
under `frontend/`** (module resolution) — write a `.tmp.mjs`, run, delete:

```js
import { chromium } from '@playwright/test';
const browser = await chromium.launch({ args: ['--no-sandbox'] });
const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 } });
const p = await ctx.newPage();
// login
await p.goto('http://localhost:5173/login');
await p.fill('#username', 'admin');
await p.fill('input[type="password"]', '1234');
await p.click('button[type="submit"]');
await p.waitForTimeout(2500);
```

Key patterns (learned the hard way):

- **Two-stage readiness gate before EVERY shot** — this is the difference
  between a manual and a folder of spinners:
  1. wait for the global loading overlay to be gone:
     `page.waitForSelector('div.absolute.inset-0.z-50', { state: 'hidden' })`
  2. wait for a data element **unique to the page content**. Never wait on
     text that also appears in the sidebar/navbar (競賽日程, 重評紀錄, 公告,
     記分板 all match nav links and fire before data loads) — prefer
     `getByRole('heading', …)` or a data string like `示範選手 01`.
  After both, a short `waitForTimeout(1200)` settles SSE numbers.
- **Eyeball every image before committing** (Read the png). A slow fetch, a
  dropped proxy connection, or a died session produces a plausible-looking
  "OK" from the driver while the frame shows Loading / 尚未登入.

- **Admin picker** (drives all 競賽過程 pages + status page's pipeline card):
  `aside select` nth(0)=contest, nth(1)=project. A single ongoing contest
  auto-selects; only `selectOption` when `inputValue()` is empty. Select the
  picker BEFORE shooting any scoped page.
- Wait on a **page-specific selector** (`text=評分佇列`, `table`,
  `text=完整提交`, `text=裁判紀錄`, `text=記分板`…), then settle
  `waitForTimeout(1800–3000)` for SSE data to arrive. Never rely on
  networkidle (SSE keeps connections open forever).
- Row navigation: `p.locator('table').last().locator('tbody tr').first()
  .locator('td').first().click()` (row onclick only fires on TD).
- Long pages (contest detail, submission detail, judgment): `fullPage: true`.
- Kiosk shot: **anonymous context** + pinned URL
  `/scoreboard?kiosk&contestId=1&projectId=1&text=…&footer=…` (encodeURIComponent
  the Chinese), wait `text=記分板`, settle 3s.
- 記分板展示 launcher shot: fill `#kiosk-text` / `#kiosk-footer` first.
- `kiosk-degraded` shot: the same URL and the same `text=` / `footer=` as
  `kiosk`, so the two read as one pair — then induce a stream failure and wait
  for the amber notice in the top-right corner:

  ```js
  await p.route('**/api/stream/public-contests/**', (route) =>
    route.fulfill({
      status: 200,
      headers: { 'content-type': 'text/event-stream', 'cache-control': 'no-cache' },
      // Give the clock a real value FIRST, then report the failure.
      body:
        ': keepalive\n\n' +
        'event: time\ndata: {"time_left":"167h53m55s","time_type":"submit"}\n\n' +
        'event: upstream-error\ndata: {"stream":"scoreboard","status":500}\n\n'
    })
  );
  // three consecutive failures back off 1s + 2s + 4s before the notice appears
  await p.getByRole('status').filter({ hasText: '即時連線異常' })
    .waitFor({ state: 'visible', timeout: 60_000 });
  ```

  Take the `time_left` from a live frame so it matches the staged contest:
  `curl -sN .../public/contest-projects/1/stream-time | head -1`.

  Two things this deliberately does NOT do, both learned by getting them wrong:

  - **Do not block the stream from the first connection.** The countdown then
    never arrives and the clock renders 已結束 — a picture that tells the
    operator the contest is over while the notice tells them to wait. A real
    outage leaves the clock frozen at its last value, which is why the fixture
    feeds one time frame before failing.
  - **Do not intercept the REST calls.** The scores on screen have to be the
    ones the system computed. Stage the *inputs* — seeded demo data, an induced
    failure — and let the app produce the output; faking the numbers makes the
    manual a drawing rather than a screenshot.

Screenshot list for the admin manual (keep these exact filenames; copy into
`frontend/static/reference/`): `status users groups images projects contests
contest-detail judge-workers clarifications submissions submission-detail
judgment rejudgings user-scores scoreboard-display kiosk kiosk-degraded
bundle-import`.

`bundle-import` is the import preview at `/dashboard/bundles/import` with a
bundle already chosen, so the 新建 / 已存在 / 會覆蓋 table is on screen — that
table is what manual section 3.6 tells the operator to read before importing.
Export one first (a contest detail page → 完整匯出) and feed it back in; do NOT
press 執行匯入.

## 4. Verify and finish

- **Look at the screenshots** (Read the png): a StateMessage/empty page usually
  means the picker wasn't selected or the window wasn't extended.
- Restore contest times (step 2), remove `.tmp.mjs` files, commit
  `frontend/static/reference/` together with any manual text changes.
