# 貢獻指南 / Contributing

歡迎回報問題與送出 PR。這份文件只寫**這個專案特有、猜不到的規則**;環境安裝與啟動請看 [README](README.md#快速上手)。

## 一條會讓 PR 被退回的規則:API 規格從 TypeSpec 開始

`typespec/` 是 API 的**單一事實來源**。三份 `openapi.yaml`(`typespec/tsp-output/`、`backend/api/`、`frontend/src/api/`)全都是產生物,而前後端的型別又是從它們產生的。

所以**手改任何一份 `openapi.yaml`、或手寫兩端的 DTO,都是不被接受的** —— 下一次有人重跑 build,你的修改就會消失,而且 `spec-check` workflow 會抓到不一致。

改 API 的正確流程:

```bash
# 1. 改規格
vim typespec/src/...

# 2. 產生
cd typespec && npm run build

# 3. 同步三份產生物
cp tsp-output/@typespec/openapi3/openapi.yaml ../backend/api/openapi.yaml
cp tsp-output/@typespec/openapi3/openapi.yaml ../frontend/src/api/openapi.yaml

# 4. 各自重新產生程式碼
cd ../backend  && make api              # oapi-codegen → api/api.gen.go(不入版控)
cd ../frontend && pnpm gen:api-client   # → src/api/gen(不入版控)
```

詳細說明見 [backend/README.md](backend/README.md#api-規格變更流程)。

**PR 請把 `.tsp` 與三份產生物一起提交。**只交其中一半的 PR 無法通過檢查,也無法讓別人在本機重現。

## 動了權限?有測試在守

`backend/api/openapi_security_test.go` 有三條不變式,新增端點時很容易踩到:

1. 每個 operation 都必須宣告 `security` —— 忘記寫不是「比較鬆」,是**完全沒有檢查**,任何人都能呼叫
2. 所有會改狀態的 method(POST/PUT/PATCH/DELETE)都必須要求某個 `Permission*`。唯一例外是登入/續期/登出這類 auth 端點:呼叫者此時還沒有(或正在取得)權限,無從要求 `Permission*`,所以它們改以「持有對應 cookie」把關——在 spec 上宣告 `RefreshToken` 或 `CsrfToken` scheme,測試也將這兩種視為已把關
3. spec 用到的 security scheme 集合,必須與 middleware 實際強制的集合**完全一致**

第 3 條的意思是:在 TypeSpec 新增一個 scheme,就必須同時在 `internal/app/identity_domain/auth/adapter/middleware/` 加對應的 middleware 並註冊到 `app/provider.Middlewares`,否則測試會紅。

真的要新增公開端點,把它加進那個檔案的 `intentionallyPublic` 允許清單 —— 讓「這是公開的」成為一個明確的決定,而不是漏寫。

## 測試

```bash
cd backend
make test:unit                # 單元測試
make test:feature             # godog BDD(testcontainers 起 mysql/mongo/redis)
make mocks                    # 改了介面之後重新產生 mockery mocks(mocks 不入版控)
```

```bash
cd frontend
pnpm run check                # svelte-check,以 0 errors 為基準
pnpm test:unit                # vitest
pnpm test:e2e --project=smoke # 對運行中的環境自建比賽並驗證主要動線
```

幾件會浪費你時間的事,先講:

- 單元測試**不需要任何準備**:它們讀入庫的測試 fixture(`internal/config/config/`)。
  根目錄的 `config_test.yaml` 只屬於 feature 測試(由其腳本產生,不入版控)。
- **feature 測試不能兩個 suite 同時跑** —— 它們會搶同一組 port。要縮小範圍請用 `TARGET_TEST=<feature 檔名>`。
- **`mocks/` 不入版控**,改了介面之後 build 會壞,跑 `make mocks`。
- **`api/api.gen.go` 也不入版控**,fresh clone 直接 `go build` 會缺它——先跑
  `make api`(產生器版本釘在 Makefile 裡,只需要 Go)。容器啟動與 CI 都會自動跑。
- feature 的 JSON 斷言裡,`"{任意字}"` 是字串萬用值,`1e+308` 是「任意正數」萬用值。

## 測試要驗到值,不只驗到欄位存在

這個專案吃過幾次虧,寫進來提醒:**斷言一個欄位是 `null`,只證明它存在,不證明它會被正確帶出來。**

新增一個會流經多層的欄位時(model → DTO → presenter → API),請至少寫一個**往返**測試:寫入一個實際的值,再從**每一個**會回傳它的端點讀回來。同一個欄位在某個端點正確、在另一個端點是 null,是這裡實際發生過的 bug。

## 分支模型

- 採 [GitHub flow](https://docs.github.com/en/get-started/using-github/github-flow) 加發佈 tag:只有一條長駐分支 `main`,從它切出、PR 回它;**沒有** develop 或 release 長駐分支,「版本」就是 `main` 上的 tag(`vX.Y.Z`)。
- 分支名建議 `類型/短描述`(如 `feat/staff-note`、`fix/oom-export`),與 commit 類型對齊即可,非強制。
- **不要在 PR 裡動 CHANGELOG 的版本段或打 tag**——你的 PR 只在 `[Unreleased]` 加條目(見下方),收版、標日期、打 tag、發 Release 由維護者觸發的 `release` workflow 完成。

## Commit 與 PR

- Commit 訊息用 [Conventional Commits](https://www.conventionalcommits.org/)(`feat:`、`fix:`、`refactor:`、`docs:`、`chore:`);破壞性變更加 `!`
- **訊息主體寫「為什麼」,不是「做了什麼」** —— diff 已經說明做了什麼。決定背後的取捨、你考慮過又否決的方案、以及未來的人可能會想改回去的理由,那些只有你知道
- 一個 PR 一個主題。混在一起的 PR 很難 review,也很難單獨 revert
- 使用者或部署者**感受得到**的變更(新功能、行為改變、UI/UX 調整、修正、部署與設定變動),請在 [CHANGELOG.md](CHANGELOG.md) 的 `[Unreleased]` 加一段;使用者無感的純內部重構不用

## 回報問題

- 一般 bug 與功能建議:開 GitHub Issue
- **安全問題請不要開公開 Issue** —— 見 [SECURITY.md](SECURITY.md)

## 授權

送出 PR 即表示同意你的貢獻以 [MIT](LICENSE) 授權釋出。
