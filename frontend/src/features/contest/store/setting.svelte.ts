import { browser } from '$app/environment';
import type { Setting, SettingStorageInterface } from '@/apps/contest/domain/setting';
import { BrowserStorage } from '@features/common/store/base.svelte';

const NOTIFICATION_ENABLED_KEY = 'notificationEnabled';

export class SettingStorage extends BrowserStorage<Setting> implements SettingStorageInterface {
	// `enableNotification` is the user's in-app preference (persisted). The browser
	// permission stays the source of truth for whether a notification can actually
	// be sent — see the `canSend` gate in the layout / Navbar. Default on first
	// visit; a stored 'false' (user muted) persists across reloads.
	data = $state<Setting>({
		enableNotification: browser ? localStorage.getItem(NOTIFICATION_ENABLED_KEY) !== 'false' : false
	});

	set(data: Setting): void {
		super.set(data);
		if (browser) {
			localStorage.setItem(NOTIFICATION_ENABLED_KEY, String(data.enableNotification));
		}
	}
}
