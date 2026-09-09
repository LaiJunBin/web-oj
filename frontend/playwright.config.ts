import { defineConfig } from '@playwright/test';
import { env, selfHost } from './e2e/fixtures';

// E2E 以環境參數抽換目標(見 e2e/fixtures.ts):
// - E2E_BASE_URL 指定要驅動的前端(dev container 對既有環境:http://localhost:5173)
// - 未指定且在 CI 時自行 build + preview(自足模式)
// - E2E_API_URL 指定 setup/teardown 造資料要打的後端 API
// baseURL 一律取自 fixtures 的 env.baseURL——setup 會用它自建 context,兩邊必須同源。

export default defineConfig({
	testDir: 'e2e',
	// 測試共用同一場自建比賽與同一個後端,循序執行避免狀態競爭
	fullyParallel: false,
	workers: 1,
	// 對長駐環境跑,網路/轉發偶發抖動在所難免:重試一次;真 bug 仍會連續失敗
	retries: process.env.CI ? 2 : 1,
	use: {
		baseURL: env.baseURL
	},
	webServer: selfHost
		? {
				command: 'npm run build && npm run preview',
				port: 4173,
				// 指令含冷 build,CI 上 60 秒的預設等待不夠
				timeout: 180_000
			}
		: undefined,
	projects: [
		{
			name: 'setup',
			testMatch: /setup\/global\.setup\.ts/,
			teardown: 'cleanup'
		},
		{
			name: 'cleanup',
			testMatch: /setup\/global\.teardown\.ts/
		},
		{
			// Tier 1:不需評分管線完成,任何環境皆可跑
			name: 'smoke',
			testMatch: /tests\/smoke\/.*\.spec\.ts/,
			dependencies: ['setup']
		},
		{
			// Tier 2:評分管線端到端(後端所在機器需能開評分容器)
			name: 'full',
			testMatch: /tests\/full\/.*\.spec\.ts/,
			dependencies: ['setup']
		}
	]
});
