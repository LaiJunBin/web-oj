import { HttpResponse } from 'msw';
import type { apiClient } from '@/api';
import type { StrictResponseOf } from '../handlers.overwritten';

export function getProfile(): StrictResponseOf<typeof apiClient.getProfile> {
	return HttpResponse.json({
		success: true,
		data: {
			name: 'Test',
			username: 'test',
			roles: ['admin', 'contestant']
		}
	});
}
