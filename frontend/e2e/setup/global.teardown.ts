import { test as teardown } from '@playwright/test';
import { AdminApi, clearState, readState } from '../fixtures';

teardown('delete the e2e contest', async () => {
	let state;
	try {
		state = readState();
	} catch {
		return; // setup 沒跑成,無事可清
	}
	try {
		const api = await AdminApi.login();
		await api.send('delete', `/contests/${state.contestId}`);
		await api.dispose();
	} catch (e) {
		// 清不掉不要讓整個測試紅掉:比賽名帶 timestamp,殘留可辨識、可手動清
		console.warn(`teardown: failed to delete e2e contest ${state.contestId}:`, e);
	} finally {
		clearState();
	}
});
