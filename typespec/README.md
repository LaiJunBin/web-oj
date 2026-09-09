# Web Judge — API 規格(TypeSpec)

整個系統的 API **單一事實來源**。前後端的 API 程式碼都由這裡產生,任何規格變更(端點、欄位、型別)一律從這裡開始——**不要**直接改後端的 `api/openapi.yaml`、`api.gen.go` 或前端的 `src/api/gen/`,它們都是產物。

## 指令

```bash
npm ci
npm run build     # 編譯 → tsp-output/@typespec/openapi3/openapi.yaml
npm run dev    # watch 模式,改檔即重編
```

`.env` 提供服務名稱與 server URL 等模板變數(無機密,已入版控)。

只是想**瀏覽 API** 的話不必建置:產出的
[`tsp-output/@typespec/openapi3/openapi.yaml`](tsp-output/@typespec/openapi3/openapi.yaml)
已入版控,用任何 OpenAPI 檢視工具開啟即可(backend 的 dev container 已內建 VS Code
的 Swagger Viewer 擴充)。

## 結構與慣例

| 路徑 | 內容 |
|---|---|
| `src/specs/admin/<module>/` | 管理端 API,依模組拆分,慣例為 `index.tsp`(端點)+ `model.tsp`(模型) |
| `src/specs/user/`、`src/specs/common/` | 選手端與公開/共用 API |
| `src/shared/` | 跨區共用:授權 scope、共用回應(如 401/404 包裝)、共用模型與列舉 |
| `main.tsp` | **自動生成,勿手改**——`npm run build` 會掃描 `src/` 重新產生 import 清單 |
| `build-utils/` | 上述的建置腳本(main.tsp 生成、環境變數注入、watch) |

新增模組:在 `src/specs/<區>/<模組>/` 建 `index.tsp` 與 `model.tsp` 即可,build 會自動納入,不需手動登記。

## 產出如何流向前後端

```bash
npm run build

# 後端
cp tsp-output/@typespec/openapi3/openapi.yaml ../backend/api/openapi.yaml
cd ../backend && make api          # oapi-codegen → api/api.gen.go(不入庫)

# 前端
cp tsp-output/@typespec/openapi3/openapi.yaml ../frontend/src/api/openapi.yaml
cd ../frontend && pnpm gen:api-client   # openapi-zod-client → src/api/gen/(不入庫)
```

兩端的 `openapi.yaml` 都要入版控(它是產物快照,讓 clone 後不必先裝 typespec 也能重生 client);兩端的**程式碼**產物(後端 `api/api.gen.go`、前端 `src/api/gen/`)則於建置/啟動時重生,不入庫。
