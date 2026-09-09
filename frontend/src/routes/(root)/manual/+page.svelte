<script lang="ts">
	import { Accordion, AccordionItem, Alert, Badge } from 'flowbite-svelte';
	import { InfoCircleSolid } from 'flowbite-svelte-icons';

	// 章節目錄:id 對應各 <section> 的 id,點擊平滑捲動至該節
	const sections = [
		{ id: 'overview', title: '1. 系統與畫面總覽' },
		{ id: 'login', title: '2. 登入' },
		{ id: 'select', title: '3. 選擇競賽與專案' },
		{ id: 'schedule', title: '4. 競賽日程' },
		{ id: 'timeline', title: '5. 開賽前後看得到什麼' },
		{ id: 'project', title: '6. 專案概述與下載素材' },
		{ id: 'grading', title: '7. 評分項目與公開測資' },
		{ id: 'submit', title: '8. 如何提交' },
		{ id: 'submissions', title: '9. 提交紀錄與即時結果' },
		{ id: 'status', title: '10. 狀態詞彙表' },
		{ id: 'scoreboard', title: '11. 記分板' },
		{ id: 'clarification', title: '12. 提問與公告、通知' },
		{ id: 'faq', title: '13. 常見問題與疑難排解' }
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
		<figcaption class="mt-2 text-center text-sm text-gray-500 dark:text-gray-400">{caption}</figcaption>
	</figure>
{/snippet}

{#snippet step(n: number, text: string)}
	<li class="flex gap-3">
		<span
			class="flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full bg-green-600 text-sm font-bold text-white"
			>{n}</span
		>
		<span class="pt-0.5">{@html text}</span>
	</li>
{/snippet}

<div class="mx-auto max-w-4xl px-4 py-8 text-gray-800 dark:text-gray-100">
	<h1 class="text-center text-4xl font-extrabold tracking-tight text-gray-900 dark:text-white">
		操作手冊
	</h1>
	<p class="mt-3 text-center text-gray-500 dark:text-gray-400">
		第一次使用 Web Judge 競賽系統?依照本手冊即可順利完成競賽。
	</p>

	<Alert color="green" class="mt-6 dark:bg-green-900 dark:text-green-100">
		<InfoCircleSolid slot="icon" class="h-5 w-5" />
		<span class="font-semibold">競賽現場三步驟快速上手：</span>
		① 用<strong>帳號</strong>與密碼<strong>登入</strong> → ②
		在上方選擇你的<strong>競賽</strong>與<strong>專案</strong> → ③
		到了提交時段,點右上角<strong>「提交」</strong>上傳你的專案壓縮檔。
	</Alert>

	<!-- 目錄 -->
	<nav
		class="mt-8 rounded-lg border border-gray-200 bg-white p-5 shadow-sm dark:border-gray-700 dark:bg-gray-800"
		aria-label="目錄"
	>
		<h2 class="mb-3 text-lg font-bold text-gray-900 dark:text-white">目錄</h2>
		<ul class="grid gap-x-6 gap-y-2 sm:grid-cols-2">
			{#each sections as s}
				<li>
					<a href={`#${s.id}`} class="text-primary-600 hover:underline dark:text-primary-400">
						{s.title}
					</a>
				</li>
			{/each}
		</ul>
	</nav>

	<!-- 1. 系統與畫面總覽 -->
	<section id="overview" class="mt-12 scroll-mt-24">
		<h2 class="border-b-2 border-gray-200 pb-2 text-2xl font-bold dark:border-gray-700">
			1. 系統與畫面總覽
		</h2>
		<p class="mt-4">
			Web Judge 是線上程式競賽的評測系統。登入後,你在<strong>畫面最上方的工具列(Navbar)</strong>可以操作大部分功能:
		</p>
		<ul class="mt-3 list-disc space-y-2 pl-6">
			<li><strong>競賽選擇 / 專案選擇</strong>：切換你要參加的競賽,以及該競賽底下的專案(題目)。</li>
			<li><strong>競賽日程</strong>：查看每個專案的開始/結束時間與配分。</li>
			<li><strong>提交</strong>：在提交時段內上傳你的作答(僅參賽者可用)。</li>
			<li><strong>右上角時鐘</strong>：顯示目前狀態 —「還有多久開放提交」、提交中的<strong>倒數</strong>、或<strong>已結束</strong>。</li>
			<li><strong>🔔 鈴鐺</strong>：開啟瀏覽器通知,即時收到公告與提問回覆。</li>
			<li><strong>登入 / 登出</strong>與<strong>深色模式</strong>切換。</li>
		</ul>
		{@render screenshot('/manual/navbar.png', '登入後的頂部工具列與各功能位置')}
	</section>

	<!-- 2. 登入 -->
	<section id="login" class="mt-12 scroll-mt-24">
		<h2 class="border-b-2 border-gray-200 pb-2 text-2xl font-bold dark:border-gray-700">
			2. 登入
		</h2>
		<ol class="mt-4 space-y-3">
			{@render step(1, '點擊右上角<strong>「登入」</strong>按鈕,進入登入頁。')}
			{@render step(2, '在<strong>「帳號」</strong>欄輸入主辦方提供的帳號(示範環境可用 <code>demo01</code>)。')}
			{@render step(3, '輸入<strong>密碼</strong>,按<strong>「登入」</strong>。')}
		</ol>
		<Alert color="yellow" class="mt-4 dark:bg-yellow-800 dark:text-yellow-200">
			<InfoCircleSolid slot="icon" class="h-5 w-5" />
			帳號與密碼由主辦方(競賽現場工作人員)提供。若顯示「登入失敗」,請確認帳號密碼是否輸入正確,或洽詢現場工作人員;若提示<strong
				>嘗試過於頻繁</strong
			>,請稍候片刻再試(見第 13 章)。
		</Alert>
		{@render screenshot('/manual/login.png', '登入畫面')}
	</section>

	<!-- 3. 選擇競賽與專案 -->
	<section id="select" class="mt-12 scroll-mt-24">
		<h2 class="border-b-2 border-gray-200 pb-2 text-2xl font-bold dark:border-gray-700">
			3. 選擇競賽與專案
		</h2>
		<p class="mt-4">
			一場競賽可能包含多個<strong>專案(題目)</strong>。登入後請確認上方選到的是正確的競賽與專案:
		</p>
		<ul class="mt-3 list-disc space-y-2 pl-6">
			<li>點擊工具列的<strong>競賽名稱</strong>下拉,選擇你要參加的競賽。</li>
			<li>點擊<strong>專案名稱</strong>下拉,選擇目前要作答的專案。</li>
			<li>你在畫面上看到的概述、評分項目、提交與記分板,都會對應<strong>目前選取的專案</strong>。</li>
		</ul>
		{@render screenshot('/manual/select.png', '競賽與專案下拉選單')}
	</section>

	<!-- 4. 競賽日程 -->
	<section id="schedule" class="mt-12 scroll-mt-24">
		<h2 class="border-b-2 border-gray-200 pb-2 text-2xl font-bold dark:border-gray-700">
			4. 競賽日程
		</h2>
		<p class="mt-4">
			點工具列的<strong>「競賽日程」</strong>,可看到目前競賽中每個專案的<strong>開始時間、結束時間、競賽時長與配分</strong>。建議開賽前先確認各專案的時間安排。
		</p>
		{@render screenshot('/manual/schedule.png', '競賽日程時間軸')}
	</section>

	<!-- 5. 開賽前後 -->
	<section id="timeline" class="mt-12 scroll-mt-24">
		<h2 class="border-b-2 border-gray-200 pb-2 text-2xl font-bold dark:border-gray-700">
			5. 開賽前後看得到什麼
		</h2>
		<p class="mt-4">部分內容會等到<strong>該專案開始</strong>後才公開:</p>
		<div class="mt-4 grid gap-4 sm:grid-cols-2">
			<div class="rounded-lg border border-gray-200 p-4 dark:border-gray-700">
				<h3 class="font-bold">開賽前</h3>
				<ul class="mt-2 list-disc space-y-1 pl-5 text-sm">
					<li>可以登入、選擇競賽/專案、看競賽日程。</li>
					<li>專案描述、素材、評分項目會顯示<strong>「待比賽開始後公開」</strong>。</li>
					<li>時鐘顯示<strong>「還有多久開放提交」</strong>。</li>
				</ul>
			</div>
			<div class="rounded-lg border border-gray-200 p-4 dark:border-gray-700">
				<h3 class="font-bold">開賽後</h3>
				<ul class="mt-2 list-disc space-y-1 pl-5 text-sm">
					<li>解鎖專案描述、下載素材、評分項目與公開測資。</li>
					<li>時鐘變成<strong>提交倒數</strong>,右上角出現<strong>「提交」</strong>按鈕。</li>
					<li>時間到後時鐘顯示<strong>「已結束」</strong>,無法再提交。</li>
				</ul>
			</div>
		</div>
		{@render screenshot('/manual/clock-wait.png', '開賽前：時鐘顯示「還有多久開放提交」,尚無提交按鈕')}
		{@render screenshot(
			'/manual/clock-active.png',
			'競賽進行中：時鐘顯示提交倒數,右上角出現綠色「提交」按鈕'
		)}
	</section>

	<!-- 6. 專案概述 -->
	<section id="project" class="mt-12 scroll-mt-24">
		<h2 class="border-b-2 border-gray-200 pb-2 text-2xl font-bold dark:border-gray-700">
			6. 專案概述與下載素材
		</h2>
		<p class="mt-4">
			在左側選單點<strong>「專案概述」</strong>,可閱讀<strong>專案描述</strong>(題目說明),以及下載<strong>「提供素材」</strong>(題目附帶的檔案)。點素材按鈕即可下載到本機。
		</p>
		{@render screenshot('/manual/overview.png', '專案概述與素材下載')}
	</section>

	<!-- 7. 評分項目 -->
	<section id="grading" class="mt-12 scroll-mt-24">
		<h2 class="border-b-2 border-gray-200 pb-2 text-2xl font-bold dark:border-gray-700">
			7. 評分項目與公開測資
		</h2>
		<p class="mt-4">
			左側選單的<strong>「評分項目」</strong>列出這個專案的所有評分項目,每一項包含:
		</p>
		<ul class="mt-3 list-disc space-y-2 pl-6">
			<li><strong>配分</strong>與<strong>說明</strong>:該項目的要求與分數。</li>
			<li><strong>公開的測試案例</strong>:可作為你自我檢查的參考。實際評分還包含未公開的測資。</li>
		</ul>
		{@render screenshot('/manual/grading.png', '評分項目與公開測資列表')}
	</section>

	<!-- 8. 如何提交 -->
	<section id="submit" class="mt-12 scroll-mt-24">
		<h2 class="border-b-2 border-gray-200 pb-2 text-2xl font-bold dark:border-gray-700">
			8. 如何提交
		</h2>
		<p class="mt-4">
			<strong>在提交時段內</strong>,右上角會出現綠色<strong>「提交」</strong>按鈕。點擊後在跳出的視窗中:
		</p>
		<ol class="mt-4 space-y-3">
			{@render step(1, '選擇<strong>評分項目</strong>：選「全部(ALL)」對<strong>所有項目</strong>評分(=完整提交);或選<strong>單一項目</strong>只評該項(=部分提交)。')}
			{@render step(2, '上傳<strong>專案壓縮檔(.zip)</strong>：必填,請將作答打包成 zip 後上傳。')}
			{@render step(3, '若專案有<strong>額外欄位</strong>(勾選/文字/檔案),依題目要求填寫。')}
			{@render step(4, '點<strong>「送出」</strong>。送出後可到「提交紀錄」查看評分進度。')}
		</ol>

		<Alert color="blue" class="mt-4 dark:bg-blue-900 dark:text-blue-100">
			<InfoCircleSolid slot="icon" class="h-5 w-5" />
			<span class="font-semibold">部分提交 vs 完整提交</span><br />
			<Badge color="indigo">完整提交</Badge> 對所有評分項目評分;<Badge color="pink">部分提交</Badge>
			只評你選的單一項目。<strong>記分板僅採計「完整提交」的結果</strong>,因此最終請務必做一次完整提交。
		</Alert>

		<Alert color="yellow" class="mt-3 dark:bg-yellow-800 dark:text-yellow-200">
			<InfoCircleSolid slot="icon" class="h-5 w-5" />
			看不到「提交」按鈕?可能是<strong>尚未開放提交/已結束</strong>,或你<strong>不在本場次的參賽名單</strong>(按鈕會呈現灰色並提示「您不是參賽者」)。詳見
			<a href="#faq" class="underline">常見問題</a>。
		</Alert>

		<Alert color="yellow" class="mt-3 dark:bg-yellow-800 dark:text-yellow-200">
			<InfoCircleSolid slot="icon" class="h-5 w-5" />
			<span class="font-semibold">提交間隔(冷卻)</span><br />
			部分題目會限制<strong>兩次提交之間的最短間隔</strong>。冷卻中提交會被擋下,「提交」按鈕會顯示<strong>剩餘秒數</strong>並暫時停用,倒數歸零後自動恢復——這不是錯誤,等一下再送即可。幾點要知道:間隔從<strong>上一次送出</strong>起算(不是從評分完成起算);<strong>取消提交不會重置</strong>間隔;各題的間隔<strong>各自獨立</strong>,某題冷卻中仍可提交其他題。
		</Alert>
		<Alert color="yellow" class="mt-3 dark:bg-yellow-800 dark:text-yellow-200">
			<InfoCircleSolid slot="icon" class="h-5 w-5" />
			<span class="font-semibold">評分時你的作品連不到網際網路</span><br />
			評分是在一個<strong>對外封閉的容器</strong>裡執行你的作品:它連不到網際網路,也連不到會場的其他機器。所以
			<strong>不要讓作品在執行時才去下載東西或呼叫外部 API</strong
			>(例如從 CDN 抓字型或函式庫、連線第三方服務),那些呼叫在評分時一定會失敗。需要的檔案請<strong
				>一起放進壓縮檔</strong
			>。這是評分環境的固定條件,對每位選手都一樣。
		</Alert>
		{@render screenshot('/manual/submit.png', '提交視窗(評分項目、壓縮檔上傳、額外欄位)')}
	</section>

	<!-- 9. 提交紀錄 -->
	<section id="submissions" class="mt-12 scroll-mt-24">
		<h2 class="border-b-2 border-gray-200 pb-2 text-2xl font-bold dark:border-gray-700">
			9. 提交紀錄與即時結果
		</h2>
		<p class="mt-4">
			左側選單的<strong>「提交紀錄」</strong>列出你所有的提交,每筆顯示是<strong>部分/完整提交</strong>、目前<strong>狀態</strong>、<strong>得分</strong>與提交時間。點一筆可進入<strong>詳情</strong>:
		</p>
		<ul class="mt-3 list-disc space-y-2 pl-6">
			<li><strong>即時更新</strong>：評分進行中,狀態與分數會自動更新,不需重新整理。</li>
			<li><strong>各測資結果</strong>：依評分項目列出每個測試案例的通過狀況;<strong>僅顯示公開測資的輸出</strong>。</li>
			<li>
				<strong>取消提交</strong>：當狀態為 <Badge color="dark">Queued</Badge> 或
				<Badge color="blue">Judging</Badge> 時,右上角會出現<strong>「取消提交」</strong>,可中止本次評分。
			</li>
		</ul>
		{@render screenshot('/manual/submission.png', '提交詳情與各測資結果')}
	</section>

	<!-- 10. 狀態詞彙表 -->
	<section id="status" class="mt-12 scroll-mt-24">
		<h2 class="border-b-2 border-gray-200 pb-2 text-2xl font-bold dark:border-gray-700">
			10. 狀態詞彙表
		</h2>
		<p class="mt-4">評分過程中你會看到以下狀態:</p>
		<Accordion multiple class="mt-4">
			<AccordionItem open>
				<span slot="header">提交狀態</span>
				<ul class="space-y-2">
					<li><Badge color="dark">Queued</Badge> 已排入佇列,等待評分。</li>
					<li><Badge color="blue">Judging</Badge> 評分進行中。</li>
					<li><Badge color="green">Judged</Badge> 評分完成。</li>
					<li><Badge color="yellow">Cancelled</Badge> 已取消。</li>
					<li><Badge color="red">Error</Badge> 評分發生錯誤(如環境或系統問題)。</li>
				</ul>
			</AccordionItem>
			<AccordionItem open>
				<span slot="header">單一測資狀態</span>
				<ul class="space-y-2">
					<li><Badge color="dark">Pending</Badge> 等待執行。</li>
					<li><Badge color="blue">Running</Badge> 執行中。</li>
					<li><Badge color="green">Passed</Badge> 通過。</li>
					<li><Badge color="red">Failed</Badge> 未通過(輸出不符預期)。</li>
					<li><Badge color="indigo">Timeout</Badge> 逾時(執行時間超過限制)。</li>
					<li><Badge color="red">Error</Badge> 執行錯誤(如程式崩潰)。</li>
					<li><Badge color="yellow">Cancelled</Badge> 已取消。</li>
					<li><Badge color="dark">NotSubmitted</Badge> 這次(部分)提交未涵蓋的測資,不執行、不計分。</li>
				</ul>
			</AccordionItem>
		</Accordion>
	</section>

	<!-- 11. 記分板 -->
	<section id="scoreboard" class="mt-12 scroll-mt-24">
		<h2 class="border-b-2 border-gray-200 pb-2 text-2xl font-bold dark:border-gray-700">
			11. 記分板
		</h2>
		<p class="mt-4">左側選單的<strong>「記分板」</strong>以格狀顯示各參賽者在每個評分項目的通過狀況:</p>
		<ul class="mt-3 list-disc space-y-2 pl-6">
			<li><strong>列</strong>=參賽者,<strong>欄</strong>=評分項目(以編號表示,滑鼠移到欄標題可看名稱)。</li>
			<li>
				格子顏色:<span class="mx-1 inline-block h-4 w-4 rounded-sm bg-green-300 align-middle"></span>綠=通過(Passed)、<span
					class="mx-1 inline-block h-4 w-4 rounded-sm bg-red-300 align-middle"
				></span>紅=未通過(Failed)。
			</li>
			<li>點欄標題可跳到對應的<strong>評分項目</strong>說明。</li>
			<li>表格較寬時,可用滑鼠<strong>拖曳</strong>捲動。</li>
		</ul>
		<Alert color="yellow" class="mt-4 dark:bg-yellow-800 dark:text-yellow-200">
			<InfoCircleSolid slot="icon" class="h-5 w-5" />
			記分板<strong>僅採計「完整提交」</strong>的評分結果。
		</Alert>
		{@render screenshot('/manual/scoreboard.png', '記分板')}
	</section>

	<!-- 12. 提問與公告 -->
	<section id="clarification" class="mt-12 scroll-mt-24">
		<h2 class="border-b-2 border-gray-200 pb-2 text-2xl font-bold dark:border-gray-700">
			12. 提問與公告、通知
		</h2>
		<p class="mt-4">左側選單的<strong>「提問與回應」</strong>分為兩部分:</p>
		<ul class="mt-3 list-disc space-y-2 pl-6">
			<li><strong>公告列表</strong>：主辦方發布的公告,請隨時留意。</li>
			<li><strong>提問列表</strong>：你自己向主辦方提出的問題。點右上角<strong>「+」</strong>,填寫<strong>主題</strong>與<strong>內容</strong>後送出;點任一筆可查看回覆。</li>
		</ul>
		<p class="mt-4">
			建議點工具列的<strong>🔔 鈴鐺</strong>開啟<strong>瀏覽器通知</strong>,即可在有新公告或提問回覆時立即收到提醒(需允許瀏覽器的通知權限)。
		</p>
		{@render screenshot('/manual/clarification.png', '提問與公告、新增提問視窗')}
	</section>

	<!-- 13. 常見問題 -->
	<section id="faq" class="mt-12 scroll-mt-24">
		<h2 class="border-b-2 border-gray-200 pb-2 text-2xl font-bold dark:border-gray-700">
			13. 常見問題與疑難排解
		</h2>
		<Accordion multiple class="mt-4">
			<AccordionItem>
				<span slot="header">看不到「提交」按鈕?</span>
				<ul class="list-disc space-y-1 pl-5">
					<li>目前<strong>尚未開放提交</strong>或<strong>已結束</strong>(看右上角時鐘)。</li>
					<li>尚未<strong>登入</strong>,或未選擇正確的競賽/專案。</li>
					<li>你<strong>不在本場次的參賽名單</strong>中 — 按鈕會呈灰色並顯示「您不是參賽者」,請洽工作人員確認報名/分組。</li>
				</ul>
			</AccordionItem>
			<AccordionItem>
				<span slot="header">按了「提交」卻顯示失敗?</span>
				<ul class="list-disc space-y-1 pl-5">
					<li>請確認已上傳<strong>.zip</strong> 壓縮檔,且必填的額外欄位都已填寫。</li>
					<li>若提示「您不在本場次參賽名單」,代表你並非參賽者,請洽工作人員。</li>
					<li>可能已<strong>超過提交時段</strong>,請確認時鐘狀態。</li>
				</ul>
			</AccordionItem>
			<AccordionItem>
				<span slot="header">「提交」按鈕變成倒數、按不下去?</span>
				<p>
					該題設有<strong>提交間隔</strong>,你剛剛的提交還在冷卻中。按鈕上的秒數歸零後會自動恢復,不需要重新整理。冷卻是<strong>逐題獨立</strong>的,這段期間仍可提交其他題;<strong>取消提交不會</strong>讓間隔重新開始。
				</p>
			</AccordionItem>
			<AccordionItem>
				<span slot="header">為什麼看不到專案描述 / 評分項目?</span>
				<p>這些內容會在<strong>該專案開始後</strong>才公開,開賽前會顯示「待比賽開始後公開」。</p>
			</AccordionItem>
			<AccordionItem>
				<span slot="header">提交後看不到部分測資的輸出?</span>
				<p>系統<strong>僅提供公開測試案例</strong>的輸出結果;未公開的測資只計分、不顯示輸出。</p>
			</AccordionItem>
			<AccordionItem>
				<span slot="header">記分板沒有我的分數 / 分數不如預期?</span>
				<p>
					記分板<strong>僅採計「完整提交」</strong>。若你只做過部分提交,請再做一次<strong>完整提交</strong>(評分項目選「全部(ALL)」)。
				</p>
			</AccordionItem>
			<AccordionItem>
				<span slot="header">狀態顯示 Error 或 Timeout?</span>
				<ul class="list-disc space-y-1 pl-5">
					<li><strong>Timeout</strong>:程式執行超過時間限制,請檢查效能或無窮迴圈。</li>
					<li><strong>Error</strong>:評分環境或前置步驟發生錯誤,不是你的程式造成的。提交詳情頁會顯示原因,並提示你向現場工作人員或裁判確認——照做即可,不必自行重交。若詳情頁另有藍色的「工作人員附註」,以附註的說明為準。</li>
				</ul>
			</AccordionItem>
			<AccordionItem>
				<span slot="header">登入失敗?</span>
				<p>請確認<strong>帳號</strong>與<strong>密碼</strong>輸入正確(注意大小寫);仍無法登入請洽現場工作人員。</p>
				<p class="mt-2">
					若提示<strong>嘗試過於頻繁</strong>,那不是帳密錯誤,而是系統對<strong
						>短時間內連續登入失敗</strong
					>的保護:<strong>稍候片刻再試即可</strong>,不需要一直重按(一直重按只會讓它更久才恢復)。若你確定帳號密碼無誤卻反覆看到這個提示,請洽現場工作人員。
				</p>
			</AccordionItem>
			<AccordionItem>
				<span slot="header">畫面右下角出現黃色告示?</span>
				<p>
					若右下角出現<strong>「即時連線異常,畫面可能不是最新的 —— 持續重新連線中」</strong
					>,表示系統暫時收不到即時更新,<strong>正在自動重新連線</strong>。這段期間倒數與記分板會停在最後一次更新的狀態;連線恢復後畫面會<strong
						>自動補回最新內容</strong
					>,告示也會自己消失——<strong>不需要重新整理</strong>。
				</p>
				<p class="mt-2">
					你已經送出的提交<strong>不受影響</strong>,它在系統裡照常評測。若這則告示持續很久沒有消失,請告知現場工作人員。
				</p>
				<p class="mt-2">
					若出現的是<strong>「即時連線已中斷(未授權),畫面停在最後一次更新」</strong
					>,表示連線已中止且<strong>不會自行恢復</strong>,請<strong>重新登入</strong
					>;若重新登入後仍出現,請告知現場工作人員。
				</p>
			</AccordionItem>
		</Accordion>
	</section>

	<p class="mt-12 border-t border-gray-200 pt-6 text-center text-sm text-gray-500 dark:border-gray-700 dark:text-gray-400">
		祝競賽順利!如有系統操作問題,請洽競賽現場工作人員。
	</p>
</div>
