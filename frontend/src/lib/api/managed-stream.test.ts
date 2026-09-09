import { afterEach, beforeEach, describe, expect, test, vi } from 'vitest';
import { createManagedStream, isTerminalStatus, retryDelay } from './managed-stream';
import type { EventSourcePolyfill } from 'event-source-polyfill';

/** Minimal stand-in for the polyfill: records listeners so a test can fire them. */
class FakeSource {
	listeners = new Map<string, ((event: unknown) => void)[]>();
	onerror: ((event: unknown) => void) | null = null;
	closed = false;

	addEventListener(name: string, handler: (event: unknown) => void) {
		const existing = this.listeners.get(name) ?? [];
		this.listeners.set(name, [...existing, handler]);
	}

	close() {
		this.closed = true;
	}

	emit(name: string, data?: unknown) {
		for (const handler of this.listeners.get(name) ?? []) handler({ data: JSON.stringify(data) });
	}

	/** The polyfill dispatches this once the response is 200 text/event-stream. */
	opened() {
		this.emit('open');
	}

	/**
	 * The polyfill's error event. A `status` is present only when the INITIAL
	 * response was rejected; a mid-stream drop carries none.
	 */
	errored(status?: number) {
		this.onerror?.(status === undefined ? {} : { status });
	}
}

function harness(overrides: Partial<Parameters<typeof createManagedStream>[0]> = {}) {
	const opened: FakeSource[] = [];
	const onDegraded = vi.fn();
	const onRecovered = vi.fn();
	const time = vi.fn();
	const stop = createManagedStream({
		connect: () => {
			const source = new FakeSource();
			opened.push(source);
			return source as unknown as EventSourcePolyfill;
		},
		on: { time },
		onDegraded,
		onRecovered,
		...overrides
	});
	return { opened, onDegraded, onRecovered, time, stop, latest: () => opened[opened.length - 1] };
}

/** A retryable upstream failure — the backend is broken, not the contest. */
const fail5xx = (source: FakeSource) =>
	source.emit('upstream-error', { stream: 'scoreboard', status: 500 });

beforeEach(() => vi.useFakeTimers());
afterEach(() => vi.useRealTimers());

describe('retryDelay', () => {
	test('doubles from one second and caps at thirty', () => {
		expect([1, 2, 3, 4, 5, 6, 7, 20].map(retryDelay)).toEqual([
			1_000, 2_000, 4_000, 8_000, 16_000, 30_000, 30_000, 30_000
		]);
	});
});

describe('isTerminalStatus', () => {
	test('a contest that ended, a board taken down, or a refused session', () => {
		// 404 is how a contest stream reports its own activation window closing,
		// and how a public scoreboard reports being un-published — normal events,
		// not faults, and retrying them would hammer the backend forever.
		expect([401, 403, 404].map(isTerminalStatus)).toEqual([true, true, true]);
		expect([200, 429, 500, 502, 503].map(isTerminalStatus)).toEqual([
			false,
			false,
			false,
			false,
			false
		]);
	});
});

