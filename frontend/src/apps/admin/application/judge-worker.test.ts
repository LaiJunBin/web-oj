import { JudgeWorkerApplication } from '@/apps/admin/application/judge-worker';
import type { JudgeWorkerServiceInterface } from '@/apps/admin/application/service/judge-worker';
import type { JudgeWorker, JudgeWorkersStorageInterface } from '@/apps/admin/domain/judge-worker';
import type { Storage } from '@/apps/common/domain/storage';
import { describe, expect, test, vi } from 'vitest';

/** 和真正的 BrowserStorage 一樣,setError 會把資料清掉——bug 的形狀要忠實複製。 */
class FakeStorage<T> implements Storage<T> {
	private value: T | null = null;
	error: Error | null = null;
	set(data: T) {
		this.value = data;
		this.error = null;
	}
	get() {
		return this.value;
	}
	clear() {
		this.value = null;
	}
	setError(error: Error) {
		this.value = null;
		this.error = error;
	}
}

const workers = [{ id: 'w1', status: 'IDLE' }] as unknown as JudgeWorker[];

function appWith(fetchJudgeWorkers: JudgeWorkerServiceInterface['fetchJudgeWorkers']) {
	// 只實作測試會碰到的方法;介面另有 create/delete/start/stop,這裡用不到,
	// 所以經 unknown 轉型而不是把七個都假造一遍。
	const service = {
		fetchJudgeWorkers,
		fetchJudgePriorityStrategies: () => Promise.resolve([]),
		fetchSystemDefaults: () => Promise.reject(new Error('unused'))
	} as unknown as JudgeWorkerServiceInterface;
	const storage: JudgeWorkersStorageInterface = new FakeStorage<JudgeWorker[]>();
	const app = new JudgeWorkerApplication(service, storage, new FakeStorage(), new FakeStorage());
	return { app, storage };
}

// 與 scoreboard 同一個模式:dashboard 的 worker 清單在 stream 斷線期間每 30 秒
// 以 REST 回補,失敗若走 setError,清單會被清空、畫面顯示「尚無評分工作者」——
// 對正在處理故障的管理員,那是一句錯的話。
describe('JudgeWorkerApplication.resyncJudgeWorkers', () => {
	test('失敗時保留畫面上的清單,連 error 都不設', async () => {
		const { app, storage } = appWith(() => Promise.reject(new Error('502')));
		storage.set(workers);
		vi.spyOn(console, 'warn').mockImplementation(() => {});

		await app.resyncJudgeWorkers();

		expect(storage.get()).toBe(workers);
		expect(storage.error).toBeNull();
	});

	test('成功時照常覆寫', async () => {
		const fresh = [] as JudgeWorker[];
		const { app, storage } = appWith(() => Promise.resolve(fresh));
		storage.set(workers);

		await app.resyncJudgeWorkers();

		expect(storage.get()).toBe(fresh);
	});
});
