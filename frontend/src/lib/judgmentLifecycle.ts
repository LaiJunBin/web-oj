import type { Judgment } from '@/apps/admin/domain/submission';

/**
 * Finding the lifecycle step that broke a judgment.
 *
 * A judgment runs lifecycle commands around the tests: before/after all at
 * the judgment level (in the project and in the test environment), and
 * before/after each at the testcase level. When one of them fails — a real
 * case: `mysql` refused the connection in before_all — the judgment ends in
 * Error and the only trace is that step's exit code and stderr. The admin API
 * already returns them; these helpers pull the failed ones out so the detail
 * page can show "which step, what error" at the top instead of burying it
 * two accordions down.
 */

export type LifecycleRun = Judgment['judgmentLifecycle']['beforeAll'][number];
export type JudgmentPhase = 'beforeAll' | 'afterAll';
export type TestcasePhase = 'beforeEach' | 'afterEach';

export type StepLocation =
	| { level: 'judgment'; environment: string; phase: JudgmentPhase; index: number }
	| {
			level: 'testcase';
			gradingItemId: number;
			testcaseId: number;
			phase: TestcasePhase;
			index: number;
	  };

export type FailedStep = { at: StepLocation; run: LifecycleRun };

export const PHASE_LABEL: Record<JudgmentPhase | TestcasePhase, string> = {
	beforeAll: 'Before All',
	afterAll: 'After All',
	beforeEach: 'Before Each',
	afterEach: 'After Each'
};

const ENVIRONMENT_LABEL: Record<string, string> = { project: '專案環境', test: '測試環境' };

export const environmentLabel = (environment: string) =>
	ENVIRONMENT_LABEL[environment] ?? environment;

/** A non-zero exit code is the only failure signal a lifecycle step has. */
export const hasFailed = (run: LifecycleRun) => run.result.exitCode !== 0;

/** Every failed step, judgment level first, in execution order. */
export function collectFailedSteps(
	judgment: Pick<Judgment, 'judgmentLifecycle' | 'gradingItems'>
): FailedStep[] {
	const failed: FailedStep[] = [];
	for (const phase of ['beforeAll', 'afterAll'] as const) {
		judgment.judgmentLifecycle[phase].forEach((run, i) => {
			if (hasFailed(run)) {
				failed.push({
					at: { level: 'judgment', environment: run.executionEnvironment, phase, index: i + 1 },
					run
				});
			}
		});
	}
	for (const gradingItem of judgment.gradingItems) {
		for (const testcase of gradingItem.testcases) {
			for (const phase of ['beforeEach', 'afterEach'] as const) {
				testcase.lifecycle[phase].forEach((run, i) => {
					if (hasFailed(run)) {
						failed.push({
							at: {
								level: 'testcase',
								gradingItemId: gradingItem.id,
								testcaseId: testcase.id,
								phase,
								index: i + 1
							},
							run
						});
					}
				});
			}
		}
	}
	return failed;
}

/** Human-readable place of a step, e.g. 「專案環境 Before All #1」 or 「題目 / 測資 Before Each #2」. */
export function describeStep(
	at: StepLocation,
	names: { gradingItem: (id: number) => string; testcase: (id: number) => string }
): string {
	if (at.level === 'judgment') {
		return `${environmentLabel(at.environment)} ${PHASE_LABEL[at.phase]} #${at.index}`;
	}
	return `${names.gradingItem(at.gradingItemId)} / ${names.testcase(at.testcaseId)} ${PHASE_LABEL[at.phase]} #${at.index}`;
}

/** The first non-empty line of an output, for a one-line summary. */
export function firstLine(text: string): string {
	return (
		text
			.split('\n')
			.find((line) => line.trim() !== '')
			?.trim() ?? ''
	);
}

/** Line count for a collapsed output's summary; 0 for whitespace-only. */
export function lineCount(text: string): number {
	return text.trim() === '' ? 0 : text.trimEnd().split('\n').length;
}
