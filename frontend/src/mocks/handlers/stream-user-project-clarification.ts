import { HttpResponse } from 'msw';

function createEventSourceResponse<T>(data: T) {
	return `data:${data}\n\n`;
}

export function streamUserProjectClarification() {
	const encoder = new TextEncoder();
	const stream = new ReadableStream({
		start(controller) {
			const interval = setInterval(() => {
				const data = {
					subject: 'Clarification: ' + Math.random().toString(36).substring(7),
					content: 'Content: ' + Math.random().toString(36).substring(7)
				};
				controller.enqueue(encoder.encode(createEventSourceResponse(JSON.stringify(data))));
			}, 5000);
		}
	});

	return new HttpResponse(stream, {
		headers: {
			'Content-Type': 'text/event-stream'
		}
	});
}
