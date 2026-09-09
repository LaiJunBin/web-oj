/**
 * The sidebar's running summary score after one submission's result arrives on
 * the live stream. A partial submission never counts toward the summary — only
 * a full submission can raise it. This is the same rule fetchProject applies
 * when it computes the summary from the full list (isPartial ? 0 : score); it
 * lives here so the SSE path and the REST path agree and the rule has one
 * testable home.
 */
export function nextSummaryScore(
	current: number,
	submissionScore: number,
	isPartial: boolean
): number {
	return isPartial ? current : Math.max(current, submissionScore);
}
