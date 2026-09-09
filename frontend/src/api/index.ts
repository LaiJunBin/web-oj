import { mergeApis, Zodios } from '@zodios/core';
import { browser } from '$app/environment';
import * as API from './gen';
import axios from 'axios';

const APIs = Object.entries(API)
	.map(([_, api]) => api)
	.flat();

const axiosInstance = axios.create({
	baseURL: import.meta.env.VITE_API_URL,
	withCredentials: true
});

axiosInstance.interceptors.request.use(async (config) => {
	if (config.method === 'options' || config.method === 'get') {
		return config;
	}

	const csrf = await apiClient.getCsrfToken();
	config.headers['csrf-token'] = csrf.data.token;
	return config;
});

// The backend auto-renews the access token via the refresh_token cookie inside
// its authorization middleware, so a 401 that actually reaches the client means
// the session is truly dead (the refresh also failed). Surface it as a single
// window event instead of letting every call fail silently; the app layer (root
// layout) decides how to react — kept decoupled here to avoid an import cycle
// with the DI container. Client-only: `goto`/UI belong to the app, not SSR.
axiosInstance.interceptors.response.use(
	(response) => response,
	(error) => {
		if (browser && error?.response?.status === 401) {
			window.dispatchEvent(new Event('auth:unauthorized'));
		}
		return Promise.reject(error);
	}
);

export const apiClient = new Zodios(
	import.meta.env.VITE_API_URL,
	APIs as unknown as ReturnType<typeof mergeApis<typeof API>>,
	{
		axiosInstance
	}
);
