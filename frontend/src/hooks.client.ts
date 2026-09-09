import { dev } from '$app/environment';

async function enableMocking() {
	if (!dev) {
		return;
	}

	const { worker } = await import('./mocks/browser.ts');

	const whitelist = [
		import.meta.env.VITE_API_URL,
		'chrome-extension://',
		location.origin
		// 'http://localhost:8080',
	];

	return worker.start({
		onUnhandledRequest(req) {
			if (!whitelist.some((url) => req.url.startsWith(url))) {
				console.error(`No handler for ${req.url}`);
			}
		}
	});
}

// await enableMocking();
