import '@testing-library/jest-dom/vitest';
import { render, screen } from '@testing-library/svelte';
import { describe, expect, test } from 'vitest';
import LifecycleResult from './LifecycleResult.svelte';
import type { LifecycleRun } from '$lib/judgmentLifecycle';

const run = (exitCode: number, stderr = '', stdout = ''): LifecycleRun => ({
	id: 1,
	executionEnvironment: 'project',
	command: 'mysql -u root -e "select 1"',
	result: { exitCode, execTime: 12, stdout, stderr }
});

// 一次真實的 before_all 失敗(mysql Access denied)只能靠翻資料才找到 stderr。
// 管理員要在詳情頁一眼看到:哪一步、exit 多少、錯什麼。
describe('LifecycleResult', () => {
	test('exit code 非零:醒目的失敗標記,stderr 預設展開', () => {
		render(LifecycleResult, {
			run: run(1, 'ERROR 1045 (28000): Access denied for user'),
			title: 'Before All #1'
		});

		expect(screen.getByText('失敗 · exit 1')).toBeInTheDocument();
		const stderr = screen.getByText(/^stderr/).closest('details');
		expect(stderr).toHaveAttribute('open');
		expect(screen.getByText(/Access denied/)).toBeInTheDocument();
	});

	test('exit 0:沒有失敗標記,輸出收合、標出行數', () => {
		render(LifecycleResult, { run: run(0, '', 'a\nb\nc'), title: 'After All #1' });

		expect(screen.queryByText(/失敗/)).not.toBeInTheDocument();
		expect(screen.getByText('exit 0')).toBeInTheDocument();
		expect(screen.getByText(/^stderr/).closest('details')).not.toHaveAttribute('open');
		expect(screen.getByText('stdout (3 行)')).toBeInTheDocument();
		expect(screen.getByText('stderr(空)')).toBeInTheDocument();
	});

	test('失敗但 stderr 是空的:不展開一個空框', () => {
		render(LifecycleResult, { run: run(2), title: 'Before Each #1' });

		expect(screen.getByText('失敗 · exit 2')).toBeInTheDocument();
		expect(screen.getByText(/^stderr/).closest('details')).not.toHaveAttribute('open');
	});
});
