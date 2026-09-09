import { container } from '@/apps/contest/container';
import type { Storage } from '@apps/common/domain/storage';

export class Setting {
	public enableNotification: boolean = false;
}

export interface SettingStorageInterface extends Storage<Setting> {}
export const useSettingStorage = () => container.get('SettingStorage') as SettingStorageInterface;
