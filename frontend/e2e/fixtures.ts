import { request, type APIRequestContext } from '@playwright/test';
import fs from 'node:fs';
import path from 'node:path';

// E2E 環境參數:全部有預設值,對照 repo 預設環境(vite 5173 + 後端 8080)
// 零設定可跑;CI/其他環境用環境變數抽換。

/** CI 且未指定 E2E_BASE_URL:playwright.config 會自行 build+preview 於 4173(自足模式) */
export const selfHost = !process.env.E2E_BASE_URL && !!process.env.CI;

export const env = {
	/** 瀏覽器要開的前端 origin(setup 自建 context 也用這個,必須與 config 同源) */
	baseURL:
		process.env.E2E_BASE_URL ?? (selfHost ? 'http://localhost:4173' : 'http://localhost:5173'),
	/** 測試 runner 直呼 API 的位址(setup/teardown 造資料用) */
	apiURL: process.env.E2E_API_URL ?? 'http://localhost:8080/api/v1',
	admin: {
		username: process.env.E2E_ADMIN_USER ?? 'admin',
		password: process.env.E2E_ADMIN_PASS ?? '1234'
	},
	contestant: {
		username: process.env.E2E_CONTESTANT_USER ?? 'demo01',
		password: process.env.E2E_CONTESTANT_PASS ?? '1234'
	}
};

export const AUTH_DIR = path.join(import.meta.dirname, '.auth');
export const ADMIN_STATE = path.join(AUTH_DIR, 'admin.json');
export const CONTESTANT_STATE = path.join(AUTH_DIR, 'contestant.json');
const STATE_FILE = path.join(import.meta.dirname, '.state.json');

/** global setup 自建的 E2E 比賽資訊,供各測試與 teardown 使用 */
export type E2EState = {
	contestId: number;
	contestName: string;
	/** contest-project id(kiosk pin、picker、提交都用這個) */
	contestProjectId: number;
	projectName: string;
};

export function writeState(state: E2EState) {
	fs.writeFileSync(STATE_FILE, JSON.stringify(state, null, '\t'));
}

export function readState(): E2EState {
	if (!fs.existsSync(STATE_FILE)) {
		throw new Error('e2e/.state.json not found — did the setup project run?');
	}
	return JSON.parse(fs.readFileSync(STATE_FILE, 'utf-8'));
}

export function clearState() {
	fs.rmSync(STATE_FILE, { force: true });
}

/**
 * 以 admin 身分呼叫後端 API 的最小 client(cookie 由 request context 自管;
 * 寫入類請求需帶 csrf-token header)。
 */
export class AdminApi {
	private constructor(private readonly ctx: APIRequestContext) {}

	static async login(): Promise<AdminApi> {
		// 不用 request context 的 baseURL:Playwright 對 '/path' 會以 origin 解析,
		// 丟掉 apiURL 內的 /api/v1 前綴 —— 一律手動組完整 URL。
		const ctx = await request.newContext();
		const api = new AdminApi(ctx);
		const res = await ctx.post(`${env.apiURL}/auth/login`, {
			headers: { 'csrf-token': await api.csrfToken() },
			data: { username: env.admin.username, password: env.admin.password }
		});
		if (!res.ok()) {
			throw new Error(`admin login failed: ${res.status()} ${await res.text()}`);
		}
		return api;
	}

	private async csrfToken(): Promise<string> {
		const res = await this.ctx.get(`${env.apiURL}/csrf-token`);
		return (await res.json()).data.token;
	}

	async get<T = unknown>(url: string): Promise<T> {
		const res = await this.ctx.get(`${env.apiURL}${url}`);
		if (!res.ok()) throw new Error(`GET ${url} → ${res.status()} ${await res.text()}`);
		return (await res.json()).data as T;
	}

	/** 二進位下載(匯出檔之類),回原始 bytes。 */
	async download(url: string): Promise<Buffer> {
		const res = await this.ctx.get(`${env.apiURL}${url}`);
		if (!res.ok()) throw new Error(`GET ${url} → ${res.status()} ${await res.text()}`);
		return res.body();
	}

	async send(method: 'post' | 'patch' | 'delete', url: string, data?: unknown): Promise<unknown> {
		const res = await this.ctx[method](`${env.apiURL}${url}`, {
			headers: { 'csrf-token': await this.csrfToken() },
			data
		});
		if (!res.ok()) throw new Error(`${method} ${url} → ${res.status()} ${await res.text()}`);
		return res.json();
	}

	async dispose() {
		await this.ctx.dispose();
	}
}
