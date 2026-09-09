# 正式架設前的檢查表

`docker-compose.yml` 的預設**就是考場的架設形態**(編譯後的部署映像;實際比賽
以同一堆疊完賽)——「正式」與「開發」的差別全在設定,下表就是那份差異清單。
熱重載開發是 opt-in 的 overlay(見 [README](README.md#啟動)),**考場請勿疊上**。情境若超出
「可信單機、單場比賽」(長期運營、面向不可信網路),請從 compose 檔頭的部署
指引出發,那超出本文件的承諾範圍。

多數欄位設錯只會拿到預設值,但下面這幾項設錯的症狀是**靜默失敗**——服務起得來,某個功能就是不動,而且訊息不會指向設定檔。

| 項目 | 要做什麼 | 設錯會怎樣 |
|---|---|---|
| `env` | 純 HTTP 的區網考場用 **`release`**,只有真的有 TLS 才用 `production` | `production` 會強制 `auth.cookie.secure=true`,瀏覽器於是不送出 auth 與 CSRF cookie:登入靜默失敗、CSRF 對不上,線索只有一行 log |
| `server.rdbDriver` | 正式一律 `mysql` | `sqlite` 是 in-memory 示範資料庫,每次重啟清空並重新植入公開的示範帳號——在 `production` 下會直接 panic 擋下 |
| API 對外埠 | 考場要讓選手瀏覽器連得到,啟動時帶 `API_BIND_HOST=0.0.0.0` | `docker-compose.yml` 預設綁 `127.0.0.1`,同機以外一律連不上 |
| API 實例數 | **只能跑一個 API 行程**(不要為了負載或高可用開第二份、也不要 `--scale`) | 修復「卡住的評測」的背景掃描只看得到**自己行程**的 worker,所以第二個實例正在評測的提交會被它當成孤兒、直接寫成 ERROR。沒有任何錯誤訊息,只有選手的成績在比賽當下莫名變成錯誤 |
| 比賽期間重啟後端 | 避免;需要更新時挑沒有評測在跑的空檔。啟動用**預設 compose**(有跑監控堆疊則 `make up:monitor`,`-f` 清單要帶齊,見 README「啟動」)、勿疊 dev overlay——air 的檔案監看會把誤觸原始碼變成自動重啟 | 提交在回應成功後的**數秒內**才完成落地;後端恰好在這個窗口被強制中斷(當機、直接 kill),該筆提交會遺失且不留任何痕跡,選手需重新提交。正常運行不會發生、窗口極短,機率非常低——但重啟是唯一會人為製造這個窗口的操作。進行中的重評(rejudge)也可能因此卡在「進行中」,屆時取消後重新發起即可 |
| `server.trustedProxies` | **只有**真的架在反向代理後面才填該代理位址 | 留空時登入頻率限制以真實來源計算(正確);亂填會讓任何人偽造 `X-Forwarded-For` 繞過限制 |
| `server.loginRateLimit*`、`loginRateLimitExemptCIDRs` | 失敗才**結算**扣額度(成功登入結算後退還),所以一間教室同時登入不會互鎖;但逐 IP 的 backstop(`loginRateLimitIpBurst`,預設 300)的 burst 要照**最大單一來源的同時登入人數**放大——每個在途登入(**含密碼正確者**)進來時都會先佔一枚 token 到結束才退,再加上手誤;真正信任的考場網段可直接填進 `loginRateLimitExemptCIDRs` 豁免 | backstop 太小時,開賽瞬間的同時登入(**即使密碼全對**)加手誤會把該 IP 的 backstop 佔滿,連密碼打對的人也一起被 429(房間層級的互鎖) |
| `judge.isolateProjectNetwork` | 保持預設 `true` | 關掉等於讓任何人上傳的程式碼能連外網與主機;開著則出題方要確保映像不在啟動時才下載東西(見管理端手冊 3.2) |
| `judge.memoryLimitMB`、`judge.cpuLimit` | 正式比賽前**依主機規格設定這兩項**——兩者**預設不限制**(`config.example.yaml` 裡就是 `0`),取捨與建議值見該檔註解;`pidsLimit`(1024)與 `maxOutputBytesPerExec`(10 MiB)預設已開,不必動 | 不設就是「評測容器共用主機、但沒有 CPU 與記憶體上限」:一份失控的提交可以吃滿整台機器,拖慢**其他選手同時進行**的評測;記憶體被吃光時,核心會在全主機範圍挑 OOM 對象,可能殺掉的是另一場評測、MySQL 或後端本身 |
| `docker.mode: remote` | 必須一併給 `tlsCACert` / `tlsCert` / `tlsKey`;只有遠端 daemon 僅暴露於可信隔離內網(如 lab 的 192.168)才可改用 `allowInsecureRemote: true` 走純 TCP | 沒給憑證又沒開 `allowInsecureRemote` 會拒絕啟動(docker API 等同 root,預設不接受無 mTLS 的遠端連線);開了 `allowInsecureRemote` 而該 port 被不可信主機碰得到,等於把 judge host 的 root 交出去 |
| 資料庫密碼 | 改掉 `config.example.yaml` 與 compose 帶的預設密碼 | 預設密碼是公開的 |
| 示範資料 | 不要在正式環境跑 `make seed` | `admin` / `demo01`~`demo05` 密碼皆為 `1234`,且公開記載於 README |

評測會在主機上實際執行選手提交的專案,請用**專用主機或 VM**,不要與其他服務共用;`docker-compose.yml` 也把 docker socket 掛進 API 容器,那等同該主機的 root 權限,別讓它直接面對不信任的網路。

## 重要欄位速覽

- `server.rdbDriver`:`sqlite`(開發預設,in-memory,**每次啟動自動 migrate + seed**)或 `mysql`。
- `server.allowedOrigins`:前端 origin(預設含 `http://localhost:5173`)。
- `auth.*TTLSeconds`、`auth.cookie.*`:token 效期與 cookie 屬性。`env: production` 會**自動強制** `cookie.secure=true`(`sameSite: None` 亦然),所以純 HTTP 區網要用 `env: release`(見上方檢查表);只有「非 production 但走 HTTPS」才需要手動把 `secure` 設為 `true`。
- judge 需要存取 Docker daemon(compose 已掛 `/var/run/docker.sock`)。

每個欄位的完整說明都寫在 `config.example.yaml` 的註解裡,值得在正式架設前讀一遍。
