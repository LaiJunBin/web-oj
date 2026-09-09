/**
 * Which contest (or project) to load on bootstrap, and what a failed load
 * means for the browser's memory of it.
 *
 * The root layout remembers the last selection in sessionStorage. That memory
 * outlives the data: a contest is deleted mid-event, a development database is
 * reset, a contestant is taken out of a contest. Loading the remembered id
 * then fails for good, and a "try again later" screen would be a lie — nothing
 * will bring it back. So the memory is checked against the freshly fetched
 * list first, and a definite "gone" from the server is treated as final.
 */

export type SelectionKey = 'currentContestId' | 'currentProjectId';

export type Choice = {
	/** The id to load; undefined when there is nothing to choose from. */
	id: number | undefined;
	/** The id came from memory and is still on the list. */
	remembered: boolean;
	/** Memory pointed at something that is no longer on the list. */
	forgotten: boolean;
};

export function rememberedId(key: SelectionKey): number | null {
	const raw = sessionStorage.getItem(key);
	const id = raw === null ? NaN : parseInt(raw, 10);
	return Number.isInteger(id) && id > 0 ? id : null;
}

/** Drops the memory. Forgetting a contest also forgets its project. */
export function forget(key: SelectionKey): void {
	sessionStorage.removeItem(key);
	if (key === 'currentContestId') {
		sessionStorage.removeItem('currentProjectId');
	}
}

/**
 * Picks what to load. A remembered id still on the list wins; otherwise the
 * first entry. The list is fresh and the memory is not, so a remembered id
 * that is missing from it is dropped rather than fetched — that request could
 * only fail.
 */
export function choose(remembered: number | null, available: number[]): Choice {
	if (remembered !== null && available.includes(remembered)) {
		return { id: remembered, remembered: true, forgotten: false };
	}
	return { id: available[0], remembered: false, forgotten: remembered !== null };
}

/**
 * Whether a failed load means the thing is gone for good. Only 404 is final.
 * A 5xx or a network failure is a bad moment, not a missing contest — the
 * backend reports infrastructure trouble as 503 — and a retry can succeed, so
 * the memory is kept and the error screen's "try again later" stays true.
 */
export function isGone(error: unknown): boolean {
	return (error as { response?: { status?: number } } | null)?.response?.status === 404;
}

/**
 * After loading a remembered selection: 'forget' when the server says it is
 * gone (drop the memory, re-select from a fresh list); 'keep' otherwise —
 * loaded fine, or failed for a reason worth retrying. A failure on a default,
 * not remembered, selection is also 'keep': there is no memory to blame, and
 * forgetting would change nothing.
 */
export function afterLoad(error: unknown, remembered: boolean): 'keep' | 'forget' {
	return remembered && isGone(error) ? 'forget' : 'keep';
}