describe('createManagedStream', () => {
	test('opens once up front and forwards events to the caller', () => {
		const { opened, time } = harness();
		expect(opened).toHaveLength(1);

		opened[0].emit('time', { left: '1h' });

		expect(time).toHaveBeenCalledOnce();
	});

	test('closes on upstream-error so the polyfill cannot race its own retry', () => {
		const { opened } = harness();

		fail5xx(opened[0]);

		expect(opened[0].closed).toBe(true);
		expect(opened, 'reconnect is scheduled, not immediate').toHaveLength(1);
	});

	test('reconnects on the backoff schedule while failures continue', () => {
		const { opened } = harness();

		fail5xx(opened[0]);
		vi.advanceTimersByTime(999);
		expect(opened).toHaveLength(1);
		vi.advanceTimersByTime(1);
		expect(opened).toHaveLength(2);

		fail5xx(opened[1]);
		vi.advanceTimersByTime(2_000);
		expect(opened).toHaveLength(3);

		fail5xx(opened[2]);
		vi.advanceTimersByTime(3_999);
		expect(opened, 'third failure waits four seconds').toHaveLength(3);
		vi.advanceTimersByTime(1);
		expect(opened).toHaveLength(4);
	});

	test('reports degradation once the streak is reached, and only once', () => {
		const { opened, onDegraded } = harness();

		for (let i = 0; i < 4; i++) {
			fail5xx(opened[i]);
			vi.advanceTimersByTime(retryDelay(i + 1));
		}

		expect(onDegraded).toHaveBeenCalledOnce();
		expect(onDegraded).toHaveBeenCalledWith({
			stream: 'scoreboard',
			status: 500,
			terminal: false
		});
	});

	test('a connection that holds up clears the streak and reports recovery', () => {
		// 恢復的訊號是 settle window,不是「收到一則資料」—— 多路連線在被判定失敗
		// 之前本來就會先送幾則正常事件出來。
		const { opened, onDegraded, onRecovered, latest } = harness();

		for (let i = 0; i < 3; i++) {
			fail5xx(opened[i]);
			vi.advanceTimersByTime(retryDelay(i + 1));
		}
		expect(onDegraded).toHaveBeenCalledOnce();

		latest().opened();
		latest().emit('time', { left: '1h' });
		expect(onRecovered, '資料到了但窗口還沒過,還不算恢復').not.toHaveBeenCalled();
		vi.advanceTimersByTime(5_000);
		expect(onRecovered).toHaveBeenCalledOnce();

		// The streak restarted, so the next failure waits one second again.
		fail5xx(latest());
		const before = opened.length;
		vi.advanceTimersByTime(1_000);
		expect(opened).toHaveLength(before + 1);
	});

	test('teardown closes the connection and cancels a pending retry', () => {
		const { opened, stop } = harness();
		fail5xx(opened[0]);

		stop();
		vi.advanceTimersByTime(60_000);

		expect(opened, 'no reconnect after teardown').toHaveLength(1);
	});

	test('teardown closes a live connection', () => {
		const { opened, stop } = harness();

		stop();

		expect(opened[0].closed).toBe(true);
	});

	test('a malformed report still counts as a failure', () => {
		const { opened, onDegraded } = harness({ degradedAfter: 1 });

		opened[0].listeners.get('upstream-error')?.[0]({ data: 'not json' });

		expect(onDegraded).toHaveBeenCalledWith({ stream: 'unknown', status: 0, terminal: false });
	});
});

describe('terminal failures are not retried', () => {
	test('a 404 upstream stops at once and says so without waiting for a streak', () => {
		const { opened, onDegraded } = harness();

		opened[0].emit('upstream-error', { stream: 'time', status: 404 });

		expect(onDegraded, 'no point waiting for a streak that can never grow').toHaveBeenCalledWith({
			stream: 'time',
			status: 404,
			terminal: true
		});
		vi.advanceTimersByTime(120_000);
		expect(opened, 'a finished contest must not be retried forever').toHaveLength(1);
	});

	test('a connect-time 404 is terminal too', () => {
		const { opened, onDegraded } = harness();

		opened[0].errored(404);

		expect(onDegraded).toHaveBeenCalledWith({ stream: 'connection', status: 404, terminal: true });
		vi.advanceTimersByTime(120_000);
		expect(opened).toHaveLength(1);
	});
});

