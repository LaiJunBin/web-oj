<script lang="ts">
	import { Accordion, AccordionItem, Alert, Badge } from 'flowbite-svelte';
	import { InfoCircleSolid } from 'flowbite-svelte-icons';

	// 章節目錄:id 對應各 <section> 的 id,點擊平滑捲動至該節
	const sections = [
		{ id: 'overview', title: '1. 後台總覽' },
		{ id: 'status', title: '2. 系統狀態(比賽指揮台)' },
		{ id: 'pre', title: '3. 賽前設定' },
		{ id: 'process', title: '4. 競賽過程操作' },
		{ id: 'kiosk', title: '5. 記分板與大螢幕(kiosk)' },
		{ id: 'glossary', title: '6. 詞彙表' },
		{ id: 'faq', title: '7. 疑難排解' }
	];
</script>

{#snippet screenshot(src: string, caption: string)}
	<figure class="my-4">
		<img
			{src}
			alt={caption}
			loading="lazy"
			class="w-full rounded-lg border border-gray-200 shadow-sm dark:border-gray-700"
		/>
		<figcaption class="mt-2 text-center text-sm text-gray-500 dark:text-gray-400">
			{caption}
		</figcaption>
	</figure>
{/snippet}

{#snippet step(n: number, text: string)}
	<li class="flex gap-3">
		<span
			class="flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full bg-blue-600 text-sm font-bold text-white"
			>{n}</span
		>
		<span class="pt-0.5">{@html text}</span>
	</li>
{/snippet}

<div class="mx-auto max-w-4xl px-4 py-8 text-gray-800 dark:text-gray-100">
	<h1 class="text-center text-4xl font-extrabold tracking-tight text-gray-900 dark:text-white">
		後台操作手冊
	</h1>
	<p class="mt-3 text-center text-gray-500 dark:text-gray-400">
		寫給主辦與監考人員:從賽前設定、比賽當天監控,到賽後成績匯出的完整流程。
	</p>

	<Alert color="blue" class="mt-6 dark:bg-blue-900 dark:text-blue-100">
		<InfoCircleSolid slot="icon" class="h-5 w-5" />
		<span class="font-semibold">比賽當天三件事:</span>
		① 開啟<strong>「系統狀態」</strong>確認評分工作者運作中 → ② 從側欄<strong
			>「競賽過程」</strong
		>選定進行中的比賽專案 → ③ 全程留意<strong>評分佇列</strong>與<strong>最近的評分錯誤</strong>。
	</Alert>

	<!-- 目錄 -->
	<nav class="mt-8 rounded-lg border border-gray-200 p-4 dark:border-gray-700">
		<p class="mb-2 font-bold">目錄</p>
		<ul class="grid grid-cols-1 gap-1 sm:grid-cols-2">
			{#each sections as s (s.id)}
				<li>
					<a href={`#${s.id}`} class="text-blue-600 hover:underline dark:text-blue-400">
						{s.title}
					</a>
				</li>
			{/each}
		</ul>
	</nav>

	<!-- 1. 後台總覽 -->
	<section id="overview" class="mt-10 scroll-mt-4">
		<h2 class="border-b-2 border-gray-200 pb-2 text-2xl font-bold dark:border-gray-700">
			1. 後台總覽
		</h2>
		<p class="mt-4">
			使用具有<strong>管理員角色</strong>的帳號登入後,點選導覽列的「後台管理」即可進入後台。左側側欄分為四區:
		</p>
		<ul class="mt-3 list-inside list-disc space-y-1">
			<li><strong>系統狀態</strong>:比賽期間的監控儀表板(見第 2 章)。</li>
			<li>
				<strong>賽前設定</strong>:使用者、群組、評分環境、專案、競賽、評分工作者的管理(見第 3
				章)。
			</li>
			<li>
				<strong>競賽過程</strong>:先在此區選定「進行中的比賽專案」,底下的公告/提交/重評/分數頁面都以它為對象(見第
				4 章)。
			</li>
			<li><strong>參考文件</strong>:本手冊。</li>
		</ul>
	</section>

	<!-- 2. 系統狀態 -->
	<section id="status" class="mt-10 scroll-mt-4">
		<h2 class="border-b-2 border-gray-200 pb-2 text-2xl font-bold dark:border-gray-700">
			2. 系統狀態(比賽指揮台)
		</h2>
		<p class="mt-4">
			這一頁是比賽當天的主畫面,所有數字皆<strong>即時更新</strong>,建議全程開著:
		</p>
		{@render screenshot('/reference/status.png', '系統狀態頁:工作者、佇列、開放中的比賽、評分錯誤')}
		<ul class="mt-3 list-inside list-disc space-y-2">
			<li>
				<strong>評分工作者</strong>:每個工作者的狀態(<Badge color="green">IDLE / RUNNING</Badge>
				為正常)。若全部停擺或出現
				<Badge color="red">ERROR</Badge>,頁面頂部會出現紅色警示——這代表提交將無法評分,請立即前往「評分工作者管理」處理。
			</li>
			<li>
				<strong>評分佇列</strong>:目前排隊中的評分數(分完整評分/部分評分/重評)與執行中數量。「<strong
					>最久等待</strong
				>」指等待被工作者<strong>接手</strong>的時間(不含評分執行本身)——單筆評分可能要跑數分鐘,提交數多於工作者時排隊數分鐘屬正常,<strong
					>等待長不代表異常</strong
				>。真正要警覺的是紅色警示「<strong>有排隊但沒有任何評分在執行中</strong>」(持續約 15
				秒以上才會顯示):這代表派工卡住了,請檢查評分工作者。
			</li>
			<li><strong>近 10 分鐘提交</strong>:提交量的迷你長條圖,可預見高峰。</li>
			<li>
				<strong>最近的評分錯誤</strong>:最新的評分失敗清單,<strong>點擊任一筆可直達該筆評分的詳細頁</strong>查看錯誤原因。
			</li>
			<li>
				<strong>評分管線</strong>(下方):選定專案的提交/評分總數、平均評分時長與平均排隊等待。
			</li>
		</ul>
	</section>

	<!-- 3. 賽前設定 -->
	<section id="pre" class="mt-10 scroll-mt-4">
		<h2 class="border-b-2 border-gray-200 pb-2 text-2xl font-bold dark:border-gray-700">
			3. 賽前設定
		</h2>
		<p class="mt-4">建議依以下順序準備一場比賽:</p>

		<h3 class="mt-6 text-xl font-bold">3.1 建立選手帳號與群組</h3>
		<ol class="mt-3 space-y-3">
			{@render step(1, '進入「<strong>使用者管理</strong>」,以「新增」逐一建立帳號,或使用批次匯入。選手帳號需具備 <strong>contestant</strong> 角色。')}
			{@render step(2, '若選手眾多,先到「<strong>群組管理</strong>」建立群組並將選手加入——之後報名比賽時可整組加入,不必逐人勾選。')}
		</ol>
		<div
			class="mt-3 rounded-lg border border-gray-200 bg-gray-50 p-4 text-sm dark:border-gray-700 dark:bg-gray-800"
		>
			<p class="font-bold">角色與帳號維護</p>
			<ul class="mt-2 list-disc space-y-1 ps-5">
				<li>
					共三種角色:<strong>admin</strong>(後台所有操作)、<strong>contestant</strong>(參賽提交)、<strong>api_reader</strong>(唯讀
					API 整合用)。同一個帳號可同時具備多個角色。
				</li>
				<li>
					<strong>權限異動即時生效</strong>:調整某個帳號的角色後,對方的下一個操作就套用新權限,不需要請他重新登入(對方畫面上的選單則在他下次重新整理後跟上)。
				</li>
				<li>
					<strong>重設密碼會讓該帳號在所有裝置上登出</strong>。賽中替選手處理忘記密碼時要留意:對方正在作答的視窗會在一小時內失效,需要用新密碼重新登入(已提交的內容不受影響)。
				</li>
				<li>
					<strong>登入有頻率限制,但只有「失敗」會扣額度</strong>——密碼正確的登入完全不消耗,所以開賽時整間教室同時登入不會互相擋。額度以<strong
						>(來源 IP, 帳號)</strong
					>計算,同一條線路底下每位選手各自獨立。另有一個較寬鬆的<strong>逐 IP</strong
					>上限,用來擋單一來源亂試不同帳號;一整間教室的<strong>手誤</strong>會共用它,人多時請架設者把
					<code>server.loginRateLimitIpBurst</code>(預設 300)照最大單一來源的人數放大,信任的考場網段也可以直接填進
					<code>server.loginRateLimitExemptCIDRs</code> 豁免。選手若收到「嘗試過於頻繁」,那不是帳密錯誤,稍候即會恢復。
				</li>
			</ul>
		</div>
		{@render screenshot('/reference/users.png', '使用者管理:建立與維護帳號')}
		{@render screenshot('/reference/groups.png', '群組管理:批次管理選手')}

		<h3 class="mt-6 text-xl font-bold">3.2 準備評分環境</h3>
		<p class="mt-3">
			「<strong>評分環境管理</strong>」維護評分用的容器映像(image)。新增或修改後系統會自動建置,可即時查看建置紀錄(build
			log);建置失敗會標示錯誤。
		</p>
		{@render screenshot('/reference/images.png', '評分環境管理:映像清單與建置狀態')}

		<div
			class="mt-3 rounded-lg border border-gray-200 bg-gray-50 p-4 text-sm dark:border-gray-700 dark:bg-gray-800"
		>
			<p class="font-bold">評測時選手容器沒有對外網路——出題與做映像都要照這個前提</p>
			<p class="mt-2">
				評測會起兩個容器:跑選手作品的<strong>選手容器</strong>,以及跑評分指令的<strong
					>測試容器</strong
				>。選手容器預設<strong>不接外網</strong>——連不到網際網路、主機、也連不到網路上其他服務,只看得到同一場評測的測試容器。理由很直接:那裡面跑的是任何人都能上傳的程式碼。
			</p>
			<ul class="mt-2 list-disc space-y-1 ps-5">
				<li>
					<strong>映像要用的東西必須在建置期就備妥</strong>。在 entrypoint 或 CMD 裡才 <code
						>apt-get</code
					>、<code>npm install</code>、拉遠端設定檔的映像,會卡在健康檢查而讓整題評測失敗。
				</li>
				<li>
					<strong>測試容器不受限制</strong>,仍可連 SonarQube 等外部服務。需要外部資源的評分邏輯請放在測試容器這一側。
				</li>
				<li>
					選手作品若在<strong>執行期</strong>呼叫外部 API 或 CDN 也會失敗。這一點選手手冊第 8
					章已經寫明,出題敘述不必重複,但設計題目時要把它算進去。
				</li>
			</ul>
			<p class="mt-2">
				這是部署層的設定(<code>judge.isolateProjectNetwork</code>,預設開啟),不是逐題選項。真的有題目非得讓選手容器連網不可,要由架設者在
				<code>config.yaml</code> 關閉它——那會對<strong>所有</strong>題目生效,請衡量後再做。
			</p>
		</div>

		<div
			class="mt-3 rounded-lg border border-gray-200 bg-gray-50 p-4 text-sm dark:border-gray-700 dark:bg-gray-800"
		>
			<p class="font-bold">映像要讓平台知道「服務何時就緒」——請務必定義 HEALTHCHECK</p>
			<p class="mt-2">
				評測開始前,平台只做一件事確認選手容器準備好了:等映像的 Docker
				<code>HEALTHCHECK</code> 回報 healthy(最多等
				<code>healthcheckTimeoutSeconds</code>)。<strong>映像若沒有定義
					HEALTHCHECK,平台不會等</strong
				>——測試會在容器一啟動就開跑,而「容器啟動」不等於「服務在聽」。在剛開機、映像尚未暖起來的主機上,這個時間差足以讓第一個測試連不上,結果被記成該選手<strong
					>答錯、0 分</strong
				>,而不是系統錯誤。
			</p>
			<ul class="mt-2 list-disc space-y-1 ps-5">
				<li>
					<strong>選手映像的 Dockerfile 要定義 <code>HEALTHCHECK</code></strong
					>,以服務真的能回應為準(例如對首頁的 HTTP 檢查)。
				</li>
				<li>
					<strong>測試指令不要自帶短的連線逾時</strong>(例如 <code>curl --max-time 5</code
					>),平台的 testcase 逾時才是上限;需要先確認服務可用的話,把等待放進
					<strong><code>before_all</code> 生命週期指令</strong>,而不是讓第一個測試連不上就把選手判
					0 分(前置步驟失敗會怎樣、哪些步驟容忍非零 exit,見 <strong>3.3</strong> 的完整規則)。
				</li>
				<li>賽前開機後先跑一次示範提交把主機暖起來,再開放比賽。</li>
			</ul>
		</div>

		<h3 class="mt-6 text-xl font-bold">3.3 建立專案(題目)</h3>
		<p class="mt-3">
			「<strong>專案管理</strong>」定義題目本體:描述、<strong>評分項目</strong>與底下的<strong
				>測試案例</strong
			>(指令、配分、比對方式)、執行的生命週期指令,以及提供給選手下載的素材。
		</p>
		{@render screenshot('/reference/projects.png', '專案管理:題目與評分項目')}

		<div
			class="mt-3 rounded-lg border border-amber-300 bg-amber-50 p-4 text-sm dark:border-amber-700 dark:bg-amber-950"
		>
			<p class="font-bold">生命週期指令的出題規則(fail-fast,先看再出題免得踩雷):</p>
			<ul class="mt-2 list-disc space-y-1 ps-5">
				<li>
					<strong>before_all / before_each 任一指令以非零 exit 結束,整筆評測立即中止並判為錯誤</strong>——選手會看到「評測前置步驟失敗,請聯絡現場工作人員」,看不到你的指令或輸出。前置步驟只放「非成功不可」的準備。
				</li>
				<li>
					<strong>after_all / after_each 容忍非零 exit</strong>:清理、關服務、刪暫存這類「失敗了也無所謂」的指令請放後置步驟,或在前置步驟自行 <code>|| true</code> 兜住。
				</li>
				<li>
					<strong>迴圈的 exit code 是最後一次迭代的結果</strong>:<code>for f in *; do grep ... "$f"; done</code> 只有最後一圈的成敗算數,中間的失敗會被吃掉——需要逐項把關就改用會累積失敗的寫法。
				</li>
				<li>
					<strong>引用可能不存在的檔案先防護</strong>:直接 <code>cat 設定檔</code> 會因檔案不存在而非零中止;用 <code>[ -f 檔案 ] && ...</code> 或先確認存在。
				</li>
			</ul>
			<p class="mt-2">
				<strong>排障入口</strong>:評分工作者的 log 對每一條生命週期都印出 Exit code(非零時加印 stderr 結尾);而在後台,<strong>評分詳情頁頂部會直接列出失敗的步驟</strong>(哪一步、exit 幾、stderr 第一行),不必翻原始資料。
			</p>
		</div>

		<h3 class="mt-6 text-xl font-bold">3.4 建立競賽並掛上專案</h3>
		<ol class="mt-3 space-y-3">
			{@render step(1, '進入「<strong>競賽管理</strong>」→「新增」,填寫名稱與時間。')}
			{@render step(2, '將 3.3 建立的專案掛入比賽,並為每個專案設定四個時間(意義見第 6 章詞彙表):<strong>上線(activate)→ 開賽(start)→ 結束(end)→ 下線(deactivate)</strong>。')}
			{@render step(3, '在比賽詳情頁把<strong>選手或群組</strong>加入參賽者——<strong>沒有報名的帳號無法提交</strong>,這是最常見的賽前疏漏。')}
			{@render step(4, '視需要設定記分板選項:精簡樣式(評分項很多時)與姓名隱碼。')}
			{@render step(5, '視需要設定<strong>提交冷卻(秒)</strong>:同一位選手對同一題兩次提交之間的最短間隔。<strong>留空</strong>沿用系統預設,<strong>0</strong> 為明確不限制。')}
			{@render step(6, '視需要調整<strong>評測資源上限</strong>(時間、行程數、記憶體):這是評測容器的上限,語意與提交冷卻相同——留空沿用系統預設,<strong>0</strong> 代表不限制。多數題目不需要改。')}
		</ol>
		<div
			class="mt-3 rounded-lg border border-gray-200 bg-gray-50 p-4 text-sm dark:border-gray-700 dark:bg-gray-800"
		>
			<p class="font-bold">提交冷卻的幾個行為,先講清楚免得賽中被質疑:</p>
			<ul class="mt-2 list-disc space-y-1 ps-5">
				<li>
					間隔從<strong>上一次提交送出的時間</strong>起算,不是從評測完成起算——否則評測較慢的題目,實際等待會遠長於設定值。
				</li>
				<li>
					<strong>取消提交不會重置間隔</strong>。取消是選手自己的操作,若能重置,提交→取消→提交就繞過了限制。
				</li>
				<li><strong>管理員不豁免</strong>:提交是選手行為,與角色無關。</li>
				<li>間隔<strong>逐競賽專案獨立</strong>:某一題在冷卻中,不影響其他題的提交。</li>
				<li>
					<strong>編輯競賽時清空欄位等於維持現值</strong>,不會變回沿用系統預設。要改回跟隨系統預設,請直接填入與系統預設相同的數字——實際數字就寫在表單欄位的
					placeholder 與下方說明(評測資源上限的預設也顯示在同一處)。
				</li>
			</ul>
		</div>

		<div
			class="mt-3 rounded-lg border border-gray-200 bg-gray-50 p-4 text-sm dark:border-gray-700 dark:bg-gray-800"
		>
			<p class="font-bold">評測資源上限:出題前值得知道的三件事</p>
			<ul class="mt-2 list-disc space-y-1 ps-5">
				<li>
					<strong>行程數算的是 task,包含執行緒</strong>,不是只算行程。chromium 只有少數幾個行程卻有數百條執行緒,sonar-scanner
					的 JVM 一個行程就有數十條——照行程數估會估得太低,把正常的題目擋掉。
				</li>
				<li>
					行程數<strong>預設是開啟的</strong>(系統預設 1024)。它擋的是評測時間上限擋不住的東西:fork
					bomb 會在幾秒內塞滿主機的行程表,遠早於逾時計時器生效。
				</li>
				<li>
					記憶體<strong>預設不限制</strong>。sonar-scanner 的 JVM 或 chromium 用掉 1～2 GB
					都算正常,貿然開啟會讓「重但正常」的題目突然失敗;要不要開,是看過自己最吃資源的那一題之後再決定。
				</li>
			</ul>
			<p class="mt-2">
				<strong>排查</strong>:某一題出現 fork 或資源相關的錯誤時,先看這一題的行程數上限——尤其是會跑
				chromium、playwright 或 sonar-scanner 的題目。
			</p>
		</div>
		{@render screenshot('/reference/contests.png', '競賽管理:比賽清單')}
		{@render screenshot('/reference/contest-detail.png', '比賽詳情:專案時間、參賽者與記分板選項')}

		<h3 class="mt-6 text-xl font-bold">3.5 啟動評分工作者</h3>
		<p class="mt-3">
			「<strong>評分工作者管理</strong>」可<strong>動態建立與刪除</strong>工作者:依比賽規模決定數量(人多題多就多開幾個),用「新增」建立、「刪除」移除,點入可看即時執行日誌。賽前務必確認至少一個工作者為運作狀態,並以一筆測試提交驗證評分管線可用。
		</p>
		{@render screenshot('/reference/judge-workers.png', '評分工作者管理:動態增減與即時日誌')}

		<h3 class="mt-6 text-xl font-bold">3.6 把整場競賽搬到另一台機器</h3>
		<p class="mt-3">
			題目、映像與競賽都可以匯出成 <strong>.zip 檔</strong>帶著走——在各自的詳情頁按「<strong
				>匯出</strong
			>」,在對應的列表頁按「<strong>匯入</strong>」。要一次搬整場,用競賽詳情頁的「<strong
				>完整匯出</strong
			>」:它把競賽、底下所有題目、以及這些題目用到的映像收進同一個檔案,對應側欄的「<strong
				>匯入完整競賽</strong
			>」。那一頁<strong>只接受「完整匯出」產生的檔案</strong>;單一題目或映像請走各自的列表頁。
		</p>
		<ol class="mt-3 space-y-3">
			{@render step(1, '在來源機器的競賽詳情頁按「<strong>完整匯出</strong>」,選映像要以哪種形態隨行(見下方比較)。')}
			{@render step(2, '到目標機器的側欄「<strong>匯入完整競賽</strong>」上傳檔案,先按「<strong>檢視內容</strong>」。')}
			{@render step(3, '看預覽:每個映像會標示<strong>新建 / 已存在(相同) / 會覆蓋</strong>,會覆蓋的還會列出受影響的題目。確認沒問題再按「<strong>執行匯入</strong>」。')}
			{@render step(4, '若匯入的是<strong>建置原始檔</strong>,匯入後映像會排入建置——建置完成前那些題目無法評測,報告會給你到評分環境管理的連結。')}
		</ol>
		{@render screenshot('/reference/bundle-import.png', '匯入完整競賽:執行前先看預覽——每個映像標示新建/已存在/會覆蓋')}

		<div
			class="mt-3 rounded-lg border border-gray-200 bg-gray-50 p-4 text-sm dark:border-gray-700 dark:bg-gray-800"
		>
			<p class="font-bold">兩種匯出形態怎麼選</p>
			<ul class="mt-2 list-disc space-y-1 ps-5">
				<li>
					<strong>建置原始檔</strong>:檔案小(KB 級),但匯入端要自己建置,過程需要能連外拉 base
					image。平常搬題目用這個。
				</li>
				<li>
					<strong>映像檔</strong>:檔案大(可能數百 MB),匯入端免建置也免連網——<strong
						>離線考場要用這個</strong
					>。它要求來源機器上所有映像都已建置完成,否則匯出會被擋下並列出還沒建置的
					tag。
				</li>
			</ul>
			<p class="mt-2">
				<strong>映像檔形態的完整匯出很吃記憶體。</strong>整個檔案是在伺服器記憶體中組裝的,高峰可達所有映像與素材總量的<strong>數倍</strong>;記憶體不足時匯出會失敗——而且是後端行程整個被系統終止(畫面上只看得到「匯出完整競賽失敗」,主機
				log 只有一行 <code>Killed</code>),所以<strong>不要在評測進行中做完整匯出</strong>。大競賽請優先選「建置原始檔」形態;真的需要映像檔隨行時,可改為<strong>分件匯出</strong>(單題、單映像在各自頁面匯出/匯入,產生的內容與完整匯出裡的部件相同),或確認主機的可用記憶體足夠再匯。
			</p>
			<p class="mt-2">
				<strong>參賽者不會一起匯入。</strong>報名名單屬於個人資料,不會被帶走——匯入之後請到新競賽自行把選手或群組加進參賽者,否則他們無法提交。這是最容易漏掉的一步。
			</p>
			<p class="mt-2">
				預覽如果出現<strong>無法匯入</strong>的紅框,代表這個檔案在這台機器上做不到(例如題目參照了檔案裡沒有、本機也沒有的映像)。執行的按鈕會停用,要先解決列出的問題。
			</p>
		</div>
	</section>

	<!-- 4. 競賽過程 -->
	<section id="process" class="mt-10 scroll-mt-4">
		<h2 class="border-b-2 border-gray-200 pb-2 text-2xl font-bold dark:border-gray-700">
			4. 競賽過程操作
		</h2>

		<h3 class="mt-6 text-xl font-bold">4.1 選定進行中的比賽專案</h3>
		<p class="mt-3">
			側欄「競賽過程」頂部有兩個下拉選單(比賽/專案),<strong
				>只會列出目前在上線期間內的比賽</strong
			>(歷史比賽請走「競賽管理」)。系統會自動帶入預設:優先跟隨導覽列目前所在的比賽(若它正在進行),否則選第一場進行中的比賽;專案則在只有一個時自動選定。選定後,內容區頂部會出現<strong
				>階段列</strong
			>:黃色=尚未開賽(倒數至開賽)、綠色=進行中(剩餘時間)、灰色=已結束。
		</p>
		{@render screenshot(
			'/reference/submissions.png',
			'側欄選定專案後,頂部階段列顯示比賽剩餘時間;此頁為提交紀錄'
		)}

		<h3 class="mt-6 text-xl font-bold">4.2 公告與選手提問</h3>
		<ol class="mt-3 space-y-3">
			{@render step(1, '「新增公告」向全體參賽者發布訊息(選手端會收到通知)。')}
			{@render step(2, '選手提問會列在中段,點入可<strong>回覆</strong>;點狀態徽章可切換「已解決/未解決」。')}
		</ol>
		{@render screenshot('/reference/clarifications.png', '公告與選手提問')}

		<h3 class="mt-6 text-xl font-bold">4.3 提交紀錄與評分詳情</h3>
		<ol class="mt-3 space-y-3">
			{@render step(1, '「提交紀錄」即時列出選定專案的所有提交(分部分/完整提交)。點任一列進入<strong>提交詳情</strong>。')}
			{@render step(2, '提交詳情可「請求重新評分」(單筆重評)或「忽略」該筆提交(不採計)。')}
			{@render step(3, '點「裁判紀錄」中的任一筆評分,進入<strong>評分詳情頁</strong>:可用狀態篩選(未通過/通過/評分中/錯誤/逾時)與名稱搜尋快速聚焦,展開單項可看每個測試案例的指令與輸出。')}
		</ol>
		<div
			class="mt-3 rounded-lg border border-gray-200 bg-gray-50 p-4 text-sm dark:border-gray-700 dark:bg-gray-800"
		>
			<p>
				<strong>評測被取消時,取消原因選手看得到</strong>:它會顯示在選手的提交詳情頁(未附原因時顯示預設的「評測已由工作人員取消」)。目前版本的後台介面沒有手動取消單筆評測的入口;要向選手補充說明,請用下方的「給選手的附註」——措辭以「會被選手閱讀」為標準,不要寫內部代號或臆測。
			</p>
		</div>
		<div
			class="mt-3 rounded-lg border border-gray-200 bg-gray-50 p-4 text-sm dark:border-gray-700 dark:bg-gray-800"
		>
			<p class="font-bold">給選手的附註(評分詳情頁的「給選手的附註」)</p>
			<p class="mt-2">
				在<strong>評分詳情頁</strong>的「<strong>給選手的附註</strong>」欄位撰寫,按「<strong>儲存附註</strong>」生效。附註<strong>選手看得到</strong>:顯示在他的提交詳情頁、標示為「工作人員附註」,與系統的錯誤原因<strong>並存、不取代</strong>。措辭比照取消理由的標準——寫給選手看,不要放指令、路徑或內部細節。
			</p>
			<p class="mt-2">
				任何狀態的評測都可以寫,不限錯誤;<strong>後寫覆蓋先寫</strong>,<strong>清空欄位再儲存即撤回</strong>。常見動線:查明原因 → 寫附註 → 按「通知選手」發模板,選手就會被指到這則附註。
			</p>
		</div>
		<div
			class="mt-3 rounded-lg border border-gray-200 bg-gray-50 p-4 text-sm dark:border-gray-700 dark:bg-gray-800"
		>
			<p class="font-bold">通知選手(評分詳情頁的「通知選手」)</p>
			<p class="mt-2">
				在評分詳情頁按「<strong>通知選手</strong>」,可對<strong>該筆提交的選手本人</strong>發一則定向訊息——只有他看得到,不是公告。內建三個模板(檢查有結果、已重新評測、需修正後重交),會自動帶入該筆提交的<strong>提交時間</strong>(與選手「提交紀錄」顯示的時間逐字一致,方便對照),送出前可自由修改。選手會在「提問與回應」的提問列表看到這則訊息,並標示為「工作人員通知」。
			</p>
			<p class="mt-2">
				訊息<strong>選手看得到</strong>,措辭比照上一則(取消理由)的標準:寫給選手看,不要放內部代號或臆測。需要先在側欄選定該場競賽專案,「通知選手」鈕才會啟用。
			</p>
		</div>
		{@render screenshot('/reference/submission-detail.png', '提交詳情:採計分數、重評與忽略、裁判紀錄')}
		{@render screenshot('/reference/judgment.png', '評分詳情頁:狀態篩選與逐項結果')}

		<h3 class="mt-6 text-xl font-bold">4.4 重評紀錄</h3>
		<p class="mt-3">
			無論從提交詳情發起的單筆重評,或從競賽管理專案頁發起的整場重評,都會列在「<strong
				>重評紀錄</strong
			>」:顯示進度(已完成/受影響),點入可逐筆<strong>比對前後分數</strong>,確認無誤後核准(採計)或駁回。
		</p>
		{@render screenshot('/reference/rejudgings.png', '重評紀錄:進度與審核')}

		<h3 class="mt-6 text-xl font-bold">4.5 分數統計與匯出</h3>
		<p class="mt-3">
			「<strong>分數統計</strong>」列出各選手採計分數,點列可跳到該筆提交;「<strong
				>匯出分數明細</strong
			>」下載完整成績檔,供賽後計分存查。
		</p>
		{@render screenshot('/reference/user-scores.png', '分數統計與匯出')}
	</section>

	<!-- 5. kiosk -->
	<section id="kiosk" class="mt-10 scroll-mt-4">
		<h2 class="border-b-2 border-gray-200 pb-2 text-2xl font-bold dark:border-gray-700">
			5. 記分板與大螢幕(kiosk)
		</h2>
		<p class="mt-4">要在場地前方擺一台大螢幕輪播記分板,操作步驟:</p>
		<ol class="mt-3 space-y-3">
			{@render step(1, '在側欄「<strong>競賽過程</strong>」選定要展示的比賽與專案。')}
			{@render step(2, '點側欄「<strong>記分板展示</strong>」,視需要填入<strong>顯示文字</strong>(如目前題目)與<strong>底部公告</strong>,下方網址會即時更新。')}
			{@render step(3, '按「<strong>複製連結</strong>」,到場地機器的瀏覽器貼上開啟(可先用「在新分頁開啟」在自己電腦預覽)。')}
			{@render step(4, '在場地機器按 <strong>F11</strong> 進入全螢幕。')}
		</ol>
		{@render screenshot(
			'/reference/scoreboard-display.png',
			'記分板展示:為選定場次產生大螢幕連結'
		)}
		{@render screenshot('/reference/kiosk.png', 'kiosk 模式:倒數、標題文字與底部公告')}
		<p class="mt-4">產生的網址各參數的意義(也可手動調整):</p>
		<code
			class="mt-3 block overflow-x-auto rounded-md bg-gray-100 p-3 text-sm whitespace-pre text-gray-700 dark:bg-gray-800 dark:text-gray-300"
			>/scoreboard?kiosk&contestId=1&projectId=1&text=第一題:靜態網頁&footer=請保持安靜,禁止交談</code
		>
		<ul class="mt-3 list-inside list-disc space-y-1">
			<li><code>kiosk</code>:隱藏導覽列與側欄,全螢幕呈現,並顯示大字競賽倒數。</li>
			<li>
				<code>contestId / projectId</code>:鎖定顯示的場次——貼到任何機器都顯示同一場;若指定的場次不存在,畫面會<strong
					>直接顯示錯誤</strong
				>,不會改播其他場次。
			</li>
			<li><code>text=…</code>:標題下方的說明文字(如目前題目),省略則不顯示。</li>
			<li><code>footer=…</code>:固定於畫面底部的公告列,省略則不顯示。</li>
		</ul>
		<Alert color="green" class="mt-4 dark:bg-green-900 dark:text-green-100">
			<InfoCircleSolid slot="icon" class="h-5 w-5" />
			kiosk 使用公開資料串流,<strong>大螢幕不需要、也不應該登入任何帳號</strong>。
		</Alert>

		<p class="mt-6">
			大螢幕會自己維持畫面,<strong>現場毋須人工操作</strong>。連線出狀況時,畫面<strong
				>右上角</strong
			>會出現黃色告示,只有兩種:
		</p>
		<ul class="mt-3 list-inside list-disc space-y-2">
			<li>
				<strong>「即時連線異常,畫面可能不是最新的 —— 持續重新連線中」</strong
				>:系統正在自己重連,期間畫面仍停留在最後一次的成績。恢復後會自動補回最新記分板,告示也會自己消失——<strong
					>不必重整或重開網頁</strong
				>。
			</li>
			<li>
				<strong>「這場比賽已結束或已下架」</strong
				>:這是正常結束,不是故障。畫面停在最後一次更新,收掉螢幕即可。會看到這句話的情形有兩種:比賽的開放時間已經過去,或是這場的記分板已在後台取消公開。
			</li>
		</ul>
		{@render screenshot(
			'/reference/kiosk-degraded.png',
			'連線中斷時,kiosk 右上角出現黃色告示「即時連線異常,畫面可能不是最新的 —— 持續重新連線中」;記分板與底部公告照常顯示,停在最後一次更新的成績'
		)}
	</section>

	<!-- 6. 詞彙表 -->
	<section id="glossary" class="mt-10 scroll-mt-4">
		<h2 class="border-b-2 border-gray-200 pb-2 text-2xl font-bold dark:border-gray-700">
			6. 詞彙表
		</h2>

		<h3 class="mt-6 text-xl font-bold">專案的四個時間</h3>
		<ul class="mt-3 list-inside list-disc space-y-1">
			<li><strong>上線 activate</strong>:選手開始「看得到」這個專案。</li>
			<li><strong>開賽 start</strong>:選手可以開始提交;競賽倒數以此起算。</li>
			<li><strong>結束 end</strong>:停止收件;逾此時間無法再提交。</li>
			<li>
				<strong>下線 deactivate</strong>:專案自選手端與「競賽過程」選單消失;賽後檢視請走「競賽管理」。
			</li>
		</ul>

		<h3 class="mt-6 text-xl font-bold">提交/評分狀態</h3>
		<ul class="mt-3 list-inside list-disc space-y-1">
			<li><Badge color="dark">Queued</Badge> 已收件、排隊等待評分。</li>
			<li><Badge color="blue">Judging</Badge> 評分中。</li>
			<li><Badge color="green">Judged</Badge> 評分完成。</li>
			<li><Badge color="dark">Rejudged</Badge> 原評分已被採計的重評結果取代(見 4.4)。</li>
			<li><Badge color="yellow">Cancelled</Badge> 已取消。</li>
			<li><Badge color="red">Error</Badge> 評分失敗(見第 7 章)。</li>
		</ul>

		<h3 class="mt-6 text-xl font-bold">評分項目/測試案例狀態</h3>
		<ul class="mt-3 list-inside list-disc space-y-1">
			<li><Badge color="dark">Pending</Badge> 未開始、<Badge color="blue">Running</Badge> 執行中。</li>
			<li>
				<Badge color="green">Passed</Badge> 通過、<Badge color="primary">Failed</Badge> 未通過。
			</li>
			<li><Badge color="red">Error</Badge> 執行錯誤、<Badge color="indigo">Timeout</Badge> 逾時。</li>
		</ul>

		<h3 class="mt-6 text-xl font-bold">重評狀態</h3>
		<ul class="mt-3 list-inside list-disc space-y-1">
			<li><strong>Finished</strong>:重評跑完,等待審核。</li>
			<li><strong>Approved</strong>:已核准,重評結果正式採計。</li>
			<li><strong>Rejected</strong>:已駁回,維持原分數。</li>
			<li><strong>Canceled</strong>:重評已取消。</li>
		</ul>
	</section>

	<!-- 7. FAQ -->
	<section id="faq" class="mt-10 scroll-mt-4">
		<h2 class="border-b-2 border-gray-200 pb-2 text-2xl font-bold dark:border-gray-700">
			7. 疑難排解
		</h2>
		<Accordion multiple class="mt-4">
			<AccordionItem>
				<span slot="header">選手說按不了「提交」/ 提交按鈕是灰色的</span>
				<p>
					最常見原因:<strong>該帳號未被加入這場比賽的參賽者</strong>。到「競賽管理」→
					該比賽詳情,把選手(或其群組)加入參賽者即可,選手端不需重新登入。也請確認目前在
					<strong>開賽~結束</strong> 的時間窗內。
				</p>
			</AccordionItem>
			<AccordionItem>
				<span slot="header">提交一直停在 Queued / 排隊數持續增加</span>
				<p>
					打開「系統狀態」看兩件事:①「執行中」是否大於
					0——有東西在跑就只是提交高峰,排隊會隨評分完成消化,必要時可到「評分工作者管理」<strong
						>新增工作者</strong
					>加速;② 若出現紅色警示「有排隊但沒有任何評分在執行中」,或工作者全部停擺/ERROR,才是真的卡住——請刪除異常的工作者並重新建立。
				</p>
			</AccordionItem>
			<AccordionItem>
				<span slot="header">評分結果是 Error</span>
				<p>
					從「系統狀態」的<strong>最近的評分錯誤</strong>點入該筆評分,錯誤原因會顯示在頁面頂部;展開評分項目可看是哪個測試案例的輸出出錯。常見原因:選手壓縮檔結構不符、專案容器啟動失敗(unhealthy)。個案可在提交詳情按「請求重新評分」。
				</p>
			</AccordionItem>
			<AccordionItem>
				<span slot="header">畫面數字好像沒在動</span>
				<p>
					後台數據透過即時串流更新,若懷疑停住(例如倒數不動),先<strong>重新整理頁面</strong>;仍異常時檢查與伺服器的網路連線。
				</p>
			</AccordionItem>
			<AccordionItem>
				<span slot="header">被登出/顯示登入已過期</span>
				<p>
					登入有效期預設 7 天,逾期會自動導向登入頁,重新登入即可,資料不受影響。
				</p>
			</AccordionItem>
			<AccordionItem>
				<span slot="header">「競賽過程」選單找不到我的比賽</span>
				<p>
					此選單只列出<strong>上線期間內</strong>的比賽。已結束或尚未上線的比賽,請走「競賽管理」進入該專案頁——提交、提問、分數統計在那裡都看得到。
				</p>
			</AccordionItem>
		</Accordion>
	</section>
</div>
