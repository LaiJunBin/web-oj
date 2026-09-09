# Web Judge — Backend

Go 後端:clean architecture + [fx](https://github.com/uber-go/fx) DI,MySQL(關聯資料)、MongoDB(提交與評分)、Redis(refresh token),judge worker 以 goroutine 池透過 Docker 啟動評測容器。所有常用指令集中在 `Makefile`。

## 設定

```bash
cp config.example.yaml config.yaml
make generate:keys        # 互動式產生 appKey / jwtKey / csrfKey / storeKey 寫入 config.yaml
```

`config.yaml` 不入版控;每個欄位的完整說明都寫在 `config.example.yaml` 的註解裡。
重要欄位速覽與正式架設的注意事項見下方檢查表章節。

## 正式架設前的檢查表

`docker-compose.yml` 的預設**就是考場的架設形態**——「正式」與「開發」的差別全在
設定。重要欄位速覽,以及**設錯只會靜默失敗**的完整檢查表,見
[DEPLOYMENT.md](DEPLOYMENT.md),正式架設前請完整讀一遍。

## 啟動

**模式 A:Docker Compose(預設即正式形態)**

預設建置**編譯後的部署映像**——快速上手與考場用同一份檔案、同一條指令,
差別只在設定(見 [DEPLOYMENT.md](DEPLOYMENT.md) 檢查表):

```bash
docker compose up -d                      # app + mysql + mongo + redis(+ mongo-express)
docker compose exec app make migrate
docker compose exec app make seed
# API 在 http://localhost:8080
```

**更新程式後要 `docker compose up -d --build`**——預設映像在建置時烤入原始碼與
二進位,單純 `git pull` 不會改變運行中的版本(即時反映原始碼的是 dev overlay)。
**有疊 overlay 的環境,每次 compose 操作的 `-f` 清單都要帶齊**(監控環境用
`make up:monitor` 最保險):compose 會把容器收斂到「這次 `-f` 描述的定義」,
少帶 overlay 就等於用不含它的定義重建 app——監控環境的症狀是 Grafana 全面
斷更、任何地方都沒有錯誤訊息(metrics 綁回容器 loopback,Prometheus 抓不到)。

**模式 A-dev:熱重載開發**

疊上 dev overlay 換成 air + 原始碼掛載;**考場嚴禁使用**(理由見該檔檔頭):

```bash
make up:dev    # = docker compose -f docker-compose.yml -f docker-compose.dev.yml up -d
```

**模式 B:本機開發(sqlite)**

`config.yaml` 設 `rdbDriver: sqlite`,只起文件庫與快取:

```bash
docker compose up -d mongo redis
make api                                   # 產生 api/api.gen.go(不入版控,產生器已釘版)
make dev                                   # air 熱重載;或 go run cmd/main.go
```

sqlite 為 in-memory,每次啟動自動 migrate + seed,關聯資料歸零重來;**MongoDB 是持久的**,重啟後與關聯資料發散時,舊提交會掛在重複使用的 id 底下——需要全新環境時請一併清空 Mongo 的 `judge` 資料庫。

**模式 C:VS Code Dev Container**

[`.devcontainer/`](.devcontainer/) 疊在**模式 A-dev** 之上(devcontainer 已預設帶入 dev overlay):「Reopen in Container」後直接進入含 Go 工具鏈的 `app` 容器開發,judge 所需的 docker socket 已掛載,`make test:*` 皆可直接執行,主機不需安裝 Go。

自己的瀏覽器連 API 用 VS Code 自動轉發的 `localhost:8080` 即可(SSH 進 VM 開發也
一樣)。要讓 **host 或其他機器直連**(例如 `VM_IP:8080`):`cp .env.example .env`
後取消 `API_BIND_HOST` 的註解,再 Rebuild Container——預設只綁 127.0.0.1 的理由
見 [DEPLOYMENT.md](DEPLOYMENT.md)「API 對外埠」一列。

## 示範資料

`make seed` 寫入的內容全部定義於
[`database/seeder/seeders/demo_data.json`](database/seeder/seeders/demo_data.json)
(檔頭 `_readme` 說明設計原則與已知地雷;改示範文字只需改這個檔):

| 角色 | 帳號 | 密碼 |
|---|---|---|
| 管理員 | `admin` | `1234` |
| 示範選手 | `demo01` ~ `demo05` | `1234` |

seed 會建立一場**進行中**的示範競賽(activate 起算 14 天、比賽 7 天),所有示範帳號均已報名。示範密碼刻意公開,僅供初次安裝體驗,正式環境請自行建立帳號。

## 建置示範映像

seed 已登錄 `test-nginx:latest`(掛載路徑 `/usr/share/nginx/html`)但未建置。首次安裝時:

1. 將 [`../resources/demo/test-nginx-image/`](../resources/demo/test-nginx-image) 的**內容**(`Dockerfile` 與 `conf.d/`)壓成 zip。
2. 以 `admin` 登入前端 → 後台「評分環境管理」→ 進入 `test-nginx:latest` → 上傳並建置,等建置紀錄顯示完成。

之後用 `demo01` 提交 `tests/data/nginx.zip`,應得到 **18 / 25**(第 3 項「首頁已部署」0/5、第 4 項的 404 隱藏測資失敗——那份範例作品刻意只做一半,詳見 demo_data.json 的說明)。

## 選配:監控堆疊(實驗性)

[`docker-compose.monitor.yml`](docker-compose.monitor.yml) 疊在主 compose 之上:
Prometheus(定期抓取後端 `:9091` 的指標)、Grafana(已 provision judge 儀表板與資料源)、
Loki + Promtail(容器日誌)、cAdvisor(容器資源):

```bash
make up:monitor    # = docker compose -f docker-compose.yml -f docker-compose.monitor.yml up -d
# Grafana: http://localhost:3000(帳號 admin,密碼 $GRAFANA_ADMIN_PASSWORD,預設 admin)
```

之後這台機器的**每次**啟動與更新都用 `make up:monitor`,不要用 `make up` / 裸的
`docker compose up`——理由見「啟動」一節的 overlay 警告。overlay 會對 app 注入
`SERVER_METRICSHOST=0.0.0.0` 讓 Prometheus 能跨容器網路抓 `app:9091`(該埠不發布
到 host);遠端檢視 Grafana 請走 SSH 轉發(`ssh -L 3000:127.0.0.1:3000 …`),
不要把埠改綁 0.0.0.0。

定位是**競賽現場的即時觀測**(當場看 worker 健康度、資源與日誌),實際比賽用過;
但它不是 production 級監控——沒有告警、沒有長期保存策略、cAdvisor 以 privileged
執行、Promtail 掛 docker socket。所有介面埠只綁 127.0.0.1(理由同 API 的預設)。

## 測試

單元測試自足:讀的是入庫的測試 fixture(`internal/config/config/`),fresh clone
直接可跑。根目錄的 `config_test.yaml` 只屬於 feature 測試——由其腳本產生、寫入
本機實際值,不入版控。

```bash
make test:unit                # 單元測試
make test:feature             # godog BDD(testcontainers 起 mysql/mongo/redis,評測走 fake runtime)
make test:feature:short-ttl   # 短效期 token 自動續期情境(獨立跑,避免影響其他情境)
make mocks                    # 重新產生 mockery mocks
```

CI 分工:`backend-ci` 只跑單元與整合測試(快速回饋);feature 套件屬重量級,
由 `e2e` workflow 在 PR 與「有變更的夜間」執行。

## API 規格變更流程

規格的單一事實來源在 [`../typespec/`](../typespec):

```bash
cd ../typespec && npm run build          # 產出 tsp-output/@typespec/openapi3/openapi.yaml
cp tsp-output/@typespec/openapi3/openapi.yaml ../backend/api/openapi.yaml
cd ../backend && make api             # oapi-codegen 重新產生 api/api.gen.go
```

`api/api.gen.go` **不入版控**:`make api` 以 `go run` 釘版執行 oapi-codegen,
模式 A 於 image 建置時產生、A-dev/C 於容器啟動時產生,CI 也會自動產生;只有
模式 B 的 fresh clone 需要手動跑一次(見上方指令)。

(前端同步:把同一份 openapi.yaml 複製到 `frontend/src/api/openapi.yaml` 後執行 `pnpm gen:api-client`。)

## 已知限制

已知限制記錄於 [KNOWN_LIMITATIONS.md](../KNOWN_LIMITATIONS.md)。
