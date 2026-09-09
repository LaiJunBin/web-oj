import { HttpResponse } from 'msw';
import type { apiClient } from '@/api';
import type { StrictResponseOf } from '../handlers.overwritten';

export function login(): StrictResponseOf<typeof apiClient.login> {
	return HttpResponse.json({
		success: true,
		data: {
			accessToken:
				'eyJpZCI6MSwibmFtZSI6IueuoeeQhuWToSIsInVzZXJuYW1lIjoiYWRtaW4iLCJyb2xlcyI6WyJBZG1pbiIsIlRlYW0gTWVtYmVyIl19',
			refreshToken: 'test'
		}
	});
}
