import { HttpResponse } from 'msw';

function createEventSourceResponse<T>(data: T) {
	return `data:${data}\n\n`;
}

const statuses = ['not_submitted', 'pending', 'running', 'passed', 'failed'];
export function streamUserProjectScoreboard() {
	const encoder = new TextEncoder();
	const stream = new ReadableStream({
		start(controller) {
			const interval = setInterval(() => {
				const gradingItemLength = Math.floor(Math.random() * 10);
				const data = {
					heads: Array(gradingItemLength)
						.fill(0)
						.map(() => Math.random().toString(36).substring(7)),
					rows: Array(Math.floor(Math.random() * 10))
						.fill(0)
						.map(() => ({
							name: Math.random().toString(36).substring(7),
							username: Math.random().toString(36).substring(7),
							statuses: Array(gradingItemLength)
								.fill(0)
								.map(() => statuses[Math.floor(Math.random() * statuses.length)])
						}))
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
