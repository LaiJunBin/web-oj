import { describe, expect, test } from 'vitest';
import {
	collectFailedSteps,
	describeStep,
	firstLine,
	lineCount,
	type LifecycleRun
} from '$lib/judgmentLifecycle';

let seq = 0;
const run = (exitCode: number, executionEnvironment = 'project'): LifecycleRun => ({
	id: ++seq,
	executionEnvironment,
	command: `step ${seq}`,
	result: { exitCode, execTime: 1, stdout: '', stderr: exitCode ? 'boom\nmore' : '' }
});

const judgment = {
	judgmentLifecycle: {
		beforeAll: [run(0), run(1, 'test'), run(0)],
		afterAll: [run(0)]
	},
	gradingItems: [
		{
			id: 10,
			testcases: [
				{ id: 100, lifecycle: { beforeEach: [run(0)], afterEach: [run(3)] } },
				{ id: 101, lifecycle: { beforeEach: [], afterEach: [] } }
			]
		}
	]
} as unknown as Parameters<typeof collectFailedSteps>[0];

const names = { gradingItem: (id: number) => `GI${id}`, testcase: (id: number) => `TC${id}` };

describe('collectFailedSteps', () => {
	test('只挑 exit code 非零的步驟,判定層在前、測資層在後,並帶上位置', () => {
		const failed = collectFailedSteps(judgment);

		expect(failed.map((f) => f.run.result.exitCode)).toEqual([1, 3]);
		expect(failed[0].at).toEqual({
			level: 'judgment',
			environment: 'test',
			phase: 'beforeAll',
			index: 2
		});
		expect(failed[1].at).toEqual({
			level: 'testcase',
			gradingItemId: 10,
			testcaseId: 100,
			phase: 'afterEach',
			index: 1
		});
	});

	test('全部成功:空清單(頁面就不掛失敗摘要)', () => {
		expect(
			collectFailedSteps({
				judgmentLifecycle: { beforeAll: [run(0)], afterAll: [] },
				gradingItems: []
			} as unknown as Parameters<typeof collectFailedSteps>[0])
		).toEqual([]);
	});
});

describe('describeStep', () => {
	test('判定層:環境 + 階段 + 序號', () => {
		expect(
			describeStep({ level: 'judgment', environment: 'test', phase: 'beforeAll', index: 2 }, names)
		).toBe('測試環境 Before All #2');
	});

	test('測資層:評分項 / 測資 + 階段 + 序號', () => {
		expect(
			describeStep(
				{ level: 'testcase', gradingItemId: 10, testcaseId: 100, phase: 'afterEach', index: 1 },
				names
			)
		).toBe('GI10 / TC100 After Each #1');
	});

	test('未知的執行環境:照原字串顯示,不吞掉', () => {
		expect(
			describeStep(
				{ level: 'judgment', environment: 'sandbox', phase: 'afterAll', index: 1 },
				names
			)
		).toBe('sandbox After All #1');
	});
});

describe('firstLine / lineCount', () => {
	test('firstLine 跳過空行、去掉前後空白', () => {
		expect(firstLine('\n  \n  ERROR 1045: Access denied  \nmore')).toBe(
			'ERROR 1045: Access denied'
		);
		expect(firstLine('   ')).toBe('');
	});

	test('lineCount 只算真的有內容的輸出', () => {
		expect(lineCount('')).toBe(0);
		expect(lineCount(' \n ')).toBe(0);
		expect(lineCount('a\nb\n')).toBe(2);
	});
});
