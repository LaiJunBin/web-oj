import { HttpResponse } from 'msw';

function createEventSourceResponse<T>(data: T) {
	return `data:${data}\n\n`;
}

function dhhmmss(time: number) {
	const days = Math.floor(time / (3600 * 24));
	const hours = Math.floor((time % (3600 * 24)) / 3600);
	const minutes = Math.floor((time % 3600) / 60);
	const seconds = time % 60;
	if (days > 0) {
		return `${days}d${hours}h${minutes}m${seconds}s`;
	}

	if (hours > 0) {
		return `${hours}h${minutes}m${seconds}s`;
	}

	if (minutes > 0) {
		return `${minutes}m${seconds}s`;
	}

	return `${seconds}s`;
}

export function streamUserProjectTime() {
	const data = {
		time_type: 'submit',
		time_left: 100
	};
	const encoder = new TextEncoder();
	const stream = new ReadableStream({
		start(controller) {
			const interval = setInterval(() => {
				data.time_left -= 1;
				controller.enqueue(
					encoder.encode(
						createEventSourceResponse(
							JSON.stringify({
								...data,
								time_left: dhhmmss(data.time_left)
							})
						)
					)
				);
				if (data.time_left === 0) {
					controller.close();
					clearInterval(interval);
				}
			}, 1000);

			controller.enqueue(
				encoder.encode(
					createEventSourceResponse(
						JSON.stringify({
							...data,
							time_left: dhhmmss(data.time_left)
						})
					)
				)
			);
		}
	});

	return new HttpResponse(stream, {
		headers: {
			'Content-Type': 'text/event-stream'
		}
	});
}
