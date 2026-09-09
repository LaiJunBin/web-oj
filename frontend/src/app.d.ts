/// <reference types="@sveltejs/adapter-node" />

// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
declare global {
	namespace App {
		// interface Error {}
		interface Locals {
			authService: AuthService;
		}
		// interface PageData {}
		interface PageState {
			id: number;
		}
	}
}

declare module 'event-source-polyfill' {
	interface EventSourceEventMap {
		time: MessageEvent;
		scoreboard: MessageEvent;
		clarification: MessageEvent;
		log: MessageEvent;
		/** Emitted by the SSE proxy when an upstream it multiplexes answers non-200. */
		'upstream-error': MessageEvent;
	}
}

export {};
