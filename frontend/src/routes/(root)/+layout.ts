import type { LayoutLoad } from './$types';
import { User, useUserStorage } from '@/apps/contest/domain/user';
import { PROFILE_DEPENDENCY } from '$lib/profileRefresh';

// 這是 universal load,SSR 時也會跑,而那時的 fetch 是 SvelteKit 的:它只在
// 「目標 host 等於本站 host 或其子網域」時才轉發瀏覽器的 cookie(埠不影響)。
// 所以後端的 hostname 必須與前端相同或為其子網域,否則這一趟拿不到 cookie、
// 後端回未認證,伺服器就把已登入的使用者渲染成登出狀態——後台網址直接輸入會被
// 守衛導回登入頁。要支援跨 host 的佈署,得把這段搬到 +layout.server.ts 並自行
// 轉發 cookie(SSE 代理已經是那樣做的,見 lib/server/backend.ts)。
export const load: LayoutLoad = async ({ fetch, depends }) => {
	// Roles are loaded here and nowhere else, and client-side navigation reuses
	// this result. An admin who changes their own roles asks for a re-run with
	// invalidate(PROFILE_DEPENDENCY) — see lib/profileRefresh.
	depends(PROFILE_DEPENDENCY);
	const userStorage = useUserStorage();
	try {
		const user: User = await fetch(`${import.meta.env.VITE_API_URL}/auth/profile`, {
			credentials: 'include'
		})
			.then((res) => res.json())
			.then(({ success, data }) => {
				if (!success) {
					throw new Error('Not authenticated');
				}

				return new User(data.name, data.username, data.roles);
			});

		userStorage.set(user);
		return { user };
	} catch {
		userStorage.clear();
	}
};
