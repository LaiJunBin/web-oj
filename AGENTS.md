# 給 AI 協作工具(與趕時間的人類)的導覽

這份文件是路由,不是規則本體——規則的完整版都在連結裡,**以連結目標為準**。

## 動手前先讀

1. [CONTRIBUTING.md](CONTRIBUTING.md) — 本專案特有、猜不到的規則,**全文必讀**(不長)
2. [README.md](README.md) — 系統架構、快速上手、workflow 一覽
3. 要改 API:[typespec/README.md](typespec/README.md) 與 [backend/README.md](backend/README.md#api-規格變更流程)
4. 要碰部署或設定:[backend/DEPLOYMENT.md](backend/DEPLOYMENT.md)
5. [KNOWN_LIMITATIONS.md](KNOWN_LIMITATIONS.md) — 已知的既有行為

## 最常被自動化工具踩爆的四件事

- **API 規格變更(端點、輸入輸出、模型)一律從 `typespec/` 改起**;純實作變更不經 typespec。三份 `openapi.yaml` 與兩端型別都是產物;手改任何一份、或手寫兩端 DTO,`spec-check` 會退你的 PR。正確流程見 CONTRIBUTING。
- **這些產物不入版控,缺了要用指令產生、不要 commit**:`backend/api/api.gen.go`(`make api`)、後端散佈於 `internal/**` 各層的 `mocks/`(`make mocks`)、`frontend/src/api/gen/`(`pnpm gen:api-client`)。fresh clone 直接 build 會缺件是正常的。
- **`config.yaml` 與金鑰不入版控**;`config.example.yaml` 才是要維護的檔案(含註解)。
- **feature 測試(godog)不能兩個 suite 並行**(搶 port);縮小範圍用 `TARGET_TEST=<檔名>`。前端以 `pnpm run check` 0 errors 為基準。

## 改動的完成定義

- Commit 用 Conventional Commits,**主體寫「為什麼」**(取捨、否決過的方案),不是重述 diff。
- 使用者可見的變更在 [CHANGELOG.md](CHANGELOG.md) 的 `[Unreleased]` 加一段。
- 新欄位流經多層(model → DTO → presenter → API)時,寫**往返**測試:實值寫入、從每個會回傳它的端點讀回(見 CONTRIBUTING「測試要驗到值」)。
- 文件是交付物的一部分:改了行為,同一輪就掃描 README/兩份操作手冊有沒有因此過時。
- 動到 UI 畫面而手冊截圖過時時,[.claude/skills/](.claude/skills/) 有重製截圖的操作手冊。