describe('connect-time failures', () => {
	test('a 500 is supervised: backoff, then a notice on the third attempt', () => {
		const { opened, onDegraded } = harness();

		opened[0].errored(500);
		expect(onDegraded, 'one failure is not yet worth a notice').not.toHaveBeenCalled();
		vi.advanceTimersByTime(1_000);
		expect(opened).toHaveLength(2);

		opened[1].errored(502);
		vi.advanceTimersByTime(2_000);
		expect(opened).toHaveLength(3);

		opened[2].errored(503);
		expect(onDegraded).toHaveBeenCalledWith({
			stream: 'connection',
			status: 503,
			terminal: false
		});
	});

	test('a wrong content-type arrives as status 200 and is still supervised', () => {
		// The polyfill rejects a 200 whose Content-Type is not text/event-stream,
		// so classifying on "is the status 200" would miss it entirely.
		const { opened } = harness();

		opened[0].errored(200);

		vi.advanceTimersByTime(1_000);
		expect(opened).toHaveLength(2);
	});

	test('a mid-stream drop carries no status and is left to the polyfill', () => {
		const { opened, onDegraded } = harness();

		opened[0].errored();

		vi.advanceTimersByTime(120_000);
		expect(opened, 'the polyfill is already reopening this one').toHaveLength(1);
		expect(opened[0].closed, 'and it must not be closed out from under it').toBe(false);
		expect(onDegraded).not.toHaveBeenCalled();
	});

	test('the caller may claim an error and end supervision', () => {
		const onError = vi.fn((_event: Event, giveUp: () => void) => giveUp());
		const { opened, onDegraded } = harness({ onError });

		opened[0].errored(401);

		expect(onError).toHaveBeenCalledOnce();
		expect(opened[0].closed, 'the caller stopped supervision').toBe(true);
		expect(onDegraded, 'the caller is showing its own message').not.toHaveBeenCalled();
		vi.advanceTimersByTime(120_000);
		expect(opened).toHaveLength(1);
	});
});

describe('one connection fails at most once', () => {
	test('two error frames on the same connection schedule a single retry', () => {
		const { opened } = harness();

		opened[0].emit('upstream-error', { stream: 'scoreboard', status: 500 });
		opened[0].emit('upstream-error', { stream: 'clarification', status: 502 });

		// Two failures would have counted two strikes and armed two timers, so the
		// second reconnect would fire at 1s and a third at 2s.
		vi.advanceTimersByTime(1_000);
		expect(opened, 'exactly one reconnect').toHaveLength(2);
		vi.advanceTimersByTime(10_000);
		expect(opened, 'and no stray timer behind it').toHaveLength(2);
	});

	test('a late error from a replaced connection does not disturb its successor', () => {
		const { opened } = harness();

		fail5xx(opened[0]);
		vi.advanceTimersByTime(1_000);
		expect(opened).toHaveLength(2);

		opened[0].emit('upstream-error', { stream: 'scoreboard', status: 500 });

		expect(opened[1].closed, 'the successor is untouched').toBe(false);
		vi.advanceTimersByTime(60_000);
		expect(opened).toHaveLength(2);
	});
});

describe('settle window', () => {
	test('a connection that stays quiet long enough counts as recovered', () => {
		const { opened, onDegraded, onRecovered, latest } = harness();

		for (let i = 0; i < 3; i++) {
			fail5xx(opened[i]);
			vi.advanceTimersByTime(retryDelay(i + 1));
		}
		expect(onDegraded).toHaveBeenCalledOnce();

		// This is the only recovery signal a silent stream can produce: it opens
		// and then delivers nothing at all until something changes upstream.
		latest().opened();
		vi.advanceTimersByTime(4_999);
		expect(onRecovered, 'not yet — the window has not elapsed').not.toHaveBeenCalled();
		vi.advanceTimersByTime(1);
		expect(onRecovered).toHaveBeenCalledOnce();
	});

	test('a failure inside the window is not a recovery', () => {
		const { opened, onRecovered } = harness({ degradedAfter: 1 });

		fail5xx(opened[0]);
		vi.advanceTimersByTime(1_000);

		opened[1].opened();
		vi.advanceTimersByTime(4_000);
		fail5xx(opened[1]);
		vi.advanceTimersByTime(60_000);

		expect(onRecovered).not.toHaveBeenCalled();
	});

	test('teardown cancels the settle timer', () => {
		const { opened, onRecovered, stop } = harness({ degradedAfter: 1 });

		fail5xx(opened[0]);
		vi.advanceTimersByTime(1_000);
		opened[1].opened();
		stop();
		vi.advanceTimersByTime(60_000);

		expect(onRecovered, 'nothing may fire against a torn-down stream').not.toHaveBeenCalled();
	});
});

