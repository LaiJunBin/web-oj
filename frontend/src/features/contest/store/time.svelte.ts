import { BrowserStorage } from '@features/common/store/base.svelte';
import type { Time, TimeStorageInterface } from '@/apps/contest/domain/time';

export class TimeStorage extends BrowserStorage<Time> implements TimeStorageInterface {}