describe('resync', () => {
	test('the first open needs none — the page fetched its own data', () => {
		const onResync = vi.fn();
		const { opened } = harness({ onResync });

		opened[0].opened();

		expect(onResync).not.toHaveBeenCalled();
	});

	test('every reconnect pulls the data back', () => {
		// The backend pushes scoreboards and worker lists only on change, so a
		// reconnected stream replays nothing and the page keeps showing whatever
		// it had. This is the only thing that fixes that.
		const onResync = vi.fn();
		const { opened, latest } = harness({ onResync });
		opened[0].opened();

		fail5xx(opened[0]);
		vi.advanceTimersByTime(1_000);
		latest().opened();

		expect(onResync).toHaveBeenCalledOnce();
	});

	test('a polyfill-driven reconnect counts too', () => {
		// A mid-stream drop is reopened by the polyfill without telling us — but
		// it still dispatches `open` on the same EventSource, so we still resync.
		const onResync = vi.fn();
		const { opened } = harness({ onResync });

		opened[0].opened();
		opened[0].opened();

		expect(onResync).toHaveBeenCalledOnce();
	});

	test('while degraded, it falls back to REST on an interval', () => {
		const onResync = vi.fn();
		const { opened } = harness({ onResync, degradedAfter: 1, resyncEveryMs: 10_000 });

		fail5xx(opened[0]);
		expect(onResync, 'not immediately — the retry may well succeed').not.toHaveBeenCalled();

		vi.advanceTimersByTime(10_000);
		vi.advanceTimersByTime(10_000);
		// Two interval ticks, plus whatever reconnects happened in that window.
		expect(onResync.mock.calls.length).toBeGreaterThanOrEqual(2);
	});

	test('a terminal failure does not start polling', () => {
		// Nothing is coming back — knocking on the backend every 30s for a contest
		// that ended is exactly what the terminal classification exists to avoid.
		const onResync = vi.fn();
		const { opened } = harness({ onResync });

		opened[0].emit('upstream-error', { stream: 'time', status: 404 });
		vi.advanceTimersByTime(300_000);

		expect(onResync).not.toHaveBeenCalled();
	});

	test('recovery stops the polling', () => {
		const onResync = vi.fn();
		const { opened, latest } = harness({ onResync, degradedAfter: 1, resyncEveryMs: 10_000 });

		fail5xx(opened[0]);
		vi.advanceTimersByTime(10_000);
		const whileDegraded = onResync.mock.calls.length;
		expect(whileDegraded).toBeGreaterThan(0);

		latest().opened();
		vi.advanceTimersByTime(5_000);
		onResync.mockClear();
		vi.advanceTimersByTime(120_000);

		expect(onResync, 'the stream is delivering again; stop polling').not.toHaveBeenCalled();
	});

	test('teardown stops the polling', () => {
		const onResync = vi.fn();
		const { opened, stop } = harness({ onResync, degradedAfter: 1, resyncEveryMs: 10_000 });

		fail5xx(opened[0]);
		stop();
		onResync.mockClear();
		vi.advanceTimersByTime(120_000);

		expect(onResync).not.toHaveBeenCalled();
	});
});

describe('a connection that delivers data and then fails', () => {
	test('still counts as a failure', () => {
		// 這是多路 route 的真實形狀:time 每秒推送,scoreboard 壞掉時錯誤訊框排在
		// 那些 time 之後。如果一則資料就把 streak 歸零,連線就會永遠以一秒為間隔
		// 重連、而畫面上永遠不會出現告示 —— 正是這個功能要消滅的那個病徵。
		const { opened, onDegraded } = harness();

		for (let i = 0; i < 4; i++) {
			opened[i].emit('time', { left: '1h' });
			fail5xx(opened[i]);
			vi.advanceTimersByTime(retryDelay(i + 1));
		}

		expect(onDegraded, '資料先到不代表這條連線是健康的').toHaveBeenCalledOnce();
	});
});
